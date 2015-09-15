package commands

import (
	"bytes"
	"github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/mitchellh/cli"
	"github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"testing"
)

func testMakeAddRoute(faultyStore bool, t *testing.T) (*bytes.Buffer, *AddRoute) {

	var kvs, _ = kvstore.NewHashKVStore("")

	var writer = new(bytes.Buffer)
	var ui = &cli.BasicUi{Writer: writer, ErrorWriter: writer}
	var addRoute = &AddRoute{
		UI:      ui,
		KVStore: kvs,
	}

	//Add the backend the command will validate
	b := &config.BackendConfig{"b1", []string{"s1", "s2", "s3"}, ""}
	err := b.Store(kvs)
	assert.Nil(t, err)

	//Enable fault injection after writing the backend def
	if faultyStore {
		kvs.InjectFaults()
	}

	return writer, addRoute
}

func TestAddRoute(t *testing.T) {

	_, addRoute := testMakeAddRoute(false, t)
	assert.NotNil(t, addRoute)

	args := []string{"-name", "route1", "-backend", "b1", "-base-uri", "/foo", "-msgprop", "SOAPAction=\"foo\""}
	status := addRoute.Run(args)
	assert.Equal(t, 0, status)
	storedBytes, err := addRoute.KVStore.Get("routes/route1")
	assert.Nil(t, err)

	r := config.JSONToRoute(storedBytes)
	assert.Equal(t, "b1", r.Backend)
	assert.Equal(t, "route1", r.Name)
	assert.Equal(t, "/foo", r.URIRoot)
	assert.Equal(t, "SOAPAction=\"foo\"", r.MsgProps)
}

func TestAddRouteMissingName(t *testing.T) {
	_, addRoute := testMakeAddRoute(false, t)
	assert.NotNil(t, addRoute)

	args := []string{"-backend", "b1", "-base-uri", "/foo", "-msgprop", "SOAPAction=\"foo\""}
	status := addRoute.Run(args)
	assert.Equal(t, 1, status)
}

func TestAddRouteUnregisteredFilters(t *testing.T) {
	_, addRoute := testMakeAddRoute(false, t)
	assert.NotNil(t, addRoute)

	args := []string{"-name", "route1", "-backend", "b1", "-base-uri", "/foo", "-msgprop", "SOAPAction=\"foo\"", "-filters", "crapFilter"}
	status := addRoute.Run(args)
	assert.Equal(t, 1, status)
}

func TestAddRouteUnknownBackend(t *testing.T) {
	_, addRoute := testMakeAddRoute(false, t)
	assert.NotNil(t, addRoute)

	args := []string{"-name", "route1", "-backend", "unnkown", "-base-uri", "/foo", "-msgprop", "SOAPAction=\"foo\""}
	status := addRoute.Run(args)
	assert.Equal(t, 1, status)
}

func TestAddRouteParseError(t *testing.T) {
	_, addRoute := testMakeAddRoute(false, t)
	assert.NotNil(t, addRoute)

	args := []string{"-crapfest"}
	status := addRoute.Run(args)
	assert.Equal(t, 1, status)
}

func TestAddRouteMissingBaseURI(t *testing.T) {
	_, addRoute := testMakeAddRoute(false, t)
	assert.NotNil(t, addRoute)

	args := []string{"-backend", "b1", "-name", "/foo", "-msgprop", "SOAPAction=\"foo\""}
	status := addRoute.Run(args)
	assert.Equal(t, 1, status)
}

func TestAddRouteMissingBackend(t *testing.T) {
	_, addRoute := testMakeAddRoute(false, t)
	assert.NotNil(t, addRoute)

	args := []string{"-base-uri", "/foo", "-name", "/foo", "-msgprop", "SOAPAction=\"foo\""}
	status := addRoute.Run(args)
	assert.Equal(t, 1, status)
}

func TestAddRouteStorageError(t *testing.T) {

	_, addRoute := testMakeAddRoute(true, t)
	assert.NotNil(t, addRoute)

	args := []string{"-name", "route1", "-backend", "b1", "-base-uri", "/foo", "-msgprop", "SOAPAction=\"foo\""}
	status := addRoute.Run(args)
	assert.Equal(t, 1, status)
}

func TestAddRouteSynopsis(t *testing.T) {
	_, addRoute := testMakeAddRoute(false, t)
	assert.NotNil(t, addRoute)
	assert.NotEqual(t, "", addRoute.Synopsis())
}
