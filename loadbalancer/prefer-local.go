package loadbalancer

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/config"
	"os"
	"strings"
)

//PreferLocalLoadBalancer is a load balancer that looks to route traffic locally before sending
//traffic remote. Local is defined as servers that have the same hostname as that the load
//balancer is deployed on. The load balancer keeps two server pools - a local pool and a remote
//pool. Local traffic is used and distributed via a round robin load balancer when a local server
//is available, other wise remote servers are used.
type PreferLocalLoadBalancer struct {
	BackendName   string
	LocalServers  LoadBalancer
	RemoteServers LoadBalancer
}

//PreferLocalLoadBalancerFactory is used to instantiate PreferLocalLoadBalancer instances.
type PreferLocalLoadBalancerFactory struct{}

func splitHostFromAddress(server string) string {
	return strings.Split(server, ".")[0]
}

func sameServer(hostname string, servername string) bool {
	lcHost := strings.ToLower(hostname)
	lcServer := strings.ToLower(servername)

	lendiff := len(lcHost) - len(lcServer)

	if lendiff == 0 {
		return lcHost == lcServer
	} else if lendiff > 0 {
		return strings.Index(lcHost, lcServer) == 0 && lcHost[len(lcServer)] == '.'
	} else {
		return strings.Index(lcServer, lcHost) == 0 && lcServer[len(lcHost)] == '.'
	}
}

//isLocal is a predicate that indicates in the given server is local or not, based on matching the
//hostname returned by os.Hostname. Note that we consider localhost as a special case, treating it as
//local.
func isLocal(server string) (bool, error) {

	if server == "localhost" {
		return true, nil
	}

	//For Virtual App Handle environments, os.Hostname will not give the virtual host
	//name, which is what we need to treat as the hostname. In these environments, we
	//treat the value of the APPHANDLE environment variable as the hostname
	appHandle := os.Getenv("APPHANDLE")
	if appHandle != "" {
		return sameServer(appHandle, server), nil
	}

	host, err := os.Hostname()
	if err != nil {
		return false, err
	}

	return sameServer(host, server), nil
}

//partitionServers partitions a slice of servers into a slice of local servers and a slice of
//remote servers
func partitionServers(servers []config.ServerConfig) ([]config.ServerConfig, []config.ServerConfig, error) {
	var localServers, remoteServers []config.ServerConfig
	for _, s := range servers {
		local, err := isLocal(s.Address)
		if err != nil {
			return nil, nil, err
		}

		if local {
			localServers = append(localServers, s)
		} else {
			remoteServers = append(remoteServers, s)
		}
	}

	return localServers, remoteServers, nil
}

//NewLoadBalancer creates an instance of PreferLocalLoadBalancer
func (pl *PreferLocalLoadBalancerFactory) NewLoadBalancer(backendName, caCertPath string, servers []config.ServerConfig) (LoadBalancer, error) {

	if backendName == "" {
		return nil, fmt.Errorf("Expected non-empty backend name")
	}

	if len(servers) == 0 {
		return nil, fmt.Errorf("Expected at least one server in servers argument")
	}

	log.Info("Creating prefer-local load balancer for backend ", backendName, " with ", len(servers), " servers")

	var preferLocalLB PreferLocalLoadBalancer
	var roundRobinFactory LoadBalancerFactory = new(RoundRobinLoadBalancerFactory)

	localServers, remoteServers, err := partitionServers(servers)
	if err != nil {
		return nil, err
	}

	preferLocalLB.BackendName = backendName

	if len(localServers) > 0 {
		localLB, err := roundRobinFactory.NewLoadBalancer(backendName, caCertPath, localServers)
		if err != nil {
			return nil, err
		}

		preferLocalLB.LocalServers = localLB
	} else {
		log.Warn("No local servers specified in prefer-local configuration")
	}

	if len(remoteServers) > 0 {
		remoteLB, err := roundRobinFactory.NewLoadBalancer(backendName, caCertPath, remoteServers)
		if err != nil {
			return nil, err
		}

		preferLocalLB.RemoteServers = remoteLB
	} else {
		log.Warn("No remote servers specified in prefer-local configuration")
	}

	return &preferLocalLB, nil
}

//getConnectAddress returns a connection address from the given load balancer via a call
//to the load balancer's GetConnectAddress method if it is not nil, otherwise returning
//an error
func getConnectAddress(poolname string, lb LoadBalancer) (string, error) {
	if lb == nil {
		return "", fmt.Errorf("No servers in %s pool configuration", poolname)
	}

	return lb.GetConnectAddress()
}

//markEndpointDown marks the endpoint down for the given load balancer via a call
//to the load balancer's MarkEndpointDown method if it is not nil, otherwise returning
//an error
func markEndpointDown(poolname string, endpoint string, lb LoadBalancer) error {
	if lb == nil {
		return fmt.Errorf("No servers in %s pool configuration", poolname)
	}

	return lb.MarkEndpointDown(endpoint)
}

//markEndpointUp marks the endpoint down for the given load balancer via a call
//to the load balancer's MarkEndpointUp method if it is not nil, otherwise returning
//an error
func markEndpointUp(poolname string, endpoint string, lb LoadBalancer) error {
	if lb == nil {
		return fmt.Errorf("No servers in %s pool configuration", poolname)
	}

	return lb.MarkEndpointUp(endpoint)
}

//GetConnectAddress returns the connect address for the PreferLocalLoadBalancer instance
func (pl *PreferLocalLoadBalancer) GetConnectAddress() (string, error) {
	address, err := getConnectAddress("local server", pl.LocalServers)
	switch err {
	case nil:
		return address, err
	default:
		log.Warn(fmt.Sprintf("No local address found: %s. Will look for remote address.", err.Error()))
		return getConnectAddress("remote server", pl.RemoteServers)
	}
}

//MarkEndpointDown marks the given endpoint down for the PreferLocalLoadBalancer instance
func (pl *PreferLocalLoadBalancer) MarkEndpointDown(endpoint string) error {
	err := markEndpointDown("local server", endpoint, pl.LocalServers)
	switch err {
	case nil:
		return nil
	default:
		return markEndpointDown("remote server", endpoint, pl.RemoteServers)
	}
}

//MarkEndpointUp marks the given endpoint down for the PreferLocalLoadBalancer instance
func (pl *PreferLocalLoadBalancer) MarkEndpointUp(endpoint string) error {
	err := markEndpointUp("local server", endpoint, pl.LocalServers)
	switch err {
	case nil:
		return nil
	default:
		return markEndpointUp("remote server", endpoint, pl.RemoteServers)
	}
}
