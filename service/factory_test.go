package service

import (
	"fmt"
	"strings"
	"testing"

	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"github.com/xtracdev/xavi/plugin"
	"github.com/xtracdev/xavi/plugin/logging"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

func initKVStore(t *testing.T) kvstore.KVStore {
	return config.BuildKVStoreTestConfig(t)
}

func TestServerFactory(t *testing.T) {
	plugin.RegisterWrapperFactory("Logging", logging.NewLoggingWrapper)

	var testKVS = initKVStore(t)
	service, err := BuildServiceForListener("listener", "0.0.0.0:8000", testKVS)
	assert.Nil(t, err)
	s := fmt.Sprintf("%s", service)
	println(s)
	assert.True(t, strings.Contains(s, "route1"))
	assert.True(t, strings.Contains(s, "hello-backend"))
	assert.True(t, strings.Contains(s, "listener"))
	assert.True(t, strings.Contains(s, "0.0.0.0:8000"))

	ms, ok := service.(*managedService)
	assert.True(t, ok)
	hch := ms.HealthCheckContext.HealthHandler()
	assert.NotNil(t, hch)

	ts := httptest.NewServer(hch)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.Nil(t, err)
	respbytes, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var health HealthResponse

	err = json.Unmarshal(respbytes, &health)
	assert.Nil(t, err)
	assert.Equal(t, "listener", health.ListenerName)
	assert.Equal(t, 1, len(health.Routes))
	assert.Equal(t, "route1", health.Routes[0].Name)
	assert.True(t, health.Routes[0].Up)
	assert.Equal(t, 1, len(health.Routes[0].Backends))
	assert.Equal(t, "hello-backend", health.Routes[0].Backends[0].Name)
	assert.Equal(t, 2, len(health.Routes[0].Backends[0].HealthyDependencies))
	assert.Equal(t, "localhost:3000", health.Routes[0].Backends[0].HealthyDependencies[0])
	assert.Equal(t, "localhost:3100", health.Routes[0].Backends[0].HealthyDependencies[1])
	assert.Equal(t, true, health.Routes[0].Backends[0].Up)
	assert.Equal(t, 0, len(health.Routes[0].Backends[0].UnhealthyDependencies))

}

func TestHealthCheckMultiBackendRoute(t *testing.T) {
	var testKVS = initKVStore(t)
	service, err := BuildServiceForListener("l2", "0.0.0.0:8000", testKVS)
	assert.Nil(t, err)

	ms, ok := service.(*managedService)
	assert.True(t, ok)
	hch := ms.HealthCheckContext.HealthHandler()
	assert.NotNil(t, hch)

	ts := httptest.NewServer(hch)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.Nil(t, err)
	respbytes, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var health HealthResponse
	err = json.Unmarshal(respbytes, &health)
	assert.Nil(t, err)

	assert.Equal(t, "l2", health.ListenerName)
	assert.Equal(t, 1, len(health.Routes))
	assert.Equal(t, "r2", health.Routes[0].Name)
	assert.True(t, health.Routes[0].Up)
	assert.Equal(t, 2, len(health.Routes[0].Backends))

	be1 := health.Routes[0].Backends[0]
	assert.Equal(t, "be1", be1.Name)
	assert.True(t, be1.Up)
	assert.Equal(t, 2, len(be1.HealthyDependencies))
	assert.Equal(t, 0, len(be1.UnhealthyDependencies))
	assert.Equal(t, "localhost:3000", be1.HealthyDependencies[0])
	assert.Equal(t, "localhost:3100", be1.HealthyDependencies[1])

}

func TestServerFactoryWithUnknownListener(t *testing.T) {
	var testKVS = initKVStore(t)
	_, err := BuildServiceForListener("no such listener", "0.0.0.0:8000", testKVS)
	assert.NotNil(t, err)
}
