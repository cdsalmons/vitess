// Copyright 2015, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proto

import (
	"reflect"
	"testing"

	"github.com/youtube/vitess/go/sqltypes"
	"github.com/youtube/vitess/go/vt/proto/query"
)

func TestBindVariablesToProto3(t *testing.T) {
	testcases := []struct {
		name string
		in   interface{}
		out  *query.BindVariable
	}{{
		name: "string",
		in:   "aa",
		out: &query.BindVariable{
			Type:  sqltypes.VarChar,
			Value: []byte("aa"),
		},
	}, {
		name: "[]byte",
		in:   []byte("aa"),
		out: &query.BindVariable{
			Type:  sqltypes.VarBinary,
			Value: []byte("aa"),
		},
	}, {
		name: "int",
		in:   int(1),
		out: &query.BindVariable{
			Type:  sqltypes.Int64,
			Value: []byte("1"),
		},
	}, {
		name: "int8",
		in:   int8(-1),
		out: &query.BindVariable{
			Type:  sqltypes.Int64,
			Value: []byte("-1"),
		},
	}, {
		name: "int16",
		in:   int16(-1),
		out: &query.BindVariable{
			Type:  sqltypes.Int64,
			Value: []byte("-1"),
		},
	}, {
		name: "int32",
		in:   int32(-1),
		out: &query.BindVariable{
			Type:  sqltypes.Int64,
			Value: []byte("-1"),
		},
	}, {
		name: "int64",
		in:   int64(-1),
		out: &query.BindVariable{
			Type:  sqltypes.Int64,
			Value: []byte("-1"),
		},
	}, {
		name: "uint",
		in:   uint(1),
		out: &query.BindVariable{
			Type:  sqltypes.Uint64,
			Value: []byte("1"),
		},
	}, {
		name: "uint8",
		in:   uint8(1),
		out: &query.BindVariable{
			Type:  sqltypes.Uint64,
			Value: []byte("1"),
		},
	}, {
		name: "uint16",
		in:   uint16(1),
		out: &query.BindVariable{
			Type:  sqltypes.Uint64,
			Value: []byte("1"),
		},
	}, {
		name: "uint32",
		in:   uint32(1),
		out: &query.BindVariable{
			Type:  sqltypes.Uint64,
			Value: []byte("1"),
		},
	}, {
		name: "uint64",
		in:   uint64(1),
		out: &query.BindVariable{
			Type:  sqltypes.Uint64,
			Value: []byte("1"),
		},
	}, {
		name: "float32",
		in:   float32(1.5),
		out: &query.BindVariable{
			Type:  sqltypes.Float64,
			Value: []byte("1.5"),
		},
	}, {
		name: "float64",
		in:   float64(1.5),
		out: &query.BindVariable{
			Type:  sqltypes.Float64,
			Value: []byte("1.5"),
		},
	}, {
		name: "sqltypes.NULL",
		in:   sqltypes.NULL,
		out:  &query.BindVariable{},
	}, {
		name: "nil",
		in:   nil,
		out:  &query.BindVariable{},
	}, {
		name: "sqltypes.Integral",
		in:   sqltypes.MakeNumeric([]byte("1")),
		out: &query.BindVariable{
			Type:  sqltypes.Int64,
			Value: []byte("1"),
		},
	}, {
		name: "sqltypes.Fractional",
		in:   sqltypes.MakeFractional([]byte("1.5")),
		out: &query.BindVariable{
			Type:  sqltypes.Float64,
			Value: []byte("1.5"),
		},
	}, {
		name: "sqltypes.String",
		in:   sqltypes.MakeString([]byte("aa")),
		out: &query.BindVariable{
			Type:  sqltypes.VarBinary,
			Value: []byte("aa"),
		},
	}, {
		name: "[]interface{}",
		in:   []interface{}{1, "aa", sqltypes.MakeFractional([]byte("1.5"))},
		out: &query.BindVariable{
			Type: sqltypes.Tuple,
			Values: []*query.Value{
				&query.Value{
					Type:  sqltypes.Int64,
					Value: []byte("1"),
				},
				&query.Value{
					Type:  sqltypes.VarChar,
					Value: []byte("aa"),
				},
				&query.Value{
					Type:  sqltypes.Float64,
					Value: []byte("1.5"),
				},
			},
		},
	}, {
		name: "[]string",
		in:   []string{"aa", "bb"},
		out: &query.BindVariable{
			Type: sqltypes.Tuple,
			Values: []*query.Value{
				&query.Value{
					Type:  sqltypes.VarChar,
					Value: []byte("aa"),
				},
				&query.Value{
					Type:  sqltypes.VarChar,
					Value: []byte("bb"),
				},
			},
		},
	}, {
		name: "[][]byte",
		in:   [][]byte{[]byte("aa"), []byte("bb")},
		out: &query.BindVariable{
			Type: sqltypes.Tuple,
			Values: []*query.Value{
				&query.Value{
					Type:  sqltypes.VarBinary,
					Value: []byte("aa"),
				},
				&query.Value{
					Type:  sqltypes.VarBinary,
					Value: []byte("bb"),
				},
			},
		},
	}, {
		name: "[]int",
		in:   []int{1, 2},
		out: &query.BindVariable{
			Type: sqltypes.Tuple,
			Values: []*query.Value{
				&query.Value{
					Type:  sqltypes.Int64,
					Value: []byte("1"),
				},
				&query.Value{
					Type:  sqltypes.Int64,
					Value: []byte("2"),
				},
			},
		},
	}, {
		name: "[]int64",
		in:   []int64{1, 2},
		out: &query.BindVariable{
			Type: sqltypes.Tuple,
			Values: []*query.Value{
				&query.Value{
					Type:  sqltypes.Int64,
					Value: []byte("1"),
				},
				&query.Value{
					Type:  sqltypes.Int64,
					Value: []byte("2"),
				},
			},
		},
	}, {
		name: "[]uint64",
		in:   []uint64{1, 2},
		out: &query.BindVariable{
			Type: sqltypes.Tuple,
			Values: []*query.Value{
				&query.Value{
					Type:  sqltypes.Uint64,
					Value: []byte("1"),
				},
				&query.Value{
					Type:  sqltypes.Uint64,
					Value: []byte("2"),
				},
			},
		},
	}}
	for _, tcase := range testcases {
		bv := map[string]interface{}{
			"bv": tcase.in,
		}
		p3, err := BindVariablesToProto3(bv)
		if err != nil {
			t.Errorf("Error on %v: %v", tcase.name, err)
		}
		if !reflect.DeepEqual(p3["bv"], tcase.out) {
			t.Errorf("Mismatch on %v: %+v, want %+v", tcase.name, p3["bv"], tcase.out)
		}
	}
}

