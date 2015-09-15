package agent

import (
	"encoding/json"
	"github.com/xtracdev/xavi/kvstore"
	"net/http"
	"strings"
)

// Agent specifies the address, handlers, and KVS used to expose commands
// via a REST interface
type Agent struct {
	address  string
	handlers map[string]func(http.ResponseWriter, *http.Request)
	kvstore  kvstore.KVStore
}

//NewAgent created a new Agent instance
func NewAgent(address string, kvs kvstore.KVStore) *Agent {
	handlersMap := make(map[string]func(http.ResponseWriter, *http.Request))
	newAgent := &Agent{address, handlersMap, kvs}
	newAgent.registerHandlers()
	return newAgent
}

func (a *Agent) addHandler(uri string, handler func(http.ResponseWriter, *http.Request)) {
	a.handlers[uri] = handler
}

func (a *Agent) registerHandlers() {
	//Each rest endpoint needs to be registered here.

	serverAPIService := NewAPIService(ServerDefCmd)
	a.addHandler(serversURI, wrap(a.kvstore, serverAPIService))

	backendAPIService := NewAPIService(BackendDefCmd)
	a.addHandler(backendsURI, wrap(a.kvstore, backendAPIService))

	routeAPIService := NewAPIService(RouteDefCmd)
	a.addHandler(routesURI, wrap(a.kvstore, routeAPIService))

	listenerAPIService := NewAPIService(ListenerDefCmd)
	a.addHandler(listenersURI, wrap(a.kvstore, listenerAPIService))

	spawnAPIService := NewAPIService(SpawnListenerDefCmd)
	a.addHandler(spawnURI, wrap(a.kvstore, spawnAPIService))

	spawnKillApiService := NewAPIService(SpawnKillerDefCmd)
	a.addHandler(spawnKillURI, wrap(a.kvstore, spawnKillApiService))
}

func wrap(kvs kvstore.KVStore, apiService *APIService) func(resp http.ResponseWriter, req *http.Request) {
	f := func(resp http.ResponseWriter, req *http.Request) {
	HAS_ERR:
		obj, err := apiService.serversEndpoint(kvs, resp, req)
		if err != nil {
			//Assume handler function has written error header
			resp.Write([]byte(err.Error()))
			return
		}

		if obj != nil {
			var buf []byte
			buf, err = json.Marshal(obj)
			if err != nil {
				goto HAS_ERR
			}
			resp.Header().Set("Content-Type", "application/json")
			resp.Write(buf)
		}
	}
	return f
}

//Start the agent listening on the address it was constructed with
func (a *Agent) Start() {
	for uri, handler := range a.handlers {
		http.HandleFunc(uri, handler)
	}

	http.ListenAndServe(a.address, nil)
}

//Parse the resource id as the last element in the URI
func resourceIDFromURI(uri string) string {
	uriComponents := strings.Split(uri, "/")
	return uriComponents[len(uriComponents)-1]
}
