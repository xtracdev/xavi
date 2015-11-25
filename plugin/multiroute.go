package plugin

import (
	"net/http"
)

type BackendHandlerMap map[string]http.Handler

type MultiBackendHandler interface {
	MultiBackendServeHTTP(BackendHandlerMap, http.ResponseWriter, *http.Request)
}

type MultiBackendHandlerFunc func(BackendHandlerMap, http.ResponseWriter, *http.Request)

func (h MultiBackendHandlerFunc) MultiBackendServeHTTP(bhMap BackendHandlerMap, w http.ResponseWriter, r *http.Request) {
	h(bhMap, w, r)
}

type MultiBackendAdapter struct {
	Ctx     BackendHandlerMap
	Handler MultiBackendHandler
}

func (mra *MultiBackendAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mra.Handler.MultiBackendServeHTTP(mra.Ctx, w, r)
}

func (mra *MultiBackendAdapter) ToHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mra.Handler.MultiBackendServeHTTP(mra.Ctx, w, r)
	}
}

type MultiBackendAdapterFactory func(BackendHandlerMap) *MultiBackendAdapter
