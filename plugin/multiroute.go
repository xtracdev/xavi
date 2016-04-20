package plugin

import (
	"net/http"

	"golang.org/x/net/context"
)

//BackendHandlerMap provides a map from backend name to the handler associated with the backend.
type BackendHandlerMap map[string]ContextHandler

//MultiBackendHandler defines an HTTP handler interface that includes backend handler context.
type MultiBackendHandler interface {
	MultiBackendServeHTTP(BackendHandlerMap, context.Context, http.ResponseWriter, *http.Request)
}

//MultiBackendHandlerFunc defines a handler function that includes backend handler context
type MultiBackendHandlerFunc func(BackendHandlerMap, context.Context, http.ResponseWriter, *http.Request)

//MultiBackendServeHTTP is a method that invokes a MultiBackendHandlerFunc handler with the associated context
//and request/response arguments
func (h MultiBackendHandlerFunc) MultiBackendServeHTTP(bhMap BackendHandlerMap, ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h(bhMap, ctx, w, r)
}

//MultiBackendAdapter is a type used to adapt a handler function with an additional context argument to
//the standard HTTP handler function for use with golang's HTTP functions.
type MultiBackendAdapter struct {
	BackendHandlerCtx BackendHandlerMap
	Handler           MultiBackendHandler
}

func (mra *MultiBackendAdapter) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	mra.Handler.MultiBackendServeHTTP(mra.BackendHandlerCtx, ctx, w, r)
}

//ToHandlerFunc converts a MultiBackendAdapter to an http.HandlerFunc
func (mra *MultiBackendAdapter) ToHandlerFunc() ContextHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		mra.Handler.MultiBackendServeHTTP(mra.BackendHandlerCtx, ctx, w, r)
	}
}

//MultiBackendAdapterFactory defines a function type for instantiating a MultiBackendAdapter
//given a backend handler map.
type MultiBackendAdapterFactory func(BackendHandlerMap) *MultiBackendAdapter
