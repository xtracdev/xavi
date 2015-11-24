package service

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestBuildRouteWithUnknownName(t *testing.T) {
	kvs := initKVStore(t)
	_, err := buildRoute("no such route", kvs)
	assert.NotNil(t, err)
}

func TestBuildRoutesWithNoPlugin(t *testing.T) {
	kvs := initKVStore(t)
	_, err := buildRoute("r1", kvs)
	if assert.NotNil(t, err) {
		assert.True(t, strings.Contains(err.Error(), "MultiRoute"))
	}
}

func TestMultiRouteConfig(t *testing.T) {
	kvs := initKVStore(t)
	r, err := buildRoute("r2", kvs)
	assert.Nil(t, err)
	assert.Equal(t, "r2", r.Name)
	assert.Equal(t, 2, len(r.Backends))
	assert.Equal(t, "be1", r.Backends[0].Name)
	assert.Equal(t, "be2", r.Backends[1].Name)
	assert.Equal(t, "test-plugin", r.MultiRoutePluginName)
}

func TestMultiRouteNoBackend(t *testing.T) {
	kvs := initKVStore(t)
	_, err := buildRoute("r3", kvs)
	assert.NotNil(t, err)
}
