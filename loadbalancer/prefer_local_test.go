package loadbalancer

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"os"
	"testing"
)

const testRemoteServerAddress = "server2.domain.com"
const testPort = 11000

func joinHostAndPort(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

func TestSplitHostFromAddress(t *testing.T) {
	host := splitHostFromAddress(".")
	assert.Equal(t, "", host)

	host = splitHostFromAddress("foo")
	assert.Equal(t, "foo", host)

	host = splitHostFromAddress("foo.domain.com")
	assert.Equal(t, "foo", host)
}

func TestIsLocal(t *testing.T) {
	//Try the host the test runs on
	host, err := os.Hostname()
	assert.Nil(t, err)
	local, err := isLocal(host)
	assert.Nil(t, err)
	assert.True(t, local)

	//Try a host the machine does not run on...
	host = testRemoteServerAddress
	local, err = isLocal(host)
	assert.Nil(t, err)
	assert.False(t, local)

	host = "xtfoobar056"
	local, err = isLocal(host)
	assert.Nil(t, err)
	assert.False(t, local)

	os.Setenv("APPHANDLE", host)
	local, err = isLocal(host)
	os.Unsetenv("APPHANDLE")
	assert.Nil(t, err)
	assert.True(t, local)

}

func TestSameServer(t *testing.T) {
	assert.True(t, sameServer("snoopd.local", "snoopd.local"))
	assert.False(t, sameServer("foobar", "foobarbaz"))
	assert.True(t, sameServer("x", "x.domain.com"))
	assert.True(t, sameServer("x.domain.com", "x"))
	assert.False(t, sameServer("x.domain.com", "x.bar.com"))
}

func makeTestLocalAndRemoteServers() []config.ServerConfig {
	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	serverConfig := config.ServerConfig{
		Name:                "server1",
		Address:             host,
		Port:                testPort,
		PingURI:             "/xtracrulesok",
		HealthCheck:         "none",
		HealthCheckInterval: 0,
		HealthCheckTimeout:  0,
	}

	serverConfig2 := config.ServerConfig{
		Name:                "server2",
		Address:             "server2.domain.com",
		Port:                testPort,
		PingURI:             "/xtracrulesok",
		HealthCheck:         "none",
		HealthCheckInterval: 0,
		HealthCheckTimeout:  0,
	}

	return []config.ServerConfig{serverConfig, serverConfig2}
}

func makeTestLocalhostAndRemoteServers() []config.ServerConfig {

	serverConfig := config.ServerConfig{
		Name:                "server1",
		Address:             "localhost",
		Port:                testPort,
		PingURI:             "/xtracrulesok",
		HealthCheck:         "none",
		HealthCheckInterval: 0,
		HealthCheckTimeout:  0,
	}

	serverConfig2 := config.ServerConfig{
		Name:                "server2",
		Address:             "server2.domain.com",
		Port:                testPort,
		PingURI:             "/xtracrulesok",
		HealthCheck:         "none",
		HealthCheckInterval: 0,
		HealthCheckTimeout:  0,
	}

	return []config.ServerConfig{serverConfig, serverConfig2}
}

func makeTestRemoteServer() []config.ServerConfig {

	serverConfig2 := config.ServerConfig{
		Name:                "server2",
		Address:             "server2.domain.com",
		Port:                testPort,
		PingURI:             "/xtracrulesok",
		HealthCheck:         "none",
		HealthCheckInterval: 0,
		HealthCheckTimeout:  0,
	}

	return []config.ServerConfig{serverConfig2}
}

func makeTestLocalServer() []config.ServerConfig {
	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	serverConfig := config.ServerConfig{
		Name:                "server1",
		Address:             host,
		Port:                testPort,
		PingURI:             "/xtracrulesok",
		HealthCheck:         "none",
		HealthCheckInterval: 0,
		HealthCheckTimeout:  0,
	}

	return []config.ServerConfig{serverConfig}
}

func TestPartitionServers(t *testing.T) {

	servers := makeTestLocalAndRemoteServers()
	host, _ := os.Hostname()

	local, remote, err := partitionServers(servers)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(local))
	assert.Equal(t, 1, len(remote))
	assert.Equal(t, host, local[0].Address)
	assert.Equal(t, testRemoteServerAddress, remote[0].Address)
}

func TestPreferLocalFactory(t *testing.T) {
	var preferLocalFactory LoadBalancerFactory = new(PreferLocalLoadBalancerFactory)
	pllb, err := preferLocalFactory.NewLoadBalancer("", "", nil)
	assert.NotNil(t, err)
	assert.Nil(t, pllb)

	pllb, err = preferLocalFactory.NewLoadBalancer("foo", "", nil)
	assert.NotNil(t, err)
	assert.Nil(t, pllb)

	pllb, err = preferLocalFactory.NewLoadBalancer("foo", "", makeTestLocalAndRemoteServers())
	assert.Nil(t, err)
	assert.NotNil(t, pllb)
}

