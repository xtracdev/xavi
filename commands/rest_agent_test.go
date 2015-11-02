package commands

import (
	"bytes"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/kvstore"
	"testing"
)

func testMakeRestAgentCmd() (*bytes.Buffer, *RESTAgent) {

	var kvs, _ = kvstore.NewHashKVStore("")

	var writer = new(bytes.Buffer)
	var ui = &cli.BasicUi{Writer: writer, ErrorWriter: writer}
	var listener = &RESTAgent{
		UI:      ui,
		KVStore: kvs,
	}

	return writer, listener
}

func TestRestAgentSynopsis(t *testing.T) {
	_, listener := testMakeRestAgentCmd()
	assert.NotEmpty(t, listener.Synopsis())
}

func TestRestAgentHelp(t *testing.T) {
	_, listener := testMakeRestAgentCmd()
	assert.NotEmpty(t, listener.Synopsis())
}

func TestRestAgentParseError(t *testing.T) {
	_, listener := testMakeRestAgentCmd()
	var args = []string{"-foofest"}
	status := listener.Run(args)
	assert.Equal(t, 1, status)
}

func TestRestAgenterMissingArgs(t *testing.T) {
	_, listener := testMakeRestAgentCmd()
	var args []string
	status := listener.Run(args)
	assert.Equal(t, 1, status)
}

func TestRestAgenterRunWithSystemPort(t *testing.T) {
	_, listener := testMakeRestAgentCmd()
	var args = []string{"-address", "0.0.0.0:80"}
	status := listener.Run(args)
	assert.Equal(t, 1, status)
}
