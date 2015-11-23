package plugin

import (
	"net/http"
)

type backendHandlerMap map[string]http.Handler

type MultirouteHandler interface {
	MultiRouteServeHTTP(backendHandlerMap, http.ResponseWriter, *http.Request)
}

type MultiRouteHandlerFunc func(backendHandlerMap, http.ResponseWriter, *http.Request)

func (h MultiRouteHandlerFunc) MultiRouteServeHTTP(bhMap backendHandlerMap, w http.ResponseWriter, r *http.Request) {
	h(bhMap, w, r)
}

type MultiRouteAdapter struct {
	Ctx     backendHandlerMap
	Handler MultirouteHandler
}

func (mra *MultiRouteAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mra.Handler.MultiRouteServeHTTP(mra.Ctx, w, r)
}
