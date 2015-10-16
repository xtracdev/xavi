package commands

import (
	"bytes"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"strings"
	"testing"
)

func testMakeListServers(faultyStore bool, withServer bool) (*bytes.Buffer, *ServerList) {

	var kvs, _ = kvstore.NewHashKVStore("")

	var writer = new(bytes.Buffer)
	var ui = &cli.BasicUi{Writer: writer, ErrorWriter: writer}
	var listServers = &ServerList{
		UI:      ui,
		KVStore: kvs,
	}

	if withServer {
		b := &config.ServerConfig{"name", "host", 123, "ping", "none", 0, 0}
		b.Store(kvs)
	}

	if faultyStore {
		kvs.InjectFaults()
	}

	return writer, listServers
}

func TestListServerEmpty(t *testing.T) {
	writer, listServers := testMakeListServers(false, false)
	var args []string
	status := listServers.Run(args)
	assert.Equal(t, 0, status)
	out := string(writer.Bytes())
	assert.True(t, strings.Contains(out, "[]"))
}

func TestListServerNonEmpty(t *testing.T) {
	writer, listServers := testMakeListServers(false, true)
	var args []string
	status := listServers.Run(args)
	assert.Equal(t, 0, status)
	out := string(writer.Bytes())
	assert.True(t, strings.Contains(out, "name"))
	assert.True(t, strings.Contains(out, "host"))
	assert.True(t, strings.Contains(out, "123"))
	assert.True(t, strings.Contains(out, "ping"))
}

func TestListServerFaultyStore(t *testing.T) {
	_, listServers := testMakeListServers(true, false)
	var args []string
	status := listServers.Run(args)
	assert.Equal(t, 1, status)
}

func TestListServerHelp(t *testing.T) {
	_, listServers := testMakeListServers(false, false)
	assert.NotEmpty(t, listServers.Help())
}

func TestListServerSynopsis(t *testing.T) {
	_, listServers := testMakeListServers(false, false)
	assert.NotEmpty(t, listServers.Synopsis())
}
