package service

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"github.com/xtracdev/xavi/loadbalancer"
)

type backend struct {
	Name         string
	LoadBalancer loadbalancer.LoadBalancer
}

func instantiateLoadBalancer(policyName string, backendName string, servers []config.ServerConfig) (loadbalancer.LoadBalancer, error) {
	factory := loadbalancer.ObtainFactoryForLoadBalancer(policyName)
	if policyName == "" || factory == nil {
		factory = new(loadbalancer.RoundRobinLoadBalancerFactory)
	}

	return factory.NewLoadBalancer(backendName, servers)
}

func buildBackends(kvs kvstore.KVStore, names []string) ([]*backend, error) {
	var backends []*backend
	for _, name := range names {
		be, err := buildBackend(kvs, name)
		if err != nil {
			return nil, err
		}

		backends = append(backends, be)
	}

	return backends, nil
}

func buildBackend(kvs kvstore.KVStore, name string) (*backend, error) {
	var b backend

	log.Info("Building backend " + name)
	backendConfig, err := config.ReadBackendConfig(name, kvs)
	if err != nil {
		return nil, err
	}

	if backendConfig == nil {
		return nil, errors.New("Backend defnition for '" + name + "' not found")
	}

	b.Name = name
	var servers []config.ServerConfig

	for _, serverName := range backendConfig.ServerNames {
		server, err := buildServer(serverName, kvs)
		if err != nil {
			return nil, err
		}
		servers = append(servers, *server)

	}

	loadBalancer, err := instantiateLoadBalancer(backendConfig.LoadBalancerPolicy, name, servers)
	if err != nil {
		return nil, err
	}

	b.LoadBalancer = loadBalancer

	return &b, nil
}

func (b *backend) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Backend: %s\n", b.Name))

	return buffer.String()
}

func (b *backend) getConnectAddress() (string, error) {
	return b.LoadBalancer.GetConnectAddress()
}
