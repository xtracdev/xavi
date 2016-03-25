package service

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"github.com/xtracdev/xavi/plugin"
	"github.com/xtracdev/xavi/plugin/logging"
)

func initKVStore(t *testing.T) kvstore.KVStore {
	kvs, _ := kvstore.NewHashKVStore("")
	loadTestConfig1(kvs, t)
	loadConfigTwoBackendsNoPluginNameSpecified(kvs, t)
	loadMultiRoute(kvs, t)
	loadRouteWithNoBackends(kvs, t)
	loadTLSOnlyBackend(kvs, t)
	loadTLSOnlyBackendBadCert(kvs,t)
	return kvs
}

func TestServerFactory(t *testing.T) {
	plugin.RegisterWrapperFactory("Logging", logging.NewLoggingWrapper)

	var testKVS = initKVStore(t)
	service, err := BuildServiceForListener("listener", "0.0.0.0:8000", testKVS)
	assert.Nil(t, err)
	s := fmt.Sprintf("%s", service)
	println(s)
	assert.True(t, strings.Contains(s, "route1"))
	assert.True(t, strings.Contains(s, "hello-backend"))
	assert.True(t, strings.Contains(s, "listener"))
	assert.True(t, strings.Contains(s, "0.0.0.0:8000"))
}

func TestServerFactoryWithUnknownListener(t *testing.T) {
	var testKVS = initKVStore(t)
	_, err := BuildServiceForListener("no such listener", "0.0.0.0:8000", testKVS)
	assert.NotNil(t, err)
}

func loadTestConfig1(kvs kvstore.KVStore, t *testing.T) {
	ln := &config.ListenerConfig{"listener", []string{"route1"}}
	err := ln.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	serverConfig1 := &config.ServerConfig{"server1", "localhost", 3000, "/hello", "none", 0, 0}
	err = serverConfig1.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	serverConfig2 := &config.ServerConfig{"server2", "localhost", 3100, "/hello", "none", 0, 0}
	err = serverConfig2.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	r := &config.RouteConfig{
		Name:     "route1",
		URIRoot:  "/hello",
		Backends: []string{"hello-backend"},
		Plugins:  []string{"Logging"},
		MsgProps: "",
	}
	err = r.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	b := &config.BackendConfig{
		Name:        "hello-backend",
		ServerNames: []string{"server1", "server2"},
	}
	err = b.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}
}

func loadConfigTwoBackendsNoPluginNameSpecified(kvs kvstore.KVStore, t *testing.T) {
	ln := &config.ListenerConfig{"l1", []string{"r1"}}
	err := ln.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	serverConfig1 := &config.ServerConfig{"s1", "localhost", 3000, "/hello", "none", 0, 0}
	err = serverConfig1.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	serverConfig2 := &config.ServerConfig{"s2", "localhost", 3100, "/hello", "none", 0, 0}
	err = serverConfig2.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	r := &config.RouteConfig{
		Name:     "r1",
		URIRoot:  "/hello",
		Backends: []string{"be1", "be2"},
		Plugins:  []string{"Logging"},
		MsgProps: "",
	}
	err = r.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	be1 := &config.BackendConfig{
		Name:        "be1",
		ServerNames: []string{"server1", "server2"},
	}
	err = be1.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	be2 := &config.BackendConfig{
		Name:        "be2",
		ServerNames: []string{"server1", "server2"},
	}
	err = be2.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}
}

func loadMultiRoute(kvs kvstore.KVStore, t *testing.T) {
	plugin.RegisterWrapperFactory("Logging", logging.NewLoggingWrapper)
	ln := &config.ListenerConfig{"l2", []string{"r2"}}
	err := ln.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	r := &config.RouteConfig{
		Name:                "r2",
		URIRoot:             "/hello",
		Backends:            []string{"be1", "be2"},
		Plugins:             []string{"Logging"},
		MsgProps:            "",
		MultiBackendAdapter: "test-plugin",
	}
	err = r.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}
}

func loadRouteWithNoBackends(kvs kvstore.KVStore, t *testing.T) {
	plugin.RegisterWrapperFactory("Logging", logging.NewLoggingWrapper)
	ln := &config.ListenerConfig{"l2", []string{"r2"}}
	err := ln.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	r := &config.RouteConfig{
		Name:                "r3",
		URIRoot:             "/hello",
		Backends:            []string{},
		Plugins:             []string{"Logging"},
		MsgProps:            "",
		MultiBackendAdapter: "test-plugin",
	}
	err = r.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}
}

func loadTLSOnlyBackend(kvs kvstore.KVStore, t *testing.T) {
	be1 := &config.BackendConfig{
		Name:        "be-tls",
		ServerNames: []string{"server1", "server2"},
		TLSOnly:     true,
		CACertPath:  "./cert.pem",
	}
	err := be1.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}
}

func loadTLSOnlyBackendBadCert(kvs kvstore.KVStore, t *testing.T) {
	be1 := &config.BackendConfig{
		Name:        "be-tls-bogus-cert",
		ServerNames: []string{"server1", "server2"},
		TLSOnly:     true,
		CACertPath:  "./badcert.pem",
	}
	err := be1.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}
}
