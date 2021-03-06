// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binlog

import (
	"strings"

	log "github.com/golang/glog"

	pb "github.com/youtube/vitess/go/vt/proto/binlogdata"
)

var (
	STREAM_COMMENT = "/* _stream "
	SPACE          = " "
)

// TablesFilterFunc returns a function that calls sendReply only if statements
// in the transaction match the specified tables. The resulting function can be
// passed into the BinlogStreamer: bls.Stream(file, pos, sendTransaction) ->
// bls.Stream(file, pos, TablesFilterFunc(sendTransaction))
func TablesFilterFunc(tables []string, sendReply sendTransactionFunc) sendTransactionFunc {
	return func(reply *pb.BinlogTransaction) error {
		matched := false
		filtered := make([]*pb.BinlogTransaction_Statement, 0, len(reply.Statements))
		for _, statement := range reply.Statements {
			switch statement.Category {
			case pb.BinlogTransaction_Statement_BL_SET:
				filtered = append(filtered, statement)
			case pb.BinlogTransaction_Statement_BL_DDL:
				log.Warningf("Not forwarding DDL: %s", statement.Sql)
				continue
			case pb.BinlogTransaction_Statement_BL_DML:
				tableIndex := strings.LastIndex(statement.Sql, STREAM_COMMENT)
				if tableIndex == -1 {
					updateStreamErrors.Add("TablesStream", 1)
					log.Errorf("Error parsing table name: %s", statement.Sql)
					continue
				}
				tableStart := tableIndex + len(STREAM_COMMENT)
				tableEnd := strings.Index(statement.Sql[tableStart:], SPACE)
				if tableEnd == -1 {
					updateStreamErrors.Add("TablesStream", 1)
					log.Errorf("Error parsing table name: %s", statement.Sql)
					continue
				}
				tableName := statement.Sql[tableStart : tableStart+tableEnd]
				for _, t := range tables {
					if t == tableName {
						filtered = append(filtered, statement)
						matched = true
						break
					}
				}
			case pb.BinlogTransaction_Statement_BL_UNRECOGNIZED:
				updateStreamErrors.Add("TablesStream", 1)
				log.Errorf("Error parsing table name: %s", statement.Sql)
				continue
			}
		}
		if matched {
			reply.Statements = filtered
		} else {
			reply.Statements = nil
		}
		return sendReply(reply)
	}
}
