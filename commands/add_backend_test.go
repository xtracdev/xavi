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

func testMakeAddBackend(faultyStore bool) (*bytes.Buffer, *AddBackend) {

	var kvs, _ = kvstore.NewHashKVStore("")
	if faultyStore {
		kvs.InjectFaults()
	}
	var writer = new(bytes.Buffer)
	var ui = &cli.BasicUi{Writer: writer, ErrorWriter: writer}
	var addBackend = &AddBackend{
		UI:      ui,
		KVStore: kvs,
	}

	return writer, addBackend
}

func TestAddBackend(t *testing.T) {
	_, addBackend := testMakeAddBackend(false)
	args := []string{"-name", "test", "-servers", "foo", "-load-balancer-policy", "round-robin"}
	status := addBackend.Run(args)
	assert.Equal(t, 0, status)

	storedBytes, err := addBackend.KVStore.Get("backends/test")
	assert.Nil(t, err)

	b := config.JSONToBackend(storedBytes)
	assert.Equal(t, "test", b.Name)
	assert.True(t, len(b.ServerNames) == 1)
	assert.Equal(t, "foo", b.ServerNames[0])
	assert.Equal(t, "round-robin", b.LoadBalancerPolicy)
}

func TestAddBackendWIthNoLoadBalancer(t *testing.T) {
	_, addBackend := testMakeAddBackend(false)
	args := []string{"-name", "test", "-servers", "foo"}
	status := addBackend.Run(args)
	assert.Equal(t, 0, status)

	storedBytes, err := addBackend.KVStore.Get("backends/test")
	assert.Nil(t, err)

	b := config.JSONToBackend(storedBytes)
	assert.Equal(t, "test", b.Name)
	assert.True(t, len(b.ServerNames) == 1)
	assert.Equal(t, "foo", b.ServerNames[0])
	assert.Equal(t, "", b.LoadBalancerPolicy)
}

func TestAddBackendWithUnregisteredLoadBalancer(t *testing.T) {
	_, addBackend := testMakeAddBackend(false)
	args := []string{"-name", "test", "-servers", "foo", "-load-balancer-policy", "unknown load balaner policy"}
	status := addBackend.Run(args)
	assert.Equal(t, 1, status)
}

func TestAddBackendMissingAllArgs(t *testing.T) {
	writer, addBackend := testMakeAddBackend(false)

	var args []string
	status := addBackend.Run(args)
	assert.Equal(t, 1, status)
	writerAsString := writer.String()
	t.Log("addBackend returned string", writerAsString)
	assert.True(t, strings.Contains(writerAsString, "Usage: xavi add-backend [options]"))
}

func TestAddBackendStoreFault(t *testing.T) {
	writer, addBackend := testMakeAddBackend(true)
	args := []string{"-name", "test", "-servers", "foo"}
	status := addBackend.Run(args)
	assert.Equal(t, 1, status)
	assert.True(t, strings.Contains(writer.String(), "Faulty store"))
}

func TestAddBackendOK(t *testing.T) {
	writer, addBackend := testMakeAddBackend(true)
	args := []string{"-name", "test", "-servers", "foo"}
	status := addBackend.Run(args)
	assert.Equal(t, 1, status)
	assert.True(t, strings.Contains(writer.String(), "Faulty store"))
}

func TestAddBackendSynopsis(t *testing.T) {
	_, addBackend := testMakeAddBackend(false)
	s := addBackend.Synopsis()
	assert.Equal(t, "Define a backend as a collection of servers", s)
}

func TestAddBackendParseArgsError(t *testing.T) {
	_, addBackend := testMakeAddBackend(false)
	args := []string{"-foofest"}
	status := addBackend.Run(args)
	assert.Equal(t, 1, status)
}
