package config

import (
	"errors"
	"github.com/xtracdev/xavi/kvstore"
)

type ServiceConfig struct {
	Listener *ListenerConfig
	Routes   []ServiceRoute
}

type ServiceRoute struct {
	Route    *RouteConfig
	Backends []ServiceBackend
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
	return sc, nil
}
