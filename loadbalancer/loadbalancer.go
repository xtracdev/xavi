package loadbalancer

import (
	"github.com/xtracdev/xavi/config"
)

//LoadBalancerEndpoint contains the information about an endpoint needed by the load balancer
//for handing connections to a consumer
type LoadBalancerEndpoint struct {
	Address string
	PingURI string
	Up      bool
}

//LoadBalancer has methods for handing out connection addressed and marking
//endpoints up or down.
type LoadBalancer interface {
	GetConnectAddress() (string, error)
	MarkEndpointDown(string) error
	MarkEndpointUp(string) error
}

//LoadBalancerFactory defines an interface for instantiating load balancers.
type LoadBalancerFactory interface {
	NewLoadBalancer(name string, servers []config.ServerConfig) (LoadBalancer, error)
}
