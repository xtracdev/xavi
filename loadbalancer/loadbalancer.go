package loadbalancer

import (
	"sync"

	"github.com/xtracdev/xavi/config"
	"github.com/armon/go-metrics"
)

//LoadBalancerEndpoint contains the information about an endpoint needed by the load balancer
//for handing connections to a consumer
type LoadBalancerEndpoint struct {
	Address    string
	PingURI    string
	Up         bool
	CACertPath string
	mu         sync.RWMutex
}

//IsUp reads the status of the endpoint. The function is safe for simultaneous use by multiple goroutines.
func (lb *LoadBalancerEndpoint) IsUp() bool {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.Up
}

//MarkLoadBalancerEndpointUp sets the status for this endpoind. The function is safe for simultaneous use by multiple goroutines.
func (lb *LoadBalancerEndpoint) MarkLoadBalancerEndpointUp(isUp bool) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.Up = isUp
	if isUp {
		metrics.SetGauge([]string{"endpoint",lb.Address},1.0)
	} else {
		metrics.SetGauge([]string{"endpoint",lb.Address},0.0)
	}
}

//LoadBalancer has methods for handing out connection addressed and marking
//endpoints up or down.
type LoadBalancer interface {
	GetConnectAddress() (string, error)
	MarkEndpointDown(string) error
	MarkEndpointUp(string) error
	GetEndpoints() (healthy []string, unhealthy []string)
}

//LoadBalancerFactory defines an interface for instantiating load balancers.
type LoadBalancerFactory interface {
	NewLoadBalancer(name, caCertPath string, servers []config.ServerConfig) (LoadBalancer, error)
}
