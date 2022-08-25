package pjrpc

import (
	"context"
	"encoding/json"
	"fmt"

	"gitlab.com/pjrpc/pjrpc/v2/storage"
)

// Handler request handler takes request context and raw json params.
// Returns type with result data and error.
type Handler func(ctx context.Context, params json.RawMessage) (interface{}, error)

// Middleware wraps handler and call before wrapped handler.
type Middleware func(next Handler) Handler

// Registrator provides methods to register handler and middlewares.
type Registrator interface {
	// RegisterMethod saves handler to router.
	RegisterMethod(methodName string, h Handler)
	// With adds midlewares to the queue. It will be call beafore methods.
	With(mws ...Middleware)
}

// Router contains storage with handlers and common middleware.
type Router struct {
	storage     *storage.Storage
	middlewares []Middleware
}

var _ Registrator = &Router{}

// NewRouter creates new router with empty handler storage.
func NewRouter() *Router {
	r := &Router{
		storage: storage.New(),
	}

	return r
}

// RegisterMethod saves handler of the method to starage.
func (r *Router) RegisterMethod(methodName string, h Handler) {
	r.storage.Put(methodName, h)
}

// With adds middlewares to the router's queue of the middlewares.
func (r *Router) With(mws ...Middleware) {
	if r.middlewares == nil {
		r.middlewares = mws
		return
	}

	r.middlewares = append(r.middlewares, mws...)
}

// MethodWith adds middleware to exclusive method, method must be registered.
func (r *Router) MethodWith(methodName string, mw Middleware) error {
	rawHandler, err := r.storage.Get(methodName)
	if err != nil {
		return fmt.Errorf("storage.Get: %w", err)
	}

	handler := rawHandler.(Handler) // nolint:errcheck,forcetypeassert // only Handler type here.
	r.storage.Put(methodName, mw(handler))

	return nil
}

// Invoke invokes handler by method name.
// If the router has middleware, it will be called first.
// After that the registered middleware will be called and then the registered handler.
func (r *Router) Invoke(ctx context.Context, methodName string, params json.RawMessage) (interface{}, error) {
	storedHandler, err := r.storage.Get(methodName)
	if err != nil {
		return nil, fmt.Errorf("storage.Get: %w", err)
	}

	handler := storedHandler.(Handler) // nolint:errcheck,forcetypeassert // we put in storage Handler type only.

	// Reverse order beacause middleware it's handler wrapper,
	// so the first registered mw must wrap handler lastly.
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		handler = r.middlewares[i](handler)
	}

	res, err := handler(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("method '%s': %w", methodName, err)
	}

	return res, nil
}
