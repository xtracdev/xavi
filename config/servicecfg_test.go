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
	assert.NotNil(t,kvs)
	sc,err := ReadServiceConfig("listener",kvs)
	assert.Nil(t,err)
	assert.NotNil(t,sc)
	if assert.NotNil(t,sc.Listener) {
		listener := sc.Listener
		assert.Equal(t, "listener", listener.Name)
		assert.Equal(t, 0, len(sc.Routes))
	}
}
