// Copyright 2015, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binlogplayertest

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"

	mproto "github.com/youtube/vitess/go/mysql/proto"
	"github.com/youtube/vitess/go/sqltypes"
	"github.com/youtube/vitess/go/vt/binlog/binlogplayer"
	"github.com/youtube/vitess/go/vt/binlog/proto"
	"github.com/youtube/vitess/go/vt/key"

	pb "github.com/youtube/vitess/go/vt/proto/binlogdata"
	pbt "github.com/youtube/vitess/go/vt/proto/topodata"
)

// keyRangeRequest is used to make a request for StreamKeyRange.
type keyRangeRequest struct {
	Position       string
	KeyspaceIdType pbt.KeyspaceIdType
	KeyRange       *pbt.KeyRange
	Charset        *pb.Charset
}

// tablesRequest is used to make a request for StreamTables.
type tablesRequest struct {
	Position string
	Tables   []string
	Charset  *pb.Charset
}

// FakeBinlogStreamer is our implementation of UpdateStream
type FakeBinlogStreamer struct {
	t      *testing.T
	panics bool
}

// NewFakeBinlogStreamer returns the test instance for UpdateStream
func NewFakeBinlogStreamer(t *testing.T) *FakeBinlogStreamer {
	return &FakeBinlogStreamer{
		t:      t,
		panics: false,
	}
}

//
// ServeUpdateStream tests
//

var testUpdateStreamRequest = "UpdateStream starting position"

var testStreamEvent = &proto.StreamEvent{
	Category:  "DML",
	TableName: "table1",
	PrimaryKeyFields: []mproto.Field{
		mproto.Field{
			Name:  "id",
			Type:  254,
			Flags: 128,
		},
	},
	PrimaryKeyValues: [][]sqltypes.Value{
		[]sqltypes.Value{
			sqltypes.MakeString([]byte("123")),
		},
	},
	Sql:           "test sql",
	Timestamp:     372,
	TransactionID: "StreamEvent returned transaction id",
}

// ServeUpdateStream is part of the the UpdateStream interface
func (fake *FakeBinlogStreamer) ServeUpdateStream(position string, sendReply func(reply *proto.StreamEvent) error) error {
	if fake.panics {
		panic(fmt.Errorf("test-triggered panic"))
	}
	if position != testUpdateStreamRequest {
		fake.t.Errorf("wrong ServeUpdateStream parameter, got %v want %v", position, testUpdateStreamRequest)
	}
	sendReply(testStreamEvent)
	return nil
}

func testServeUpdateStream(t *testing.T, bpc binlogplayer.Client) {
	ctx := context.Background()
	c, errFunc, err := bpc.ServeUpdateStream(ctx, testUpdateStreamRequest)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if se, ok := <-c; !ok {
		t.Fatalf("got no response")
	} else {
		if !reflect.DeepEqual(*se, *testStreamEvent) {
			t.Errorf("got wrong result, got \n%#v expected \n%#v", *se, *testStreamEvent)
		}
	}
	if se, ok := <-c; ok {
		t.Fatalf("got a response when error expected: %v", se)
	}
	if err := errFunc(); err != nil {
		t.Errorf("got unexpected error: %v", err)
	}
}

func testServeUpdateStreamPanics(t *testing.T, bpc binlogplayer.Client) {
	ctx := context.Background()
	c, errFunc, err := bpc.ServeUpdateStream(ctx, testUpdateStreamRequest)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if se, ok := <-c; ok {
		t.Fatalf("got a response when error expected: %v", se)
	}
	err = errFunc()
	if err == nil || !strings.Contains(err.Error(), "test-triggered panic") {
		t.Errorf("wrong error from panic: %v", err)
	}
}

//
// StreamKeyRange tests
//

var testKeyRangeRequest = &keyRangeRequest{
	Position:       "KeyRange starting position",
	KeyspaceIdType: pbt.KeyspaceIdType_UINT64,
	KeyRange: &pbt.KeyRange{
		Start: key.Uint64Key(0x7000000000000000).Bytes(),
		End:   key.Uint64Key(0x9000000000000000).Bytes(),
	},
	Charset: &pb.Charset{
		Client: 12,
		Conn:   13,
		Server: 14,
	},
}

var testBinlogTransaction = &pb.BinlogTransaction{
	Statements: []*pb.BinlogTransaction_Statement{
		{
			Category: pb.BinlogTransaction_Statement_BL_ROLLBACK,
			Charset: &pb.Charset{
				Client: 120,
				Conn:   130,
				Server: 140,
			},
			Sql: "my statement",
		},
	},
	Timestamp:     78,
	TransactionId: "BinlogTransaction returned transaction id",
}

