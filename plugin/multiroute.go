package plugin

import (
	"net/http"
)

type BackendHandlerMap map[string]http.Handler

type MultirouteHandler interface {
	MultiRouteServeHTTP(BackendHandlerMap, http.ResponseWriter, *http.Request)
}

type MultiRouteHandlerFunc func(BackendHandlerMap, http.ResponseWriter, *http.Request)

func (h MultiRouteHandlerFunc) MultiRouteServeHTTP(bhMap BackendHandlerMap, w http.ResponseWriter, r *http.Request) {
	h(bhMap, w, r)
}

type MultiRouteAdapter struct {
	Ctx     BackendHandlerMap
	Handler MultirouteHandler
}

func (mra *MultiRouteAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mra.Handler.MultiRouteServeHTTP(mra.Ctx, w, r)
}

type MultiRouteAdapterFactory func(BackendHandlerMap) *MultiRouteAdapter
