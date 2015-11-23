package service

import (
	"bytes"
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"github.com/xtracdev/xavi/plugin"
)

type route struct {
	Name                 string
	URIRoot              string
	Backends             []*backend
	WrapperFactories     []plugin.WrapperFactory
	MsgProps             string
	MultiRoutePluginName string
}

func makeRouteNotFoundError(name string) error {
	return errors.New("Route '" + name + "' not found")
}

func buildRoute(name string, kvs kvstore.KVStore) (*route, error) {
	var r route

	r.Name = name

	routeConfig, err := config.ReadRouteConfig(name, kvs)
	if err != nil {
		return nil, err
	}

	if routeConfig == nil {
		return nil, makeRouteNotFoundError(name)
	}

	backends, err := buildBackends(kvs, routeConfig.Backends)
	if err != nil {
		return nil, err
	}

	r.URIRoot = routeConfig.URIRoot
	r.Backends = backends

	if len(r.Backends) > 1 && r.MultiRoutePluginName == "" {
		return nil, errors.New("MultiRoute plugin name must be provided when multiple backends are configured")
	}

	for _, pluginName := range routeConfig.Plugins {
		factory, err := plugin.LookupWrapperFactory(pluginName)
		if err != nil {
			return nil, fmt.Errorf("No wrapper factory with name %s in registry", pluginName)
		}

		log.Debug("adding wrapper factory to factories")
		r.WrapperFactories = append(r.WrapperFactories, factory)

	}

	r.MsgProps = routeConfig.MsgProps

	return &r, nil
}

func (r route) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Route: %s\n", r.Name))
	buffer.WriteString(fmt.Sprintf("\tUri root: %s\n", r.URIRoot))
	buffer.WriteString(fmt.Sprintf("\tBackends:\n"))
	for _, be := range r.Backends {
		buffer.WriteString(fmt.Sprintf("\tBackend: %s\n", be))
	}
	return buffer.String()
}
