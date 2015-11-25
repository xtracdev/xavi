package config

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/kvstore"
	"testing"
)

func TestJSON2Route(t *testing.T) {
	routeDef := `
		{
			"name":"route1",
			"uriRoot":"/hello",
			"backends":["hello-backend","hello-backend2"],
			"plugins":["plugin1","plugin2","plugin3"],
			"MsgProps":"SOAPAction:\"foo\"",
			"MultiBackendAdapter":"multi-backend-adapter"
		}`

	var r RouteConfig

	json.Unmarshal([]byte(routeDef), &r)
	testVerifyRouteRead(&r, t)

}

func testVerifyRouteRead(r *RouteConfig, t *testing.T) {
	assert.Equal(t, "route1", r.Name)
	assert.Equal(t, "/hello", r.URIRoot)
	assert.Equal(t, 2, len(r.Backends))
	assert.Equal(t, "hello-backend", r.Backends[0])
	assert.Equal(t, "hello-backend2", r.Backends[1])
	assert.Equal(t, 3, len(r.Plugins))
	assert.Equal(t, "plugin1", r.Plugins[0])
	assert.Equal(t, "plugin2", r.Plugins[1])
	assert.Equal(t, "plugin3", r.Plugins[2])
	assert.Equal(t, "SOAPAction:\"foo\"", r.MsgProps)
	assert.Equal(t, "multi-backend-adapter", r.MultiBackendAdapter)
}

func TestRouteStoreAndRetrieve(t *testing.T) {
	var testKVS, _ = kvstore.NewHashKVStore("")

	//Read - not found
	r, err := ReadRouteConfig("route1", testKVS)
	assert.Nil(t, err)
	assert.Nil(t, r, "Expected route to be nil")

	//Read - empty list
	routes, err := ListRouteConfigs(testKVS)
	assert.Nil(t, err)
	assert.Nil(t, routes)

	//Store
	var plugins = []string{"plugin1", "plugin2", "plugin3"}
	r = &RouteConfig{
		Name:                "route1",
		URIRoot:             "/hello",
		Backends:            []string{"hello-backend", "hello-backend2"},
		Plugins:             plugins,
		MsgProps:            "SOAPAction:\"foo\"",
		MultiBackendAdapter: "multi-backend-adapter",
	}
	err = r.Store(testKVS)
	assert.Nil(t, err)

	//Read - found
	r, err = ReadRouteConfig("route1", testKVS)
	assert.Nil(t, err)
	testVerifyRouteRead(r, t)

	//Read - list
	routes, err = ListRouteConfigs(testKVS)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(routes))
	testVerifyRouteRead(routes[0], t)
}
