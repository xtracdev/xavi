package service

import (
	"errors"
	log "github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
)

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
	var managedService managedService

	log.Info("Building service for listener " + name)
	listenerConfig, err := readListenerConfig(name, kvs)
	if err != nil {
		log.Info("Listener definition not found")
		return nil, err
	}

	managedService.ListenerName = name
	managedService.Address = address
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

	return &managedService, nil
}
