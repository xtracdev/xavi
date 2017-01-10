package commands

import (
	"bytes"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"testing"
)

func testMakeListListeners(faultyStore bool, withListener bool) (*bytes.Buffer, *ListenerList) {

	var kvs, _ = kvstore.NewHashKVStore("")
	if faultyStore {
		kvs.InjectFaults()
	}
	var writer = new(bytes.Buffer)
	var ui = &cli.BasicUi{Writer: writer, ErrorWriter: writer}
	var listListeners = &ListenerList{
		UI:      ui,
		KVStore: kvs,
	}

	if withListener {
		b := &config.ListenerConfig{"l1", []string{"r1", "r2", "r3"}, true}
		b.Store(kvs)
	}

	return writer, listListeners
}

func TestListListenerEmpty(t *testing.T) {
	writer, listListeners := testMakeListListeners(false, false)
	var args []string
	status := listListeners.Run(args)
	assert.Equal(t, 0, status)
	out := string(writer.Bytes())
	assert.Contains(t, out, "[]")
}

func TestListListenerNonEmpty(t *testing.T) {
	writer, listListeners := testMakeListListeners(false, true)
	var args []string
	status := listListeners.Run(args)
	assert.Equal(t, 0, status)
	out := string(writer.Bytes())
	assert.Contains(t, out, "l1")
	assert.Contains(t, out, "r1")
	assert.Contains(t, out, "r2")
	assert.Contains(t, out, "r3")
}

func TestListListenerFaultyStore(t *testing.T) {
	_, listListeners := testMakeListListeners(true, false)
	var args []string
	status := listListeners.Run(args)
	assert.Equal(t, 1, status)
}

func TestListListenerHelp(t *testing.T) {
	_, listListeners := testMakeListListeners(false, false)
	assert.NotEmpty(t, listListeners.Help())
}

func TestListListenerSynopsis(t *testing.T) {
	_, listListeners := testMakeListListeners(false, false)
	assert.NotEmpty(t, listListeners.Synopsis())
}
