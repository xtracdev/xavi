package config

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/kvstore"
)

//ServiceConfig is the top level structure that joins up all the configuration
//from the listener def into a single structure, reading the definition from the
//KVStore
type ServiceConfig struct {
	Listener *ListenerConfig
	Routes   []*ServiceRoute
}

//ServiceRoute contains the Route definition read from config and the associated
//backend configs as well
type ServiceRoute struct {
	Route    *RouteConfig
	Backends []*ServiceBackend
}

//ServiceBackend contains the backend definition and all the linked server
//definitions for the backend
type ServiceBackend struct {
	Backend *BackendConfig
	Servers []*ServerConfig
}

var (
	ErrNoListenerName = errors.New("No listener name specified")
	ErrNoKVStore      = errors.New("No kv store provided")
)

//ReadServiceConfig reads all configuration for a given listener and links all the definitions
//together
func ReadServiceConfig(listenerName string, kvs kvstore.KVStore) (*ServiceConfig, error) {
	log.Infof("ReadServiceConfig: Reading service configuration for listener %s", listenerName)

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
	log.Info("ReadServiceConfig: Read listener configuration")

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
	log.Infof("ReadServiceConfig: Read route config for %s", routeName)
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
	log.Infof("ReadServiceConfig: Read backend config for %s", backendName)

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
	log.Infof("ReadServiceConfig: Read server config for %s", serverName)

	serverConfig, err := ReadServerConfig(serverName, kvs)
	if err != nil {
		return nil, err
	}

	if serverConfig == nil {
		return nil, errors.New("No definition for server '" + serverName + "' found")
	}

	return serverConfig, err
}

//LogConfig logs information associated with the ServiceConfig
func (sc *ServiceConfig) LogConfig() {
	log.Infof("Logging service config for listener %s:", sc.Listener.Name)
	log.Infof("%v", *sc.Listener)
	for _,r := range sc.Routes {
		log.Infof("route config for %s:", r.Route.Name)
		log.Infof("%v", *r.Route)
		for _, b := range r.Backends {
			log.Infof("backend config for %s:", b.Backend.Name)
			log.Infof("%v", *b.Backend)
			for _, s := range b.Servers {
				log.Infof("server config for %s:", s.Name)
				log.Infof("%v", *s)
			}
		}
	}
}
