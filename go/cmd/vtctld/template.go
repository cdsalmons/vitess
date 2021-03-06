package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path"
	"reflect"
	"strings"

	"golang.org/x/net/context"

	"github.com/youtube/vitess/go/vt/topo"
	"github.com/youtube/vitess/go/vt/topotools"

	pb "github.com/youtube/vitess/go/vt/proto/topodata"
)

// FHtmlize writes data to w as debug HTML (using definition lists).
func FHtmlize(w io.Writer, data interface{}) {
	v := reflect.Indirect(reflect.ValueOf(data))
	typ := v.Type()
	switch typ.Kind() {
	case reflect.Struct:
		fmt.Fprintf(w, "<dl class=\"%s\">", typ.Name())
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			if field.PkgPath != "" {
				continue
			}
			fmt.Fprintf(w, "<dt>%v</dt>", field.Name)
			fmt.Fprint(w, "<dd>")
			FHtmlize(w, v.Field(i).Interface())
			fmt.Fprint(w, "</dd>")
		}
		fmt.Fprintf(w, "</dl>")
	case reflect.Slice:
		fmt.Fprint(w, "<ul>")
		for i := 0; i < v.Len(); i++ {
			fmt.Fprint(w, "<li>")
			FHtmlize(w, v.Index(i).Interface())
			fmt.Fprint(w, "</li>")
		}
		fmt.Fprint(w, "</ul>")
	case reflect.Map:
		fmt.Fprintf(w, "<dl class=\"map\">")
		for _, k := range v.MapKeys() {
			fmt.Fprint(w, "<dt>")
			FHtmlize(w, k.Interface())
			fmt.Fprint(w, "</dt>")
			fmt.Fprint(w, "<dd>")
			FHtmlize(w, v.MapIndex(k).Interface())
			fmt.Fprint(w, "</dd>")
		}
		fmt.Fprintf(w, "</dl>")
	default:
		printed := fmt.Sprintf("%v", v.Interface())
		if printed == "" {
			printed = "&nbsp;"
		}
		fmt.Fprint(w, printed)
	}
}

// Htmlize returns a debug HTML representation of data.
func Htmlize(data interface{}) string {
	b := new(bytes.Buffer)
	FHtmlize(b, data)
	return b.String()
}

func link(text, href string) string {
	return fmt.Sprintf("<a href=%q>%v</a>", href, text)
}

func breadCrumbs(fullPath string) template.HTML {
	parts := strings.Split(fullPath, "/")
	paths := make([]string, len(parts))
	for i, part := range parts {
		if i == 0 {
			paths[i] = "/"
			continue
		}
		paths[i] = path.Join(paths[i-1], part)
	}
	b := new(bytes.Buffer)
	for i, part := range parts[1 : len(parts)-1] {
		fmt.Fprint(b, "/"+link(part, paths[i+1]))
	}
	fmt.Fprintf(b, "/"+parts[len(parts)-1])
	return template.HTML(b.String())
}

// FuncMap defines functions accessible in templates. It can be modified in
// init() method by plugins to provide extra formatting.
var FuncMap = template.FuncMap{
	"htmlize": func(o interface{}) template.HTML {
		return template.HTML(Htmlize(o))
	},
	"hasprefix": strings.HasPrefix,
	"intequal": func(left, right int) bool {
		return left == right
	},
	"breadcrumbs": breadCrumbs,
	"keyspace": func(keyspace string) template.HTML {
		if explorer == nil {
			return template.HTML(keyspace)
		}
		return template.HTML(link(keyspace, explorer.GetKeyspacePath(keyspace)))
	},
	"srv_keyspace": func(cell, keyspace string) template.HTML {
		if explorer == nil {
			return template.HTML(keyspace)
		}
		return template.HTML(link(keyspace, explorer.GetSrvKeyspacePath(cell, keyspace)))
	},
	"shard": func(keyspace string, shard *topotools.ShardNodes) template.HTML {
		if explorer == nil {
			return template.HTML(shard.Name)
		}
		return template.HTML(link(shard.Name, explorer.GetShardPath(keyspace, shard.Name)))
	},
	"srv_shard": func(cell, keyspace string, shard *topotools.ShardNodes) template.HTML {
		if explorer == nil {
			return template.HTML(shard.Name)
		}
		return template.HTML(link(shard.Name, explorer.GetSrvShardPath(cell, keyspace, shard.Name)))
	},
	"tablet": func(alias *pb.TabletAlias, shortname string) template.HTML {
		if explorer == nil {
			return template.HTML(shortname)
		}
		return template.HTML(link(shortname, explorer.GetTabletPath(alias)))
	},
}

