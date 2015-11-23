package config

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/kvstore"
)

//RouteConfig represents a route configuration definition
type RouteConfig struct {
	Name               string
	URIRoot            string
	Backends           []string
	Plugins            []string
	MultiBackendPlugin string
	MsgProps           string
}

//JSONToRoute unmarshals the JSON representation of a route definition
func JSONToRoute(bytes []byte) *RouteConfig {
	var r *RouteConfig
	if bytes == nil {
		return r
	}

	r = new(RouteConfig)
	json.Unmarshal(bytes, r)
	return r
}

//Store persists the route definition using the supplied KVS
func (routeConfig *RouteConfig) Store(kvs kvstore.KVStore) error {
	b, err := json.Marshal(routeConfig)
	if err != nil {
		return nil
	}

	key := fmt.Sprintf("routes/%s", routeConfig.Name)
	log.Info(fmt.Sprintf("adding %s under key %s", string(b), key))
	err = kvs.Put(key, b)
	if err != nil {
		return err
	}

	return nil
}

//ReadRouteConfig retrieves the specified route config using the supplied key value store
func ReadRouteConfig(name string, kvs kvstore.KVStore) (*RouteConfig, error) {
	//Read the definition from the key store
	bv, err := readKey("routes/"+name, kvs)
	if err != nil {
		return nil, err
	}

	return JSONToRoute(bv), nil
}

//ListRouteConfigs returns the route configs in the key value store
func ListRouteConfigs(kvs kvstore.KVStore) ([]*RouteConfig, error) {
	pairs, err := kvs.List("routes/")
	if err != nil {
		return nil, err
	}

	var routes []*RouteConfig
	for _, p := range pairs {
		routes = append(routes, JSONToRoute(p.Value))
	}

	return routes, nil
}
