// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vtgate

import (
	mproto "github.com/youtube/vitess/go/mysql/proto"
	"github.com/youtube/vitess/go/vt/proto/vtrpc"
	"github.com/youtube/vitess/go/vt/vterrors"
)

// A list of all vtrpc.ErrorCodes, ordered by priority. These priorities are
// used when aggregating multiple errors in VtGate.
// Higher priority error codes are more urgent for users to see. They are
// prioritized based on the following question: assuming a scatter query produced multiple
// errors, which of the errors is the most likely to give the user useful information
// about why the query failed and how they should proceed?
const (
	PrioritySuccess = iota
	PriorityTransientError
	PriorityQueryNotServed
	PriorityDeadlineExceeded
	PriorityCancelled
	PriorityIntegrityError
	PriorityNotInTx
	PriorityUnknownError
	PriorityInternalError
	PriorityResourceExhausted
	PriorityUnauthenticated
	PriorityPermissionDenied
	PriorityBadInput
)

var errorPriorities = map[vtrpc.ErrorCode]int{
	vtrpc.ErrorCode_SUCCESS:            PrioritySuccess,
	vtrpc.ErrorCode_CANCELLED:          PriorityCancelled,
	vtrpc.ErrorCode_UNKNOWN_ERROR:      PriorityUnknownError,
	vtrpc.ErrorCode_BAD_INPUT:          PriorityBadInput,
	vtrpc.ErrorCode_DEADLINE_EXCEEDED:  PriorityDeadlineExceeded,
	vtrpc.ErrorCode_INTEGRITY_ERROR:    PriorityIntegrityError,
	vtrpc.ErrorCode_PERMISSION_DENIED:  PriorityPermissionDenied,
	vtrpc.ErrorCode_RESOURCE_EXHAUSTED: PriorityResourceExhausted,
	vtrpc.ErrorCode_QUERY_NOT_SERVED:   PriorityQueryNotServed,
	vtrpc.ErrorCode_NOT_IN_TX:          PriorityNotInTx,
	vtrpc.ErrorCode_INTERNAL_ERROR:     PriorityInternalError,
	vtrpc.ErrorCode_TRANSIENT_ERROR:    PriorityTransientError,
	vtrpc.ErrorCode_UNAUTHENTICATED:    PriorityUnauthenticated,
}

// aggregateVtGateErrorCodes aggregates a list of errors into a single error code.
// It does so by finding the highest priority error code in the list.
func aggregateVtGateErrorCodes(errors []error) vtrpc.ErrorCode {
	highCode := vtrpc.ErrorCode_SUCCESS
	for _, e := range errors {
		code := vterrors.RecoverVtErrorCode(e)
		if errorPriorities[code] > errorPriorities[highCode] {
			highCode = code
		}
	}
	return highCode
}

// AggregateVtGateErrors aggregates several VtErrors.
func AggregateVtGateErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}
	return vterrors.FromError(
		aggregateVtGateErrorCodes(errors),
		vterrors.ConcatenateErrors(errors),
	)
}

// AddVtGateError will update a mproto.RPCError with details from a VTGate error.
func AddVtGateError(err error, replyErr **mproto.RPCError) {
	if err == nil {
		return
	}
	*replyErr = vterrors.RPCErrFromVtError(err)
}

// RPCErrorToVtRPCError converts a VTGate error into a vtrpc error.
func RPCErrorToVtRPCError(rpcErr *mproto.RPCError) *vtrpc.RPCError {
	if rpcErr == nil {
		return nil
	}
	return &vtrpc.RPCError{
		Code:    vtrpc.ErrorCode(rpcErr.Code),
		Message: rpcErr.Message,
	}
}
