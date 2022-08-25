package pjrpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"strings"

	"gitlab.com/pjrpc/pjrpc/v2/pjson"
	"gitlab.com/pjrpc/pjrpc/v2/storage"
)

var (
	_ http.Handler = &ServerHTTP{}
	_ Registrator  = &ServerHTTP{}
)

// ServerHTTP it is a http handler with JSON-RPC router.
type ServerHTTP struct {
	// Contains handlers by methods.
	Router *Router

	// Logger is a optional logger that writes errors in write response method.
	// Also writes panics in handlers.
	Logger *log.Logger

	// Panic handler calls when your rpc handler make panic.
	// There is default panic handler (DefaultRestoreOnPanic) after created by New method.
	OnPanic func(resp *Response, err error)

	// OnFailedParseRequest handler calls when server can't parse client request.
	// There is default handler (DefaultOnFailedParseRequest) after created by New method.
	OnFailedParseRequest func(r *http.Request, err error) *ErrorResponse
}

// serveResult represents http request with parsed body.
// It must contain at least one request.
// Each request has response empty or with parse error.
type serveResult struct {
	requests  BatchRequests
	responses BatchResponses
	isBatch   bool
}

// NewServerHTTP creates new server with default panic handler and empty router and logger.
func NewServerHTTP() *ServerHTTP {
	s := &ServerHTTP{
		Router: NewRouter(),
		Logger: nil,

		OnPanic:              nil,
		OnFailedParseRequest: nil,
	}

	s.OnPanic = s.DefaultRestoreOnPanic
	s.OnFailedParseRequest = s.DefaultOnFailedParseRequest

	return s
}

// ServeHTTP implements interface of handler in http package.
func (s *ServerHTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := s.parseRequest(r)
	if err != nil {
		s.sendError(w, res, err)
		return
	}

	for i := 0; i < len(res.requests); i++ {
		s.invokeMethod(res, i, r)
	}

	s.sendServeResult(w, res)
}

// RegisterMethod saves handler of the method to starage.
func (s *ServerHTTP) RegisterMethod(methodName string, h Handler) {
	s.Router.RegisterMethod(methodName, h)
}

// With adds middlewares to the router's queue of the middlewares.
func (s *ServerHTTP) With(mws ...Middleware) {
	s.Router.With(mws...)
}

// DefaultRestoreOnPanic is a default panic handler that converts panic in JSON-RPC protocol error.
// It prints error to log if it set and puts normal response body with JSON-RPC error.
func (s *ServerHTTP) DefaultRestoreOnPanic(resp *Response, err error) {
	s.logError("DefaultRestoreOnPanic", "panic", err)
	resp.SetError(JRPCErrInternalError())
}

// DefaultOnFailedParseRequest default handler calls when server can't to parse client request.
// Returns JSON-RPC error with static text.
func (s *ServerHTTP) DefaultOnFailedParseRequest(r *http.Request, err error) *ErrorResponse {
	return JRPCErrInvalidRequest("failed to parse request")
}

// SetLogger sets your io.Writer as error logger of the pjrpc server.
// Also you can create and set your own *log.Logger in Logger field.
func (s *ServerHTTP) SetLogger(w io.Writer) {
	s.Logger = log.New(w, "[pjrpc-server] ", 0)
}

func (s *ServerHTTP) logError(place, message string, err error) {
	if s.Logger == nil {
		return
	}

	s.Logger.Printf("(%s): %s: %s\n", place, message, err)
}

func convertError(err error) *ErrorResponse {
	jrpcErr := &ErrorResponse{}
	if errors.As(err, &jrpcErr) {
		return jrpcErr
	}

	if errors.Is(err, storage.ErrRouteNotFound) {
		return JRPCErrMethodNotFound()
	}

	return JRPCErrServerError(-32000, err.Error())
}

func isRequestValid(req *Request) error {
	if req.JSONRPC != JSONRPCVersion {
		return JRPCErrInvalidRequest("wrong value of the field 'jsonrpc'")
	}

	return nil
}

func createResponse(req *Request) *Response {
	return &Response{
		JSONRPC: JSONRPCVersion,
		ID:      req.ID,
		Result:  nil,
		Error:   nil,
	}
}

func parseSingleRequest(res *serveResult, decoder *pjson.Decoder) error {
	r := new(Request)
	err := decoder.Decode(&r)
	if err != nil {
		return fmt.Errorf("decoder.Decode: %w", err)
	}

	resp := createResponse(r)
	if err = isRequestValid(r); err != nil {
		resp.SetError(err)
	}

	res.requests = BatchRequests{r}
	res.responses = BatchResponses{resp}

	return nil
}

