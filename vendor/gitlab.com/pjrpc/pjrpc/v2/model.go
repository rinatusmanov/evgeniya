package pjrpc

import (
	"bytes"
	"encoding/json"
	"fmt"

	"gitlab.com/pjrpc/pjrpc/v2/pjson"
)

const (
	// ContentTypeHeaderName name of the header Content-Type.
	ContentTypeHeaderName = "Content-Type"
	// ContentTypeHeaderValue value of the header Content-Type.
	ContentTypeHeaderValue = "application/json"
	// JSONRPCVersion is a version of supported JSON-RPC protocol.
	JSONRPCVersion = "2.0"
)

// convertRawMessageToString trims " symbols from JSON RawMessage and returns value of the ID.
func convertRawMessageToString(m json.RawMessage) string {
	if m == nil {
		return ""
	}

	return string(bytes.Trim(m, `"`))
}

// Request model of the JSON-RPC request.
type Request struct {
	// A String specifying the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`
	// An identifier established by the Client that MUST contain a String, Number, or NULL value if included.
	ID json.RawMessage `json:"id,omitempty"`
	// A String containing the name of the method to be invoked.
	Method string `json:"method"`
	// A Structured value that holds the parameter values to be used during the invocation of the method.
	Params json.RawMessage `json:"params,omitempty"`
}

// GetID returns id of the request as a string.
func (r *Request) GetID() string {
	return convertRawMessageToString(r.ID)
}

// ErrorResponse model of the response with error.
type ErrorResponse struct {
	// A Number that indicates the error type that occurred.
	Code int `json:"code"`
	// A String providing a short description of the error.
	Message string `json:"message"`
	// A Primitive or Structured value that contains additional information about the error.
	Data json.RawMessage `json:"data,omitempty"`
}

// Error implementation error interface.
func (e *ErrorResponse) Error() string {
	if e == nil {
		return "<nil>"
	}

	tmpl := "JSON-RPC Error: [%d] %s"
	data := ""
	if len(e.Data) != 0 {
		tmpl += " (%s)"
		data = string(e.Data)
	} else {
		tmpl += "%s"
	}

	return fmt.Sprintf(tmpl, e.Code, e.Message, data)
}

// Response model of the response object.
type Response struct {
	// A String specifying the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`
	// It MUST be the same as the value of the id member in the Request.
	ID json.RawMessage `json:"id"`
	// This member is REQUIRED on success. The value of this member is determined
	// by the method invoked on the Server.
	Result json.RawMessage `json:"result,omitempty"`
	// This member is REQUIRED on error. This member MUST NOT exist if there was no error
	// triggered during invocation.
	Error *ErrorResponse `json:"error,omitempty"`
}

// GetID returns id of the response as a string.
func (r *Response) GetID() string {
	return convertRawMessageToString(r.ID)
}

// SetError sets error to Response model.
func (r *Response) SetError(err error) {
	r.Error = convertError(err)
	r.Result = nil
}

// SetResult sets result to Response model.
func (r *Response) SetResult(result interface{}) {
	resultJSON, err := pjson.Marshal(result)
	if err != nil {
		r.SetError(JRPCErrInternalError("failed to marshal response"))
		return
	}

	r.Result = resultJSON
	r.Error = nil
}

// BatchRequests client MAY send an Array filled with Request objects.
// See spec https://www.jsonrpc.org/specification#batch page.
type BatchRequests []*Request

// BatchResponses the Server should respond with an Array containing the corresponding Response objects,
// after all of the batch Request objects have been processed.
type BatchResponses []*Response
