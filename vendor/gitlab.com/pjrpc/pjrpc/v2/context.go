package pjrpc

import (
	"context"
	"net/http"
)

type ctxKey int

const ctxKeyData ctxKey = 1

// ContextData router data of the request as context.
type ContextData struct {
	// It's raw HTTP request from router.
	HTTTRequest *http.Request
	// It's parsed body of the request.
	JRPCRequest *Request
	// It will be true if this JSON-RPC requst is a part of the one request.
	IsBatch bool
}

// ContextSetData sets ContextData to the context.
func ContextSetData(ctx context.Context, data *ContextData) context.Context {
	return context.WithValue(ctx, ctxKeyData, data)
}

// ContextGetData returns ContextData from the context.
func ContextGetData(ctx context.Context) (d *ContextData, ok bool) {
	d, ok = ctx.Value(ctxKeyData).(*ContextData)
	return d, ok
}
