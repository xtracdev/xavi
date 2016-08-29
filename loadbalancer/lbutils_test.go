package loadbalancer

import (
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"testing"
)

func TestLBUtilsBuildFromConfig(t *testing.T) {
	kvs := config.BuildKVStoreTestConfig(t)
	assert.NotNil(t, kvs)
	sc, err := config.ReadServiceConfig("listener", kvs)
	assert.Nil(t, err)
	assert.NotNil(t, sc)

	config.RecordActiveConfig(sc)

	lb, backend, err := NewLoadBalancer("route1", "hello-backend")
	assert.Nil(t, err)
	assert.NotNil(t, lb)
	assert.NotNil(t, backend)

	assert.Equal(t, "hello-backend", backend.Name)
	assert.Equal(t, "", backend.CACertPath)
	assert.Equal(t, 2, len(backend.ServerNames))

	h, _ := lb.GetEndpoints()
	if assert.True(t, len(h) == 2) {
		assert.Equal(t, "localhost:3000", h[0])
		assert.Equal(t, "localhost:3100", h[1])
	}
}

func TestLBUtilsNoSuchBackend(t *testing.T) {
	kvs := config.BuildKVStoreTestConfig(t)
	assert.NotNil(t, kvs)
	sc, err := config.ReadServiceConfig("listener", kvs)
	assert.Nil(t, err)
	assert.NotNil(t, sc)

	config.RecordActiveConfig(sc)

	lb, backend, err := NewLoadBalancer("route1", "no-such-backed")
	assert.Nil(t, lb)
	assert.Nil(t, backend)
	assert.NotNil(t, err)
	assert.Equal(t, ErrBackendNotFound, err)
}
