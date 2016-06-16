package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMissingListener(t *testing.T) {
	sc, err := ReadServiceConfig("", nil)
	assert.Nil(t, sc)
	assert.Equal(t, ErrNoListenerName, err)
}

func TestMissinKVStore(t *testing.T) {
	sc, err := ReadServiceConfig("imma-listening", nil)
	assert.Nil(t, sc)
	assert.Equal(t, ErrNoKVStore, err)
}

func TestBuildServiceConfig(t *testing.T) {
	kvs := BuildKVStoreTestConfig(t)
	assert.NotNil(t, kvs)
	sc, err := ReadServiceConfig("listener", kvs)
	assert.Nil(t, err)
	assert.NotNil(t, sc)
	if assert.NotNil(t, sc.Listener) {
		listener := sc.Listener
		assert.Equal(t, "listener", listener.Name)
		if assert.Equal(t, 1, len(sc.Routes)) {
			assert.Equal(t, "route1", sc.Routes[0].Route.Name)
			assert.Equal(t, "/hello", sc.Routes[0].Route.URIRoot)
			if assert.Equal(t, 1, len(sc.Routes[0].Backends)) {
				backend := sc.Routes[0].Backends[0]
				assert.Equal(t, "hello-backend", backend.Backend.Name)
				if assert.Equal(t, 2, len(backend.Servers)) {
					assert.Equal(t, "server1", backend.Servers[0].Name)
					assert.Equal(t, "server2", backend.Servers[1].Name)
				}
			}
		}
	}
}
