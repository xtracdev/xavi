package config

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/kvstore"
	"testing"
)

func TestJSON2Listener(t *testing.T) {
	listenerDef := `
		{
			"name":"l1", "routeNames":["route1","route2"], "healthEndpoint":false
		}
	`
	var l ListenerConfig

	json.Unmarshal([]byte(listenerDef), &l)
	testVerifyListenerRead(&l, t)

}

func testVerifyListenerRead(ln *ListenerConfig, t *testing.T) {
	assert.Equal(t, "l1", ln.Name)
	assert.Equal(t, 2, len(ln.RouteNames))
	assert.Equal(t, "route1", ln.RouteNames[0])
	assert.Equal(t, "route2", ln.RouteNames[1])
	assert.False(t, ln.HealthEndpoint)
}

func TestListenerStoreAndRetrieve(t *testing.T) {
	var testKVS, _ = kvstore.NewHashKVStore("")

	//Read - not found
	ln, err := ReadListenerConfig("l1", testKVS)
	assert.Nil(t, err)
	assert.Nil(t, ln, "Expected listener to be nil")

	//Read - empty list
	listeners, err := ListListenerConfigs(testKVS)
	assert.Nil(t, err)
	assert.Nil(t, listeners)

	//Store
	ln = &ListenerConfig{"l1", []string{"route1", "route2"}, false}
	err = ln.Store(testKVS)
	assert.Nil(t, err)

	//Read - found
	ln, err = ReadListenerConfig("l1", testKVS)
	assert.Nil(t, err)
	testVerifyListenerRead(ln, t)

	//Grab a list of backends
	listeners, err = ListListenerConfigs(testKVS)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(listeners))
	testVerifyListenerRead(listeners[0], t)
}