// StreamKeyRange is part of the the UpdateStream interface
func (fake *FakeBinlogStreamer) StreamKeyRange(position string, keyspaceIdType pbt.KeyspaceIdType, keyRange *pbt.KeyRange, charset *pb.Charset, sendReply func(reply *pb.BinlogTransaction) error) error {
	if fake.panics {
		panic(fmt.Errorf("test-triggered panic"))
	}
	req := &keyRangeRequest{
		Position:       position,
		KeyspaceIdType: keyspaceIdType,
		KeyRange:       keyRange,
		Charset:        charset,
	}
	if !reflect.DeepEqual(req, testKeyRangeRequest) {
		fake.t.Errorf("wrong StreamKeyRange parameter, got %+v want %+v", req, testKeyRangeRequest)
	}
	sendReply(testBinlogTransaction)
	return nil
}

func testStreamKeyRange(t *testing.T, bpc binlogplayer.Client) {
	ctx := context.Background()
	c, errFunc, err := bpc.StreamKeyRange(ctx, testKeyRangeRequest.Position, testKeyRangeRequest.KeyspaceIdType, testKeyRangeRequest.KeyRange, testKeyRangeRequest.Charset)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if se, ok := <-c; !ok {
		t.Fatalf("got no response")
	} else {
		if !reflect.DeepEqual(*se, *testBinlogTransaction) {
			t.Errorf("got wrong result, got %v expected %v", *se, *testBinlogTransaction)
		}
	}
	if se, ok := <-c; ok {
		t.Fatalf("got a response when error expected: %v", se)
	}
	if err := errFunc(); err != nil {
		t.Errorf("got unexpected error: %v", err)
	}
}

func testStreamKeyRangePanics(t *testing.T, bpc binlogplayer.Client) {
	ctx := context.Background()
	c, errFunc, err := bpc.StreamKeyRange(ctx, testKeyRangeRequest.Position, testKeyRangeRequest.KeyspaceIdType, testKeyRangeRequest.KeyRange, testKeyRangeRequest.Charset)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if se, ok := <-c; ok {
		t.Fatalf("got a response when error expected: %v", se)
	}
	err = errFunc()
	if err == nil || !strings.Contains(err.Error(), "test-triggered panic") {
		t.Errorf("wrong error from panic: %v", err)
	}
}

//
// StreamTables test
//

var testTablesRequest = &tablesRequest{
	Position: "Tables starting position",
	Tables:   []string{"table1", "table2"},
	Charset: &pb.Charset{
		Client: 12,
		Conn:   13,
		Server: 14,
	},
}

// StreamTables is part of the the UpdateStream interface
func (fake *FakeBinlogStreamer) StreamTables(position string, tables []string, charset *pb.Charset, sendReply func(reply *pb.BinlogTransaction) error) error {
	if fake.panics {
		panic(fmt.Errorf("test-triggered panic"))
	}
	req := &tablesRequest{
		Position: position,
		Tables:   tables,
		Charset:  charset,
	}
	if !reflect.DeepEqual(req, testTablesRequest) {
		fake.t.Errorf("wrong StreamTables parameter, got %+v want %+v", req, testTablesRequest)
	}
	sendReply(testBinlogTransaction)
	return nil
}

func testStreamTables(t *testing.T, bpc binlogplayer.Client) {
	ctx := context.Background()
	c, errFunc, err := bpc.StreamTables(ctx, testTablesRequest.Position, testTablesRequest.Tables, testTablesRequest.Charset)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if se, ok := <-c; !ok {
		t.Fatalf("got no response")
	} else {
		if !reflect.DeepEqual(*se, *testBinlogTransaction) {
			t.Errorf("got wrong result, got %v expected %v", *se, *testBinlogTransaction)
		}
	}
	if se, ok := <-c; ok {
		t.Fatalf("got a response when error expected: %v", se)
	}
	if err := errFunc(); err != nil {
		t.Errorf("got unexpected error: %v", err)
	}
}

func testStreamTablesPanics(t *testing.T, bpc binlogplayer.Client) {
	ctx := context.Background()
	c, errFunc, err := bpc.StreamTables(ctx, testTablesRequest.Position, testTablesRequest.Tables, testTablesRequest.Charset)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if se, ok := <-c; ok {
		t.Fatalf("got a response when error expected: %v", se)
	}
	err = errFunc()
	if err == nil || !strings.Contains(err.Error(), "test-triggered panic") {
		t.Errorf("wrong error from panic: %v", err)
	}
}

// HandlePanic is part of the the UpdateStream interface
func (fake *FakeBinlogStreamer) HandlePanic(err *error) {
	if x := recover(); x != nil {
		*err = fmt.Errorf("Caught panic: %v", x)
	}
}

// Run runs the test suite
func Run(t *testing.T, bpc binlogplayer.Client, endPoint *pbt.EndPoint, fake *FakeBinlogStreamer) {
	if err := bpc.Dial(endPoint, 30*time.Second); err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	// no panic
	testServeUpdateStream(t, bpc)
	testStreamKeyRange(t, bpc)
	testStreamTables(t, bpc)

	// panic now, and test
	fake.panics = true
	testServeUpdateStreamPanics(t, bpc)
	testStreamKeyRangePanics(t, bpc)
	testStreamTablesPanics(t, bpc)
}