func parseBatchRequest(res *serveResult, decoder *pjson.Decoder) error {
	err := decoder.Decode(&res.requests)
	if err != nil {
		return fmt.Errorf("decoder.Decode: %w", err)
	}

	res.responses = make(BatchResponses, len(res.requests))

	for i, r := range res.requests {
		res.responses[i] = createResponse(r)

		if err = isRequestValid(r); err != nil {
			res.responses[i].SetError(err)
		}
	}

	return nil
}

func (s *ServerHTTP) parseRequest(r *http.Request) (*serveResult, error) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			s.logError("parseRequest", "failed to close request body", err)
		}
	}()

	if !strings.Contains(r.Header.Get(ContentTypeHeaderName), ContentTypeHeaderValue) {
		return nil, JRPCErrInvalidRequest("content-type must be 'application/json'")
	}

	contentLength := r.ContentLength
	if contentLength < 0 {
		contentLength = 0
	}

	body := bytes.NewBuffer(make([]byte, 0, contentLength))

	_, err := body.ReadFrom(r.Body)
	if err != nil {
		return nil, JRPCErrParseError("failed to read body")
	}

	if body.Len() == 0 {
		return nil, JRPCErrInvalidRequest("body length is 0")
	}

	firstSymbol, _, err := body.ReadRune()
	if err != nil {
		s.logError("parseRequest", "failed to body.ReadRune", err)
		return nil, JRPCErrInternalError()
	}
	if err = body.UnreadRune(); err != nil {
		s.logError("parseRequest", "failed to body.UnreadRune", err)
		return nil, JRPCErrInternalError()
	}

	res := &serveResult{
		requests:  nil,
		responses: nil,
		isBatch:   firstSymbol == '[',
	}

	decoder := pjson.NewDecoder(body)

	if res.isBatch {
		err = parseBatchRequest(res, decoder)
	} else {
		err = parseSingleRequest(res, decoder)
	}

	if err != nil {
		return res, s.OnFailedParseRequest(r, err)
	}

	return res, nil
}

func (s *ServerHTTP) sendResponseBody(w http.ResponseWriter, body json.RawMessage) {
	w.Header().Set(ContentTypeHeaderName, ContentTypeHeaderValue)
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(body); err != nil {
		s.logError("sendResponseBody", "failed to write response", err)
		return
	}
}

func (s *ServerHTTP) sendError(w http.ResponseWriter, res *serveResult, err error) {
	var srcBody interface{}

	resp := &Response{JSONRPC: JSONRPCVersion}
	resp.SetError(err)

	if res != nil && res.isBatch {
		srcBody = BatchResponses{resp}
	} else {
		srcBody = resp
	}

	body, _ := pjson.Marshal(srcBody) // nolint:errcheck // unreal error when encoding regular struct.

	s.sendResponseBody(w, body)
}

func (s *ServerHTTP) sendServeResult(w http.ResponseWriter, res *serveResult) {
	var body []byte

	if res.isBatch {
		body, _ = pjson.Marshal(res.responses) // nolint:errcheck // unreal error when encoding regular struct.
	} else {
		body, _ = pjson.Marshal(res.responses[0]) // nolint:errcheck // unreal error when encoding regular struct.
	}

	s.sendResponseBody(w, body)
}

func (s *ServerHTTP) restoreOnPanic(resp *Response, r interface{}, method string) {
	buf := make([]byte, 65536) // 64 << 10.
	buf = buf[:runtime.Stack(buf, false)]

	err := fmt.Errorf("%w: '%s': %v\n%s", ErrPanicInHandler, method, r, buf)

	if s.OnPanic != nil {
		s.OnPanic(resp, err)
	} else {
		s.logError("restorePanic", "panic", err)
	}

	if s.OnPanic == nil && s.Logger == nil {
		println(err.Error()) // nolint:forbidigo // This is the last way to print an error.
	}
}

func (s *ServerHTTP) invokeMethod(res *serveResult, index int, r *http.Request) {
	if res.responses[index].Error != nil { // Error could have been set in during parsing.
		return
	}

	jReq := res.requests[index]

	data := &ContextData{
		HTTTRequest: r,
		JRPCRequest: jReq,
		IsBatch:     res.isBatch,
	}

	ctx := ContextSetData(r.Context(), data)

	defer func() {
		if rec := recover(); rec != nil {
			s.restoreOnPanic(res.responses[index], rec, jReq.Method)
		}
	}()

	result, err := s.Router.Invoke(ctx, jReq.Method, jReq.Params)
	if err != nil {
		res.responses[index].SetError(err)
	} else {
		res.responses[index].SetResult(result)
	}
}
