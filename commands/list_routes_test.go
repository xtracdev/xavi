package commands

import (
	"bytes"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"testing"
)

func testMakeListRoutes(faultyStore bool, withRoute bool) (*bytes.Buffer, *RouteList) {

	var kvs, _ = kvstore.NewHashKVStore("")

	var writer = new(bytes.Buffer)
	var ui = &cli.BasicUi{Writer: writer, ErrorWriter: writer}
	var listRoutes = &RouteList{
		UI:      ui,
		KVStore: kvs,
	}

	if withRoute {
		b := &config.RouteConfig{
			Name:     "Route1",
			URIRoot:  "/foo",
			Backends: []string{"b1"},
			Plugins:  nil,
			MsgProps: "",
		}
		b.Store(kvs)
	}

	if faultyStore {
		kvs.InjectFaults()
	}

	return writer, listRoutes
}

func TestListRouteEmpty(t *testing.T) {
	writer, listRoutes := testMakeListRoutes(false, false)
	var args []string
	status := listRoutes.Run(args)
	assert.Equal(t, 0, status)
	out := string(writer.Bytes())
	assert.Contains(t, out, "[]")
}

func TestListRouteNonEmpty(t *testing.T) {
	writer, listRoutes := testMakeListRoutes(false, true)
	var args []string
	status := listRoutes.Run(args)
	assert.Equal(t, 0, status)
	out := string(writer.Bytes())
	assert.Contains(t, out, "Route1")
	assert.Contains(t, out, "/foo")
	assert.Contains(t, out, "b1")

}

func TestListRouteFaultyStore(t *testing.T) {
	_, listRoutes := testMakeListRoutes(true, false)
	var args []string
	status := listRoutes.Run(args)
	assert.Equal(t, 1, status)
}

func TestListRouteHelp(t *testing.T) {
	_, listRoutes := testMakeListRoutes(false, false)
	assert.NotEmpty(t, listRoutes.Help())
}

func TestListRouteSynopsis(t *testing.T) {
	_, listRoutes := testMakeListRoutes(false, false)
	assert.NotEmpty(t, listRoutes.Synopsis())
}
