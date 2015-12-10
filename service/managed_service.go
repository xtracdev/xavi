package service

import (
	"bytes"
	"expvar"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/plugin"
	"golang.org/x/net/context"
	"net/http"
	"strings"
)

//Managed service contains the configuration we boot a listener from.
type managedService struct {
	Address      string
	ListenerName string
	Routes       []route
}

//Collect the routes based on URI. A single URI may have multiple routes, but all but one route must
//have a guard condition expressed via MsgProps on the route definition.
func (ms *managedService) organizeRoutesByUri() map[string][]route {
	urimap := make(map[string][]route)
	for _, route := range ms.Routes {
		uriEntry := urimap[route.URIRoot]
		uriEntry = append(uriEntry, route)
		urimap[route.URIRoot] = uriEntry
	}

	return mapsUrisToOrderedRoutes(urimap)
}

//Order the routes for each URI with guarded routes in the front of the slice, and any
//unguarded routes at the end of the slice
func mapsUrisToOrderedRoutes(uriRouteMap map[string][]route) map[string][]route {
	orderedMap := make(map[string][]route)
	for uri, routes := range uriRouteMap {
		orderedRoutes := orderRoutes(routes)
		orderedMap[uri] = orderedRoutes
	}
	return orderedMap
}

//A guard function returns false if the guard condition expressed by MsgProps for a route
//is not satisfied, true otherwise.
type guardFunction func(req *http.Request) (bool, error)

//guardAndHandler is a pair consisting of a guard condition for a uri, and the handler that handles the
//request if the guard condition is satisfied.
type guardAndHandler struct {
	Guard     guardFunction
	HandlerFn plugin.ContextHandlerFunc
}

func makeGHEntryForSingleBackendRoute(r route) guardAndHandler {
	guardFn := makeGuardFunction(r)

	requestHandler := &requestHandler{
		Transport: &http.Transport{DisableKeepAlives: false, DisableCompression: false},
		Backend:   r.Backends[0],
	}

	handlerFn := requestHandler.toContextHandlerFunc()

	handler := plugin.WrapHandlerFunc(handlerFn, r.WrapperFactories)

	ghEntry := guardAndHandler{Guard: guardFn, HandlerFn: handler}

	return ghEntry
}

func makeGHEntryForMultipleBackends(r route) guardAndHandler {
	guardFn := makeGuardFunction(r)

	//Create a backend handler map that will map the backend name to the
	//wrapped handler for the backend
	var handlerMap plugin.BackendHandlerMap = make(plugin.BackendHandlerMap)

	//Lookup the factory for the multiroute handler
	factoryName := r.MultiBackendPluginName
	factory, err := plugin.LookupMultiBackendAdapterFactory(factoryName)
	if err != nil {
		panic("Cannot configure service - no such MultiRoutePluginName: " + factoryName)
	}

	//Go through the backends and build a handler for each
	for _, backend := range r.Backends {
		log.Debug("handler for ", backend.Name)
		requestHandler := &requestHandler{
			Transport: &http.Transport{DisableKeepAlives: false, DisableCompression: false},
			Backend:   backend,
		}

		var handlerFn plugin.ContextHandlerFunc = requestHandler.toContextHandlerFunc()

		handlerMap[backend.Name] = plugin.ContextHandlerFunc(handlerFn)
	}

	//Use the factory to create a wrapped handler that can service the requests
	log.Info("creating handler via factory using", handlerMap)
	multiRouteHandler := factory(handlerMap)

	//Now wrap the handler function with the plugins.
	handler := plugin.WrapHandlerFunc(multiRouteHandler.ToHandlerFunc(), r.WrapperFactories)

	return guardAndHandler{Guard: guardFn, HandlerFn: handler}
}

//Map the routes to a guard and handler pair. The guard function is generated for the URI based on
//the route MsgProps, and the handler is the request handler wrapped by the plugin chaining.
func mapRoutesToGuardAndHandler(uriRouteMap map[string][]route) map[string][]guardAndHandler {
	ghMap := make(map[string][]guardAndHandler)
	for uri, routes := range uriRouteMap {
		for _, r := range routes {
			ghEntries := ghMap[uri]

			var ghEntry guardAndHandler

			if len(r.Backends) == 1 {
				ghEntry = makeGHEntryForSingleBackendRoute(r)
			} else {
				ghEntry = makeGHEntryForMultipleBackends(r)
			}

			ghEntries = append(ghEntries, ghEntry)
			ghMap[uri] = ghEntries
		}
	}

	return ghMap
}