func TestBindVariablesToProto3Errors(t *testing.T) {
	testcases := []struct {
		name string
		in   interface{}
		out  string
	}{{
		name: "chan",
		in:   make(chan bool),
		out:  "key: bv: unexpected type chan bool",
	}, {
		name: "empty []interface{}",
		in:   []interface{}{},
		out:  "empty list not allowed: bv",
	}, {
		name: "empty []string",
		in:   []string{},
		out:  "empty list not allowed: bv",
	}, {
		name: "empty [][]byte",
		in:   [][]byte{},
		out:  "empty list not allowed: bv",
	}, {
		name: "empty []int",
		in:   []int{},
		out:  "empty list not allowed: bv",
	}, {
		name: "empty []int64",
		in:   []int64{},
		out:  "empty list not allowed: bv",
	}, {
		name: "empty []uint64",
		in:   []uint64{},
		out:  "empty list not allowed: bv",
	}, {
		name: "chan in []interface{}",
		in:   []interface{}{make(chan bool)},
		out:  "key: bv: unexpected type chan bool",
	}}
	for _, tcase := range testcases {
		bv := map[string]interface{}{
			"bv": tcase.in,
		}
		_, err := BindVariablesToProto3(bv)
		if err == nil || err.Error() != tcase.out {
			t.Errorf("Error: %v, want %v", err, tcase.out)
		}
	}
}

