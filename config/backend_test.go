package config

import (
	"encoding/json"
	"github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/kvstore"
	"testing"
)

func TestJSON2Backend(t *testing.T) {
	backendDef := `
	{"name":"hello-backend", "serverNames" : ["s1","s2", "s3"],"loadBalancerPolicy":"round-robin"}
	`
	var b BackendConfig

	json.Unmarshal([]byte(backendDef), &b)

	testVerifyBackendConfig(&b, t)
}

func testVerifyBackendConfig(b *BackendConfig, t *testing.T) {
	assert.Equal(t, "hello-backend", b.Name)
	assert.Equal(t, 3, len(b.ServerNames))
	assert.Equal(t, "s1", b.ServerNames[0])
	assert.Equal(t, "s2", b.ServerNames[1])
	assert.Equal(t, "s3", b.ServerNames[2])
	assert.Equal(t, "round-robin", b.LoadBalancerPolicy)
}

func TestBackendStoreAndRetrieve(t *testing.T) {
	var testKVS, _ = kvstore.NewHashKVStore("")

	//Read - not found
	b, err := ReadBackendConfig("hello-backend", testKVS)
	assert.Nil(t, err)
	assert.Nil(t, b, "Expected backend to be nil")

	//Read - empty list
	backends, err := ListBackendConfigs(testKVS)
	assert.Nil(t, err)
	assert.Nil(t, backends)

	//Store
	b = &BackendConfig{"hello-backend", []string{"s1", "s2", "s3"}, "round-robin"}
	err = b.Store(testKVS)
	assert.Nil(t, err)

	//Read - found
	b, err = ReadBackendConfig("hello-backend", testKVS)
	assert.Nil(t, err)
	testVerifyBackendConfig(b, t)

	//Grab a list of backends
	backends, err = ListBackendConfigs(testKVS)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(backends))

	testVerifyBackendConfig(backends[0], t)
}
