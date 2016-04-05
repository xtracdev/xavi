package commands

import (
	"bytes"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"testing"
)

func testMakeListBackends(faultyStore bool, withBackend bool) (*bytes.Buffer, *BackendList) {

	var kvs, _ = kvstore.NewHashKVStore("")
	if faultyStore {
		kvs.InjectFaults()
	}
	var writer = new(bytes.Buffer)
	var ui = &cli.BasicUi{Writer: writer, ErrorWriter: writer}
	var listBackends = &BackendList{
		UI:      ui,
		KVStore: kvs,
	}

	if withBackend {
		b := &config.BackendConfig{
			Name:        "b1",
			ServerNames: []string{"s1", "s2", "s3"},
		}
		b.Store(kvs)
	}

	return writer, listBackends
}

func TestListBackendEmpty(t *testing.T) {
	writer, listBackends := testMakeListBackends(false, false)
	var args []string
	status := listBackends.Run(args)
	assert.Equal(t, 0, status)
	out := string(writer.Bytes())
	assert.Contains(t, out, "[]")
}

func TestListBackendNonEmpty(t *testing.T) {
	writer, listBackends := testMakeListBackends(false, true)
	var args []string
	status := listBackends.Run(args)
	assert.Equal(t, 0, status)
	out := string(writer.Bytes())
	assert.Contains(t, out, "b1")
	assert.Contains(t, out, "s1")
	assert.Contains(t, out, "s2")
	assert.Contains(t, out, "s3")
}

func TestListBackendFaultyStore(t *testing.T) {
	_, listBackends := testMakeListBackends(true, false)
	var args []string
	status := listBackends.Run(args)
	assert.Equal(t, 1, status)
}

func TestListBackendHelp(t *testing.T) {
	_, listBackends := testMakeListBackends(false, false)
	assert.NotEmpty(t, listBackends.Help())
}

func TestListBackendSynopsis(t *testing.T) {
	_, listBackends := testMakeListBackends(false, false)
	assert.NotEmpty(t, listBackends.Synopsis())
}
