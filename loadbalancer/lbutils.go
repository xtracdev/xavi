package loadbalancer

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/config"
)

var ErrBackendNotFound = errors.New("Given backed end not found in active listener config")

func findBackend(backend string) (*config.ServiceBackend, error) {
	for _, listenerName := range config.ActiveListenerNames() {
		sc := config.ActiveConfigForListener(listenerName)

		for _, r := range sc.Routes {
			log.Infof("route config for %s:", r.Route.Name)
			for _, b := range r.Backends {
				if b.Backend.Name == backend {
					return b, nil
				}
			}
		}
	}

	return nil, ErrBackendNotFound
}

func serversForBackend(backend *config.ServiceBackend) []config.ServerConfig {
	servers := make([]config.ServerConfig, 0)

	if backend == nil {
		return servers
	}

	for _, s := range backend.Servers {
		log.Infof("server config for %s:", s.Name)
		servers = append(servers, *s)
	}

	return servers
}

// NewLoadBalancer instantiates a load balancer based on the named backend configuration. Backend
// names are scoped to routes, thus the route is given to ensure the correct backend is returned
// if multiple backend definitions with the same name are given.
func NewLoadBalancer(backendName string) (LoadBalancer, *config.BackendConfig, error) {
	backend, err := findBackend(backendName)
	if err != nil {
		return nil, nil, err
	}

	servers := serversForBackend(backend)

	backendConfig := backend.Backend
	factory := ObtainFactoryForLoadBalancer(backendConfig.LoadBalancerPolicy)
	if factory == nil {
		factory = new(RoundRobinLoadBalancerFactory)
	}

	lb, err := factory.NewLoadBalancer(backendConfig.Name, backendConfig.CACertPath, servers)

	return lb, backendConfig, err
}
