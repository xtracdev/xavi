package loadbalancer

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"testing"
)

func TestNilServersSlice(t *testing.T) {
	var roundRobinFactory LoadBalancerFactory = new(RoundRobinLoadBalancerFactory)
	_, err := roundRobinFactory.NewLoadBalancer("backend", "", nil)
	assert.NotNil(t, err)
}

func TestEmptyBackendName(t *testing.T) {
	serverConfig := config.ServerConfig{
		Name:    "server1",
		Address: "server1.domain.com",
		Port:    11000,
		PingURI: "/xtracrulesok",
	}

	servers := []config.ServerConfig{serverConfig}

	var roundRobinFactory LoadBalancerFactory = new(RoundRobinLoadBalancerFactory)
	_, err := roundRobinFactory.NewLoadBalancer("", "", servers)
	assert.NotNil(t, err)
}

func TestEmptyServersSlice(t *testing.T) {
	var roundRobinFactory LoadBalancerFactory = new(RoundRobinLoadBalancerFactory)
	var servers []config.ServerConfig
	_, err := roundRobinFactory.NewLoadBalancer("backend", "", servers)
	assert.NotNil(t, err)
}

func TestSingleServerConfig(t *testing.T) {
	serverConfig := config.ServerConfig{
		Name:    "server1",
		Address: "server1.domain.com",
		Port:    11000,
		PingURI: "/xtracrulesok",
	}

	servers := []config.ServerConfig{serverConfig}

	var roundRobinFactory LoadBalancerFactory = new(RoundRobinLoadBalancerFactory)
	rr, err := roundRobinFactory.NewLoadBalancer("backend", "", servers)
	assert.NotNil(t, rr)
	assert.Nil(t, err)

	address := fmt.Sprintf("%s:%d", serverConfig.Address, serverConfig.Port)
	for i := 0; i < 5; i++ {
		addr, err := rr.GetConnectAddress()
		assert.Nil(t, err)
		assert.Equal(t, address, addr)
	}

	h, u := rr.GetEndpoints()
	assert.Equal(t, 0, len(u))
	assert.Equal(t, 1, len(h))
	assert.Equal(t, address, h[0])
}

func TestMultiServerConfig(t *testing.T) {
	serverConfig := config.ServerConfig{
		Name:    "server1",
		Address: "server1.domain.com",
		Port:    11000,
		PingURI: "/xtracrulesok",
	}

	serverConfig2 := config.ServerConfig{
		Name:    "server2",
		Address: "server2.domain.com",
		Port:    11000,
		PingURI: "/xtracrulesok",
	}

	servers := []config.ServerConfig{serverConfig, serverConfig2}

	var roundRobinFactory LoadBalancerFactory = new(RoundRobinLoadBalancerFactory)
	rr, err := roundRobinFactory.NewLoadBalancer("backend", "", servers)
	assert.NotNil(t, rr)
	assert.Nil(t, err)

	for i := 0; i < 5; i++ {
		addr, err := rr.GetConnectAddress()
		assert.Nil(t, err)
		assert.Equal(t, fmt.Sprintf("%s:%d", serverConfig.Address, serverConfig.Port), addr)

		addr, err = rr.GetConnectAddress()
		assert.Nil(t, err)
		assert.Equal(t, fmt.Sprintf("%s:%d", serverConfig2.Address, serverConfig2.Port), addr)
	}
}

func TestMarkEndpointDown(t *testing.T) {
	serverConfig := config.ServerConfig{
		Name:    "server1",
		Address: "server1.domain.com",
		Port:    11000,
		PingURI: "/xtracrulesok",
	}

	servers := []config.ServerConfig{serverConfig}

	var roundRobinFactory LoadBalancerFactory = new(RoundRobinLoadBalancerFactory)
	rr, err := roundRobinFactory.NewLoadBalancer("backend", "", servers)
	assert.NotNil(t, rr)
	assert.Nil(t, err)

	err = rr.MarkEndpointDown("")
	assert.NotNil(t, err)

	err = rr.MarkEndpointDown("no host port combo here")
	assert.NotNil(t, err)

	err = rr.MarkEndpointDown("notmyserver:123")
	assert.NotNil(t, err)

	err = rr.MarkEndpointDown("server1.domain.com:11000")
	assert.Nil(t, err)

	h, u := rr.GetEndpoints()
	assert.Equal(t, 0, len(h))
	assert.Equal(t, 1, len(u))

	_, err = rr.GetConnectAddress()
	assert.NotNil(t, err)

	err = rr.MarkEndpointUp("server1.domain.com:11000")
	assert.Nil(t, err)

	_, err = rr.GetConnectAddress()
	assert.Nil(t, err)

}