func TestProto3ToBindVariables(t *testing.T) {
	testcases := []struct {
		name string
		in   *query.BindVariable
		out  interface{}
	}{{
		name: "Int16",
		in: &query.BindVariable{
			Type:  sqltypes.Int16,
			Value: []byte("-1"),
		},
		out: int64(-1),
	}, {
		name: "Uint16",
		in: &query.BindVariable{
			Type:  sqltypes.Uint16,
			Value: []byte("1"),
		},
		out: uint64(1),
	}, {
		name: "Float64",
		in: &query.BindVariable{
			Type:  sqltypes.Float64,
			Value: []byte("1.5"),
		},
		out: float64(1.5),
	}, {
		name: "VarChar",
		in: &query.BindVariable{
			Type:  sqltypes.VarChar,
			Value: []byte("aa"),
		},
		out: []byte("aa"),
	}, {
		name: "Null",
		in: &query.BindVariable{
			Type: sqltypes.Null,
		},
		out: nil,
	}, {
		name: "nil",
		in:   nil,
		out:  nil,
	}, {
		name: "Tuple",
		in: &query.BindVariable{
			Type: sqltypes.Tuple,
			Values: []*query.Value{
				&query.Value{
					Type:  sqltypes.Int64,
					Value: []byte("1"),
				},
				&query.Value{
					Type:  sqltypes.VarChar,
					Value: []byte("aa"),
				},
				&query.Value{
					Type:  sqltypes.Float64,
					Value: []byte("1.5"),
				},
			},
		},
		out: []interface{}{int64(1), []byte("aa"), float64(1.5)},
	}}
	for _, tcase := range testcases {
		p3 := map[string]*query.BindVariable{
			"bv": tcase.in,
		}
		bv, err := Proto3ToBindVariables(p3)
		if err != nil {
			t.Errorf("Error on %v: %v", tcase.name, err)
		}
		if !reflect.DeepEqual(bv["bv"], tcase.out) {
			t.Errorf("Mismatch on %v: %+v, want %+v", tcase.name, bv["bv"], tcase.out)
		}
	}
}

func TestProto3ToBindVariablesErrors(t *testing.T) {
	testcases := []struct {
		name string
		in   *query.BindVariable
		out  string
	}{{
		name: "Int64",
		in: &query.BindVariable{
			Type:  sqltypes.Int64,
			Value: []byte("aa"),
		},
		out: `strconv.ParseInt: parsing "aa": invalid syntax`,
	}, {
		name: "Uint64",
		in: &query.BindVariable{
			Type:  sqltypes.Uint64,
			Value: []byte("-1"),
		},
		out: `strconv.ParseUint: parsing "-1": invalid syntax`,
	}, {
		name: "Float64",
		in: &query.BindVariable{
			Type:  sqltypes.Float64,
			Value: []byte("aa"),
		},
		out: `strconv.ParseFloat: parsing "aa": invalid syntax`,
	}, {
		name: "Tuple",
		in: &query.BindVariable{
			Type: sqltypes.Tuple,
			Values: []*query.Value{
				&query.Value{
					Type:  sqltypes.Int64,
					Value: []byte("aa"),
				},
			},
		},
		out: `strconv.ParseInt: parsing "aa": invalid syntax`,
	}}
	for _, tcase := range testcases {
		p3 := map[string]*query.BindVariable{
			"bv": tcase.in,
		}
		_, err := Proto3ToBindVariables(p3)
		if err == nil || err.Error() != tcase.out {
			t.Errorf("Error: %v, want %v", err, tcase.out)
		}
	}
}
