package main

import (
	"github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/env"
	"github.com/xtracdev/xavi/plugin"
	"os"
	"testing"
)

func TestPluginRegistration(t *testing.T) {
	registerPlugins()
	assert.True(t, plugin.RegistryContains("Logging"))
}

func TestXapHook(t *testing.T) {
	os.Setenv(env.LoggingOpts, "")
	err := addXapHook()
	assert.Nil(t, err)

	os.Setenv(env.LoggingOpts, env.Tcplog)
	err = addXapHook()
	assert.NotNil(t, err)

	os.Setenv(env.TcplogAddress, "parse this ### yes?")
	err = addXapHook()
	assert.NotNil(t, err)

	os.Setenv(env.TcplogAddress, "0.0.0.0:80")
	err = addXapHook()
	assert.NotNil(t, err)
}

func TestGrabCommandLineArgs(t *testing.T) {
	grabCommandLineArgs()
}
