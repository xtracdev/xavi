package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"io/ioutil"
	"net/http"
)

const (
	routesURI = "/v1/routes/"
)

//Errors
var (
	errRouteNotFound        = errors.New("Route definition not found")
	errRouteResourceMissing = errors.New("Route resource not present in url - expected /v1/routes/route-resource")
)

//RouteDefCmd is the RouteDef instance used to expose as an API endpoint.
var RouteDefCmd RouteDef

//RouteDef is used to hang the ApiCommand functions needed for exposing route def capabilities
//via a REST API
type RouteDef struct{}

//GetURIRoot returns the URI root used to serve route def API calls
func (RouteDef) GetURIRoot() string {
	return routesURI
}

//PutDefinition creates or updates a route definition
func (RouteDef) PutDefinition(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	body, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	log.Info(fmt.Sprintf("Put request with payload %s", string(body)))
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return nil, nil
	}

	routeName := resourceIDFromURI(req.RequestURI)
	log.Info(routeName)
	if routeName == "" {
		resp.WriteHeader(http.StatusNotFound)
		return nil, errServiceResourceMissing
	}

	routeConfig := new(config.RouteConfig)
	err = json.Unmarshal(body, routeConfig)
	if err != nil {
		log.Warn("Error unmarshaling request body")
		resp.WriteHeader(http.StatusBadRequest)
		return nil, err
	}

	routeConfig.Name = routeName
	err = routeConfig.Store(kvs)
	if err != nil {
		log.Warn("Error persisting route definition")
		resp.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	err = kvs.Flush()
	if err != nil {
		return nil, err
	}

	return nil, nil
}

//GetDefinitionList returns a list of route definitions
func (RouteDef) GetDefinitionList(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	log.Info("route GetDefinitionList service called")
	routes, err := config.ListRouteConfigs(kvs)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	if routes == nil {
		routes = make([]*config.RouteConfig, 0)
	}

	return routes, nil

}

//GetDefinition retrieves a specific route definition
func (RouteDef) GetDefinition(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	routeName := resourceIDFromURI(req.RequestURI)

	routeConfig, err := config.ReadRouteConfig(routeName, kvs)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	if routeConfig == nil {
		resp.WriteHeader(http.StatusNotFound)
		return nil, errRouteNotFound
	}

	return routeConfig, err

}

//DoPost handles post requests, which are not allowed for RouteDef.
func (RouteDef) DoPost(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	resp.WriteHeader(http.StatusMethodNotAllowed)
	return nil, nil
}
