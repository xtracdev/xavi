package config

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func simpleHC(endpoint string, transport *http.Transport) <-chan bool {
	statusChannel := make(chan bool)

	go func() {
		statusChannel <- true
	}()

	return statusChannel
}

func TestCustomHCNoFunction(t *testing.T) {
	kvs := BuildKVStoreTestConfig(t)
	ListenContext = true
	err := RegisterHealthCheckForServer(kvs, "not a server name", nil)
	ListenContext = false
	if assert.NotNil(t, err) {
		assert.Equal(t, err, ErrNoHealthCheckFn)
	}
}

func TestCustomHCNoSuchServer(t *testing.T) {
	kvs := BuildKVStoreTestConfig(t)
	ListenContext = true
	err := RegisterHealthCheckForServer(kvs, "not a server name", simpleHC)
	ListenContext = false
	if assert.NotNil(t, err) {
		assert.Equal(t, err, ErrNoSuchServer)
	}
}

func TestCustomHCLookup(t *testing.T) {
	kvs := BuildKVStoreTestConfig(t)
	var hc1 HealthCheckFn = simpleHC
	ListenContext = true
	err := RegisterHealthCheckForServer(kvs, "server1", hc1)
	ListenContext = false

	if assert.Nil(t, err) {
		hcfn := HealthCheckForServer("server1")
		assert.NotNil(t, hcfn)
	}
}

func TestCustomHCNoSuchBackend(t *testing.T) {
	kvs := BuildKVStoreTestConfig(t)
	ListenContext = true
	err := RegisterHealthCheckForBackend(kvs, "Not a backend", simpleHC)
	ListenContext = false
	if assert.NotNil(t, err) {
		assert.Equal(t, err, ErrNoSuchBackend)
	}
}

func TestCustomHCNoFnForBackend(t *testing.T) {
	kvs := BuildKVStoreTestConfig(t)
	ListenContext = true
	err := RegisterHealthCheckForBackend(kvs, "Not a backend", nil)
	ListenContext = false
	if assert.NotNil(t, err) {
		assert.Equal(t, err, ErrNoHealthCheckFn)
	}
}

func TestCustomHCBackendConfig(t *testing.T) {
	kvs := BuildKVStoreTestConfig(t)
	var hc1 HealthCheckFn = simpleHC
	ListenContext = true
	err := RegisterHealthCheckForBackend(kvs, "hello-backend", hc1)
	ListenContext = false
	if assert.Nil(t, err) {
		hcfn := HealthCheckForServer("server1")
		assert.NotNil(t, hcfn)
		hcfn = HealthCheckForServer("server2")
		assert.NotNil(t, hcfn)
	}
}
