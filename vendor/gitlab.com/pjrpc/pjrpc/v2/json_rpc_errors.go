package pjrpc

import (
	"encoding/json"
)

func newJRPCError(code int, message string, data ...interface{}) *ErrorResponse {
	e := &ErrorResponse{
		Code:    code,
		Message: message,
		Data:    nil,
	}

	if len(data) != 0 {
		var err error
		e.Data, err = json.Marshal(data[0])
		if err != nil {
			// just don't use bad type in data.
			panic("can't json.Marshal error data: " + err.Error())
		}
	}

	return e
}

// JRPCErrParseError invalid JSON was received by the server.
// An error occurred on the server while parsing the JSON text.
func JRPCErrParseError(data ...interface{}) *ErrorResponse {
	return newJRPCError(-32700, "Parse error", data...)
}

// JRPCErrInvalidRequest the JSON sent is not a valid Request object.
func JRPCErrInvalidRequest(data ...interface{}) *ErrorResponse {
	return newJRPCError(-32600, "Invalid Request", data...)
}

// JRPCErrMethodNotFound the method does not exist / is not available.
func JRPCErrMethodNotFound(data ...interface{}) *ErrorResponse {
	return newJRPCError(-32601, "Method not found", data...)
}

// JRPCErrInvalidParams invalid method parameter(s).
func JRPCErrInvalidParams(data ...interface{}) *ErrorResponse {
	return newJRPCError(-32602, "Invalid params", data...)
}

// JRPCErrInternalError internal JSON-RPC error.
func JRPCErrInternalError(data ...interface{}) *ErrorResponse {
	return newJRPCError(-32603, "Internal error", data...)
}

// JRPCErrServerError reserved for implementation-defined server-errors.
// Codes -32000 to -32099.
func JRPCErrServerError(code int, data ...interface{}) *ErrorResponse {
	return newJRPCError(code, "Server error", data...)
}
