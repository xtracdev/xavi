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

func testMakeAddServer(faultyStore bool) (*bytes.Buffer, *AddServer) {

	var kvs, _ = kvstore.NewHashKVStore("")
	if faultyStore {
		kvs.InjectFaults()
	}
	var writer = new(bytes.Buffer)
	var ui = &cli.BasicUi{Writer: writer, ErrorWriter: writer}
	var addServer = &AddServer{
		UI:      ui,
		KVStore: kvs,
	}

	return writer, addServer
}

func TestAddServer(t *testing.T) {

	_, addServer := testMakeAddServer(false)

	args := []string{"-address", "an-address", "-port", "42", "-name", "test-name", "-ping-uri", "/dev/null",
		"-health-check", "http-get", "-health-check-interval", "42", "-health-check-timeout", "10"}
	addServer.Run(args)
	storedBytes, err := addServer.KVStore.Get("servers/test-name")
	assert.Nil(t, err)

	s := config.JSONToServer(storedBytes)
	assert.Equal(t, "test-name", s.Name)
	assert.Equal(t, "an-address", s.Address)
	assert.Equal(t, 42, s.Port)
	assert.Equal(t, "/dev/null", s.PingURI)
	assert.Equal(t, "http-get", s.HealthCheck)
	assert.Equal(t, 42, s.HealthCheckInterval)
	assert.Equal(t, 10, s.HealthCheckTimeout)
}

func TestAddServerParseArgsError(t *testing.T) {
	_, addServer := testMakeAddServer(false)
	args := []string{"-foofest"}
	status := addServer.Run(args)
	assert.Equal(t, 1, status)
}

func TestAddServerErrorHandling(t *testing.T) {

	_, addServer := testMakeAddServer(true)

	args := []string{"-address", "an-address", "-port", "42", "-name", "test-name", "-ping-uri", "/dev/null"}
	status := addServer.Run(args)
	assert.Equal(t, 1, status)
	_, err := addServer.KVStore.Get("servers/test-name")
	assert.NotNil(t, err)

}

func TestAddServerMissingAllArgs(t *testing.T) {
	writer, addServer := testMakeAddServer(false)

	var args []string
	status := addServer.Run(args)
	assert.Equal(t, 1, status)
	assert.True(t, strings.Contains(writer.String(), addServer.Help()))
}

func TestAddServerMissingAddress(t *testing.T) {
	writer, addServer := testMakeAddServer(false)

	args := []string{"-port", "42", "-name", "test-name", "-ping-uri", "/dev/null"}
	status := addServer.Run(args)
	assert.Equal(t, 1, status)
	assert.True(t, strings.Contains(writer.String(), "Address name must be specified"))
	assert.False(t, strings.Contains(writer.String(), "Port must be specified"))
	assert.False(t, strings.Contains(writer.String(), "Name must be specified"))
}

func TestAddServerMissingPort(t *testing.T) {
	writer, addServer := testMakeAddServer(false)

	args := []string{"-address", "an-address", "-name", "test-name", "-ping-uri", "/dev/null"}
	status := addServer.Run(args)
	assert.Equal(t, 1, status)
	assert.False(t, strings.Contains(writer.String(), "Address name must be specified"))
	assert.True(t, strings.Contains(writer.String(), "Port must be specified"))
	assert.False(t, strings.Contains(writer.String(), "Name must be specified"))
}

func TestAddServerMissingName(t *testing.T) {
	writer, addServer := testMakeAddServer(false)

	args := []string{"-address", "an-address", "-port", "42", "-ping-uri", "/dev/null"}
	status := addServer.Run(args)
	assert.Equal(t, 1, status)
	assert.False(t, strings.Contains(writer.String(), "Address name must be specified"))
	assert.False(t, strings.Contains(writer.String(), "Port must be specified"))
	assert.True(t, strings.Contains(writer.String(), "Name must be specified"))
}

func TestAddServerInvalidHealthCheck(t *testing.T) {
	writer, addServer := testMakeAddServer(false)

	args := []string{"-address", "an-address", "-port", "42", "-name", "test-name", "-ping-uri", "/dev/null",
		"-health-check", "invalid-health-check", "-health-check-interval", "42", "-health-check-timeout", "10"}
	status := addServer.Run(args)
	assert.Equal(t, 1, status)
	assert.True(t, strings.Contains(writer.String(), "health check"))
}

func TestAddServerInvalidHealthCheckTimeout(t *testing.T) {
	writer, addServer := testMakeAddServer(false)

	args := []string{"-address", "an-address", "-port", "42", "-name", "test-name", "-ping-uri", "/dev/null",
		"-health-check", "none", "-health-check-interval", "10", "-health-check-timeout", "20"}
	status := addServer.Run(args)
	assert.Equal(t, 1, status)
	assert.True(t, strings.Contains(writer.String(), "health check"))
	assert.True(t, strings.Contains(writer.String(), "less than"))
}

func TestAddServerSynopsis(t *testing.T) {
	_, addServer := testMakeAddServer(false)
	synopsis := addServer.Synopsis()
	assert.NotEqual(t, "", synopsis)
	assert.True(t, strings.ToUpper(synopsis)[0] == synopsis[0], "Synopsis must start with an uppercase char")
}

func TestAddServerCustomHealthcheck(t *testing.T) {

	_, addServer := testMakeAddServer(false)

	args := []string{"-address", "an-address", "-port", "42", "-name", "test-name", "-ping-uri", "/dev/null",
		"-health-check", "custom-http", "-health-check-interval", "42", "-health-check-timeout", "10"}
	status := addServer.Run(args)
	assert.Equal(t, 0, status)
	storedBytes, err := addServer.KVStore.Get("servers/test-name")
	assert.Nil(t, err)
	if assert.NotNil(t, storedBytes) {

		s := config.JSONToServer(storedBytes)
		assert.Equal(t, "test-name", s.Name)
		assert.Equal(t, "an-address", s.Address)
		assert.Equal(t, 42, s.Port)
		assert.Equal(t, "/dev/null", s.PingURI)
		assert.Equal(t, "custom-http", s.HealthCheck)
		assert.Equal(t, 42, s.HealthCheckInterval)
		assert.Equal(t, 10, s.HealthCheckTimeout)
	}
}
