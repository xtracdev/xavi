package config

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/kvstore"
	"testing"
)

func TestJSON2Server(t *testing.T) {
	serverDef := `
		{
			"name":"s1",
			"address":"0.0.0.0",
			"port":5000,
			"pingURI":"/ping",
			"healthCheck":"none",
			"healthCheckInterval":0,
			"healthCheckTimeout":0}`

	var s ServerConfig

	json.Unmarshal([]byte(serverDef), &s)

	testVerifyServerRead(&s, t)
}

func testVerifyServerRead(server *ServerConfig, t *testing.T) {
	assert.Equal(t, server.Name, "s1")
	assert.Equal(t, server.Address, "0.0.0.0")
	assert.Equal(t, server.Port, 5000)
	assert.Equal(t, server.PingURI, "/ping")
	assert.Equal(t, server.HealthCheck, "none")
	assert.Equal(t, server.HealthCheckInterval, 0)
	assert.Equal(t, server.HealthCheckTimeout, 0)
}

func TestServerStoreAndRetrieve(t *testing.T) {
	var testKVS, _ = kvstore.NewHashKVStore("")

	//Read - not found
	server, err := ReadServerConfig("s1", testKVS)
	assert.Nil(t, err)
	assert.Nil(t, server, "Expected server to be nil")

	//Read - empty list
	servers, err := ListServerConfigs(testKVS)
	assert.Nil(t, err)
	assert.Nil(t, servers)

	//Store
	server = &ServerConfig{"s1", "0.0.0.0", 5000, "/ping", "none", 0, 0}
	err = server.Store(testKVS)
	assert.Nil(t, err)

	//Read - found
	server, err = ReadServerConfig("s1", testKVS)
	assert.Nil(t, err)
	testVerifyServerRead(server, t)

	//Grab a list of servers
	servers, err = ListServerConfigs(testKVS)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(servers))
	testVerifyServerRead(servers[0], t)

}