func TestPreferLocalGetConnectAddress(t *testing.T) {
	t.Log("Creating prefer-local load balancer")
	var preferLocalFactory LoadBalancerFactory = new(PreferLocalLoadBalancerFactory)
	pllb, err := preferLocalFactory.NewLoadBalancer("foo", "", makeTestLocalAndRemoteServers())
	assert.Nil(t, err)
	assert.NotNil(t, pllb)

	t.Log("Load balancer returns local when it's available")
	host, _ := os.Hostname()
	address, err := pllb.GetConnectAddress()
	assert.Nil(t, err)
	assert.Equal(t, joinHostAndPort(host, testPort), address)

	for i := 0; i < 5; i++ {
		address, err = pllb.GetConnectAddress()
		assert.Nil(t, err)
		assert.Equal(t, joinHostAndPort(host, testPort), address)
	}

	t.Log("When local endpoints are marked down, load balancer returns remote address")
	err = pllb.MarkEndpointDown(joinHostAndPort(host, testPort))
	assert.Nil(t, err)

	address, err = pllb.GetConnectAddress()
	assert.Nil(t, err)
	assert.Equal(t, joinHostAndPort(testRemoteServerAddress, testPort), address)

	t.Log("When all endpoints are marked down, load balancer returns an error")
	err = pllb.MarkEndpointDown(joinHostAndPort(testRemoteServerAddress, testPort))
	assert.Nil(t, err)

	address, err = pllb.GetConnectAddress()
	assert.NotNil(t, err)

	t.Log("When I mark the remote endpoint up again then it get the remote endpoint")
	err = pllb.MarkEndpointUp(joinHostAndPort(testRemoteServerAddress, testPort))
	assert.Nil(t, err)

	address, err = pllb.GetConnectAddress()
	assert.Nil(t, err)
	assert.Equal(t, joinHostAndPort(testRemoteServerAddress, testPort), address)

	t.Log("When I mark the local endpoint up again then it get the local endpoint")
	err = pllb.MarkEndpointUp(joinHostAndPort(host, testPort))
	assert.Nil(t, err)

	address, err = pllb.GetConnectAddress()
	assert.Nil(t, err)
	assert.Equal(t, joinHostAndPort(host, testPort), address)

	h, u := pllb.GetEndpoints()
	assert.Equal(t, 0, len(u))
	assert.Equal(t, 2, len(h))

}

func TestPrefLocalWithLocalOnly(t *testing.T) {
	t.Log("Creating prefer-local load balancer")
	var preferLocalFactory LoadBalancerFactory = new(PreferLocalLoadBalancerFactory)
	pllb, err := preferLocalFactory.NewLoadBalancer("foo", "", makeTestLocalServer())
	assert.Nil(t, err)
	assert.NotNil(t, pllb)

	t.Log("Load balancer returns local when it's available")
	host, _ := os.Hostname()
	address, err := pllb.GetConnectAddress()
	assert.Nil(t, err)
	assert.Equal(t, joinHostAndPort(host, testPort), address)

	t.Log("When local endpoints are marked down, load balancer returns an error")
	err = pllb.MarkEndpointDown(joinHostAndPort(host, testPort))
	assert.Nil(t, err)

	_, err = pllb.GetConnectAddress()
	assert.NotNil(t, err)
}

func TestPrefLocalWithRemoteOnly(t *testing.T) {
	t.Log("Creating prefer-local load balancer")
	var preferLocalFactory LoadBalancerFactory = new(PreferLocalLoadBalancerFactory)
	pllb, err := preferLocalFactory.NewLoadBalancer("foo", "", makeTestRemoteServer())
	assert.Nil(t, err)
	assert.NotNil(t, pllb)

	t.Log("Load balancer returns remote when it's available")
	address, err := pllb.GetConnectAddress()
	assert.Nil(t, err)
	assert.Equal(t, joinHostAndPort(testRemoteServerAddress, testPort), address)

	t.Log("When local endpoints are marked down, load balancer returns an error")
	err = pllb.MarkEndpointDown(joinHostAndPort(testRemoteServerAddress, testPort))
	assert.Nil(t, err)

	_, err = pllb.GetConnectAddress()
	assert.NotNil(t, err)
}

func TestPrefLocalWithLocalhost(t *testing.T) {
	t.Log("Create prefer-local load balancer with localhost and remote servers")
	var preferLocalFactory LoadBalancerFactory = new(PreferLocalLoadBalancerFactory)
	pllb, err := preferLocalFactory.NewLoadBalancer("foo", "", makeTestLocalhostAndRemoteServers())
	assert.Nil(t, err)
	assert.NotNil(t, pllb)

	t.Log("localhost server should always be returned")
	for i := 0; i < 5; i++ {
		address, err := pllb.GetConnectAddress()
		assert.Nil(t, err)
		assert.Equal(t, joinHostAndPort("localhost", testPort), address)
	}

}
