package config

import (
	"errors"
	"github.com/xtracdev/xavi/kvstore"
)

type ServiceConfig struct {
	Listener *ListenerConfig
	Routes   []*ServiceRoute
}

type ServiceRoute struct {
	Route    *RouteConfig
	Backends []*ServiceBackend
}

type ServiceBackend struct {
	Backend *BackendConfig
	Servers []*ServerConfig
}

var (
	ErrNoListenerName = errors.New("No listener name specified")
	ErrNoKVStore      = errors.New("No kv store provided")
)

func ReadServiceConfig(listenerName string, kvs kvstore.KVStore) (*ServiceConfig, error) {
	if listenerName == "" {
		return nil, ErrNoListenerName
	}

	if kvs == nil {
		return nil, ErrNoKVStore
	}
	sc := new(ServiceConfig)

	err := readStartingWithListener(sc, listenerName, kvs)
	if err != nil {
		return nil, err
	}

	return sc, nil
}

func readStartingWithListener(sc *ServiceConfig, listenerName string, kvs kvstore.KVStore) error {

	//Read the listener def from the kv store
	lc, err := ReadListenerConfig(listenerName, kvs)
	if err != nil {
		return err
	}

	if lc == nil {
		return errors.New("Listener config '" + listenerName + "' not found")
	}

	sc.Listener = lc

	//Iterate through the routes and populate the slice of route configs
	for _, routeName := range lc.RouteNames {
		route, err := readRouteForListener(sc, routeName, kvs)
		if err != nil {
			return err
		}

		sc.Routes = append(sc.Routes, route)
	}

	return nil
}

func readRouteForListener(sc *ServiceConfig, routeName string, kvs kvstore.KVStore) (*ServiceRoute, error) {
	routeConfig, err := ReadRouteConfig(routeName, kvs)
	if err != nil {
		return nil, err
	}

	if routeConfig == nil {
		return nil, errors.New("Route config '" + routeName + "' not found")
	}

	sr := new(ServiceRoute)
	sr.Route = routeConfig

	//Iterate through the backends associated with the route
	for _, backendName := range routeConfig.Backends {
		backend, err := readBackendForRoute(sr, backendName, kvs)
		if err != nil {
			return nil, err
		}

		sr.Backends = append(sr.Backends, backend)
	}

	return sr, nil
}

func readBackendForRoute(sr *ServiceRoute, backendName string, kvs kvstore.KVStore) (*ServiceBackend, error) {
	backendConfig, err := ReadBackendConfig(backendName, kvs)
	if err != nil {
		return nil, err
	}

	if backendConfig == nil {
		return nil, errors.New("Backend defnition for '" + backendName + "' not found")
	}

	be := new(ServiceBackend)
	be.Backend = backendConfig

	for _, serverName := range backendConfig.ServerNames {
		server, err := readServerForBackend(be, serverName, kvs)
		if err != nil {
			return nil, err
		}

		be.Servers = append(be.Servers, server)
	}

	return be, nil
}

func readServerForBackend(be *ServiceBackend, serverName string, kvs kvstore.KVStore) (*ServerConfig, error) {
	serverConfig, err := ReadServerConfig(serverName, kvs)
	if err != nil {
		return nil, err
	}

	if serverConfig == nil {
		return nil, errors.New("No definition for server '" + serverName + "' found")
	}

	return serverConfig, err
}
