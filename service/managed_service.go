package service

import (
	"bytes"
	"expvar"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/plugin"
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
func (ms *managedService) mapUrisToRoutes() map[string][]route {
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

//A guard function returns false if the guard condition expresssed by MsgProps for a route
//is not satisfied, true otherwise.
type guardFunction func(req *http.Request) (bool, error)

//guardAndHandler is a pair consisting of a guard condition for a uri, and the handler that handles the
//request if the guard condition is satisfied.
type guardAndHandler struct {
	Guard     guardFunction
	HandlerFn http.HandlerFunc
}

//Map the routes to a guard and handler pair. The guard function is generated for the URI based on
//the route MsgProps, and the handler is the request handler wrapped by the plugin chaing.
func mapRoutesToGuardAndHandler(uriRouteMap map[string][]route) map[string][]guardAndHandler {
	ghMap := make(map[string][]guardAndHandler)
	for uri, routes := range uriRouteMap {
		for _, r := range routes {
			ghEntries := ghMap[uri]
			guardFn := makeGuardFunction(r)

			requestHandler := &requestHandler{
				Transport: &http.Transport{DisableKeepAlives: false, DisableCompression: false},
				Backend:   r.Backends[0],
			}

			handlerFn := requestHandler.toHandlerFunc()

			handler := plugin.WrapHandlerFunc(handlerFn, r.WrapperFactories)

			ghEntry := guardAndHandler{Guard: guardFn, HandlerFn: handler}

			ghEntries = append(ghEntries, ghEntry)
			ghMap[uri] = ghEntries
		}
	}

	return ghMap
}

//Make a uri handler map by reducing the gaurded URI handlers into a single handler that
//delegates the call to the first matching route guard condition.
func makeURIHandlerMap(ghMap map[string][]guardAndHandler) map[string]http.Handler {
	handlerMap := make(map[string]http.Handler)
	for uri, guardAndHandlers := range ghMap {
		handlerMap[uri] = reduceHandlers(guardAndHandlers)
	}
	return handlerMap
}

//Reduce handlers creates a single handler function from all the guarded and unguarded handlers
//associated with a route URI.
func reduceHandlers(guardHandlerPairs []guardAndHandler) http.Handler {
	log.Debug("reduceHandlers called")
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
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
				handlerFn(rw, req)
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

	log.Debug("creating header value comparison gaurd")
	return func(req *http.Request) (bool, error) {
		log.Debug(fmt.Sprintf("test header %s for val %s", headerAndValue[0], headerAndValue[1]))
		return req.Header.Get(headerAndValue[0]) == headerAndValue[1], nil
	}
}

//Order the routes by putting those with guard conditions in the front of the slice, with optionally a
//ungaurded route at the rear of the slice. Note only a single unguarded route for a URI may be configured.
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

//Run starts up a listener hosting the configuration assocaited with the managed service instance.
func (ms *managedService) Run() {
	mux := http.NewServeMux()

	log.Debug("Arranging routes by uri and generating handlers")
	uriToRoutesMap := ms.mapUrisToRoutes()
	uriToGuardAndHandlerMap := mapRoutesToGuardAndHandler(uriToRoutesMap)
	uriHandlerMap := makeURIHandlerMap(uriToGuardAndHandlerMap)

	for uri, handler := range uriHandlerMap {
		mux.Handle(uri, handler)
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
