package pjrpc

import (
	"errors"
)

var (
	// ErrPanicInHandler returns in the server when handler called panic.
	ErrPanicInHandler = errors.New("panic in handler")

	// ErrBadStatusCode returns in client when http response has code not 200.
	ErrBadStatusCode = errors.New("bad status code")

	// ErrWrongContentType returns in client when http response has content-type not application/json.
	ErrWrongContentType = errors.New("wrong content type")
)
