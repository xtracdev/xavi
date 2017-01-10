package service

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
)

//TODO - rewrite this module to build and use a ServiceConfig instance instead of a listener name

func readListenerConfig(name string, kvs kvstore.KVStore) (lc *config.ListenerConfig, err error) {
	lc, err = config.ReadListenerConfig(name, kvs)

	if lc == nil {
		err = errors.New("Listener config '" + name + "' not found")
	}
	return
}

//BuildServiceForListener builds a runnable service based on the given name, retrieving
//definitions using the supplied KVStore and listening on the supplied address.
func BuildServiceForListener(name string, address string, kvs kvstore.KVStore) (Service, error) {
	var managedService = &managedService{
		ListenerName: name,
		Address:      address,
	}

	log.Info("Building service for listener " + name)
	listenerConfig, err := readListenerConfig(name, kvs)
	if err != nil {
		log.Info("Listener definition not found")
		return nil, err
	}

	log.Info("reading routes...")
	for _, routeName := range listenerConfig.RouteNames {
		log.Info("route " + routeName + "...")
		route, err := buildRoute(routeName, kvs)
		if err != nil {
			return nil, err
		}
		managedService.AddRoute(route)
		if err != nil {
			return nil, err
		}
	}

	return managedService, nil
}

func BuildHealthContextForListener(name string, kvs kvstore.KVStore) (*HealthCheckContext, error) {
	log.Info("Building health check context for listener " + name)

	hcc := &HealthCheckContext{
		ListenerName:         name,
		EnableHealthEndpoint: true,
	}

	listenerConfig, err := readListenerConfig(name, kvs)
	if err != nil {
		log.Info("Listener definition not found")
		return nil, err
	}

	hcc.EnableHealthEndpoint = listenerConfig.HealthEndpoint
	log.Info("reading routes...")
	for _, routeName := range listenerConfig.RouteNames {
		log.Info("route " + routeName + "...")
		route, err := buildRoute(routeName, kvs)
		if err != nil {
			return nil, err
		}
		hcc.AddRouteContext(route)
	}

	return hcc, nil
}
