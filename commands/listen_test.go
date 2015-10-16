package commands

import (
	"bytes"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"testing"
)

func testMakeListenCmd(faultyStore bool, withListener bool) (*bytes.Buffer, *Listen) {

	var kvs, _ = kvstore.NewHashKVStore("")

	var writer = new(bytes.Buffer)
	var ui = &cli.BasicUi{Writer: writer, ErrorWriter: writer}
	var listener = &Listen{
		UI:      ui,
		KVStore: kvs,
	}

	if withListener {
		b := &config.ListenerConfig{"l1", []string{}}
		b.Store(kvs)
	}

	if faultyStore {
		kvs.InjectFaults()
	}

	return writer, listener
}

func TestListenerSynopsis(t *testing.T) {
	_, listener := testMakeListenCmd(false, false)
	assert.NotEmpty(t, listener.Synopsis())
}

func TestListenerHelp(t *testing.T) {
	_, listener := testMakeListenCmd(false, false)
	assert.NotEmpty(t, listener.Synopsis())
}

func TestListenerParseError(t *testing.T) {
	_, listener := testMakeListenCmd(false, false)
	var args = []string{"-crapfest"}
	status := listener.Run(args)
	assert.Equal(t, 1, status)
}

func TestListenerMissingArgs(t *testing.T) {
	_, listener := testMakeListenCmd(false, false)
	var args []string
	status := listener.Run(args)
	assert.Equal(t, 1, status)
}

func TestListenerArgsWithNonexistentDef(t *testing.T) {
	_, listener := testMakeListenCmd(false, false)
	var args = []string{"-ln", "larry", "-address", "0.0.0.0:666", "-cpuprofile", "cpuxxx"}
	status := listener.Run(args)
	assert.Equal(t, 1, status)
}

func TestListenerWithInvalidProfilePath(t *testing.T) {
	_, listener := testMakeListenCmd(false, false)
	var args = []string{"-ln", "larry", "-address", "0.0.0.0:666", "-cpuprofile", "/yabba/dabba/doo"}
	status := listener.Run(args)
	assert.Equal(t, 1, status)
}

func TestListenerRunWithSystemPort(t *testing.T) {
	_, listener := testMakeListenCmd(false, true)
	var args = []string{"-ln", "l1", "-address", "0.0.0.0:80"}
	status := listener.Run(args)
	assert.Equal(t, 1, status)
}
