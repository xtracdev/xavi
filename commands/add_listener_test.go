package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
)

func testMakeAddListener(faultyStore bool) (*bytes.Buffer, *AddListener) {

	var kvs, _ = kvstore.NewHashKVStore("")
	if faultyStore {
		kvs.InjectFaults()
	}
	var writer = new(bytes.Buffer)
	var ui = &cli.BasicUi{Writer: writer, ErrorWriter: writer}
	var addListener = &AddListener{
		UI:      ui,
		KVStore: kvs,
	}

	return writer, addListener
}

func TestAddListenerMissingAllArgs(t *testing.T) {
	writer, addListener := testMakeAddListener(false)

	var args []string
	status := addListener.Run(args)
	assert.Equal(t, 1, status)
	assert.True(t, strings.Contains(writer.String(), addListener.Help()))
}

func TestAddListener(t *testing.T) {
	_, addListener := testMakeAddListener(false)
	args := []string{"-name", "test", "-routes", "foo,bar"}
	status := addListener.Run(args)
	assert.Equal(t, 0, status)

	storedBytes, err := addListener.KVStore.Get("listeners/test")
	assert.Nil(t, err)

	b := config.JSONToListener(storedBytes)
	assert.Equal(t, "test", b.Name)
	assert.True(t, len(b.RouteNames) == 2)
	assert.Equal(t, "foo", b.RouteNames[0])
	assert.Equal(t, "bar", b.RouteNames[1])
}

func TestAddListenerParseArgsError(t *testing.T) {
	_, addListener := testMakeAddListener(false)
	args := []string{"-foofest"}
	status := addListener.Run(args)
	assert.Equal(t, 1, status)
}

func TestAddListenerStoreFault(t *testing.T) {
	writer, addListener := testMakeAddListener(true)
	args := []string{"-name", "test", "-routes", "foo,bar"}
	status := addListener.Run(args)
	assert.Equal(t, 1, status)
	assert.True(t, strings.Contains(writer.String(), "Faulty store"))
}