// TemplateLoader is a helper class to load html templates.
type TemplateLoader struct {
	Directory string
	usesDummy bool
	template  *template.Template
}

func (loader *TemplateLoader) compile() (*template.Template, error) {
	return template.New("main").Funcs(FuncMap).ParseGlob(path.Join(loader.Directory, "[a-z]*.html"))
}

func (loader *TemplateLoader) makeErrorTemplate(errorMessage string) *template.Template {
	return template.Must(template.New("error").Parse(fmt.Sprintf("Error in template: %s", errorMessage)))
}

// NewTemplateLoader returns a template loader with templates from
// directory. If directory is "", fallbackTemplate will be used
// (regardless of the wanted template name). If debug is true,
// templates will be recompiled each time.
func NewTemplateLoader(directory string, debug bool) *TemplateLoader {
	loader := &TemplateLoader{Directory: directory}
	if directory == "" {
		loader.usesDummy = true
		loader.template = template.Must(template.New("dummy").Funcs(FuncMap).Parse(`
<!DOCTYPE HTML>
<html>
<head>
<style>
    html {
      font-family: monospace;
    }
    dd {
      margin-left: 2em;
    }
</style>
</head>
<body>
  {{ htmlize . }}
</body>
</html>
`))
		return loader
	}
	if !debug {
		tmpl, err := loader.compile()
		if err != nil {
			panic(err)
		}
		loader.template = tmpl
	}
	return loader
}

// Lookup will find a template by name and return it
func (loader *TemplateLoader) Lookup(name string) (*template.Template, error) {
	if loader.usesDummy {
		return loader.template, nil
	}
	var err error
	source := loader.template
	if source == nil {
		source, err = loader.compile()
		if err != nil {
			return nil, err
		}
	}
	tmpl := source.Lookup(name)
	if tmpl == nil {
		err := fmt.Errorf("template %v not available", name)
		return nil, err
	}
	return tmpl, nil
}

// ServeTemplate executes the named template passing data into it. If
// the format GET parameter is equal to "json", serves data as JSON
// instead.
func (loader *TemplateLoader) ServeTemplate(templateName string, data interface{}, w http.ResponseWriter, r *http.Request) {
	switch r.URL.Query().Get("format") {
	case "json":
		j, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			httpErrorf(w, r, "JSON error%s", err)
			return
		}
		w.Write(j)
	default:
		tmpl, err := loader.Lookup(templateName)
		if err != nil {
			httpErrorf(w, r, "error in template loader: %v", err)
			return
		}
		if err := tmpl.Execute(w, data); err != nil {
			httpErrorf(w, r, "error executing template: %v", err)
		}
	}
}

var (
	modifyDbTopology     func(context.Context, topo.Server, *topotools.Topology) error
	modifyDbServingGraph func(context.Context, topo.Server, *topotools.ServingGraph)
)

// SetDbTopologyPostprocessor installs a hook that can modify
// topotools.Topology struct before it's displayed.
func SetDbTopologyPostprocessor(f func(context.Context, topo.Server, *topotools.Topology) error) {
	if modifyDbTopology != nil {
		panic("Cannot set multiple DbTopology postprocessors")
	}
	modifyDbTopology = f
}

// SetDbServingGraphPostprocessor installs a hook that can modify
// topotools.ServingGraph struct before it's displayed.
func SetDbServingGraphPostprocessor(f func(context.Context, topo.Server, *topotools.ServingGraph)) {
	if modifyDbServingGraph != nil {
		panic("Cannot set multiple DbServingGraph postprocessors")
	}
	modifyDbServingGraph = f
}
