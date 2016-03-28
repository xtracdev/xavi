package service

import (
	"bytes"
	"crypto/x509"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"github.com/xtracdev/xavi/loadbalancer"
	"io/ioutil"
)

type backend struct {
	Name         string
	LoadBalancer loadbalancer.LoadBalancer
	TLSOnly      bool
	CACert       *x509.CertPool
}

var ErrCACertFile = errors.New("CACert file contained no certificates")

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

	b.TLSOnly = backendConfig.TLSOnly

	b.CACert, err = createCertPool(backendConfig)
	if err != nil {
		return nil, err
	}

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

func createCertPool(backendConfig *config.BackendConfig) (*x509.CertPool, error) {
	if backendConfig.CACertPath == "" {
		return nil, nil
	}

	log.Debug("Creating cert pool for backend ", backendConfig.Name)

	pool := x509.NewCertPool()

	pemData, err := ioutil.ReadFile(backendConfig.CACertPath)
	if err != nil {
		return nil, err
	}

	ok := pool.AppendCertsFromPEM(pemData)
	if !ok {
		return nil, ErrCACertFile
	}

	return pool, nil
}