//Make a uri handler map by reducing the guarded URI handlers into a single handler that
//delegates the call to the first matching route guard condition.
func makeURIHandlerMap(ghMap map[string][]guardAndHandler) map[string]plugin.ContextHandler {
	handlerMap := make(map[string]plugin.ContextHandler)
	for uri, guardAndHandlers := range ghMap {
		handlerMap[uri] = reduceHandlers(guardAndHandlers)
	}
	return handlerMap
}

//Reduce handlers creates a single handler function from all the guarded and unguarded handlers
//associated with a route URI.
func reduceHandlers(guardHandlerPairs []guardAndHandler) plugin.ContextHandler {
	log.Debug("reduceHandlers called")
	return plugin.ContextHandlerFunc(func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
		var handled = false
		for _, ghPair := range guardHandlerPairs {
			guardSatisfied, err := ghPair.Guard(req)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError) //Malformed MsgProp
				handled = true
				break
			}
			if guardSatisfied {
				handled = true
				handlerFn := ghPair.HandlerFn
				handlerFn(ctx, rw, req)
				break
			}
		}

		if !handled {
			rw.WriteHeader(http.StatusNotFound)
		}
	})
}

//makeGuardFunction returns a function that checks the guard condition for a route as expressed by
//the route's MsgProps configuration.
func makeGuardFunction(r route) guardFunction {
	if r.MsgProps == "" {
		log.Debug("creating always true guard function function")
		return func(req *http.Request) (bool, error) {
			return true, nil
		}
	}

	headerAndValue := strings.Split(r.MsgProps, "=")
	if len(headerAndValue) != 2 {
		log.Info("unable to process guard condition: ", r.MsgProps)
		return func(req *http.Request) (bool, error) {
			return false, fmt.Errorf("Unable to process guard condition for %s - %s", r.URIRoot, r.MsgProps)
		}
	}

	log.Debug("creating header value comparison guard")
	return func(req *http.Request) (bool, error) {
		log.Debug(fmt.Sprintf("test header %s for val %s", headerAndValue[0], headerAndValue[1]))
		return req.Header.Get(headerAndValue[0]) == headerAndValue[1], nil
	}
}

//Order the routes by putting those with guard conditions in the front of the slice, with optionally a
//unguarded route at the rear of the slice. Note only a single unguarded route for a URI may be configured.
func orderRoutes(routes []route) []route {
	if len(routes) <= 1 {
		return routes
	}

	var guarded, unguarded []route
	for _, r := range routes {
		switch r.MsgProps {
		case "":
			unguarded = append(unguarded, r)
		default:
			guarded = append(guarded, r)
		}
	}

	if len(unguarded) > 1 {
		panic(fmt.Sprintf("Multiple unguarded routes for uri %s", unguarded[0].URIRoot))
	}

	log.Debug("Added ", len(guarded), " routes and ", len(unguarded), " routes for ", routes[0].URIRoot)

	return append(guarded, unguarded...)
}

func (ms *managedService) mapUrisToRoutes() map[string]plugin.ContextHandler {
	log.Debug("Arranging routes by uri and generating handlers")
	uriToRoutesMap := ms.organizeRoutesByUri()
	uriToGuardAndHandlerMap := mapRoutesToGuardAndHandler(uriToRoutesMap)
	uriHandlerMap := makeURIHandlerMap(uriToGuardAndHandlerMap)
	return uriHandlerMap
}

//Run starts up a listener hosting the configuration associated with the managed service instance.
func (ms *managedService) Run() {
	mux := http.NewServeMux()

	uriHandlerMap := ms.mapUrisToRoutes()
	for uri, handler := range uriHandlerMap {
		adapter := &plugin.ContextAdapter{
			Ctx:     context.Background(),
			Handler: handler,
		}
		mux.Handle(uri, adapter)
	}

	//Expvar handler
	mux.HandleFunc("/debug/vars", expvarHandler)

	server := &http.Server{Handler: mux, Addr: ms.Address}

	if err := server.ListenAndServe(); err != nil {
		msg := fmt.Sprintf("Starting service for listener %s failed: %v", ms.ListenerName, err)
		log.Error(msg)
	}

}

//AddRoute adds a route to the managed service
func (ms *managedService) AddRoute(route *route) {
	ms.Routes = append(ms.Routes, *route)
}

//String provides a string representation of the configuration associated with the managed service.
func (ms *managedService) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Managed service %s at address %s\n", ms.ListenerName, ms.Address))
	for _, r := range ms.Routes {
		buffer.WriteString(fmt.Sprintf("%s\n", r))
	}

	return buffer.String()
}

//expvar exports on the default service mux, which we are not using here. So the following
//code from expvar.go has been lifter so we can add the expvar GET
func expvarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}
