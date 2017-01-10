package config

import (
	"github.com/xtracdev/xavi/kvstore"
	"github.com/xtracdev/xavi/plugin"
	"github.com/xtracdev/xavi/plugin/logging"
	"testing"
)

func BuildKVStoreTestConfig(t *testing.T) kvstore.KVStore {
	kvs, _ := kvstore.NewHashKVStore("")
	loadTestConfig1(kvs, t)
	loadConfigTwoBackendsNoPluginNameSpecified(kvs, t)
	loadMultiRoute(kvs, t)
	loadRouteWithNoBackends(kvs, t)
	loadTLSOnlyBackend(kvs, t)
	loadTLSOnlyBackendBadCert(kvs, t)
	return kvs
}

func loadTestConfig1(kvs kvstore.KVStore, t *testing.T) {
	ln := &ListenerConfig{"listener", []string{"route1"}, true}
	err := ln.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	serverConfig1 := &ServerConfig{"server1", "localhost", 3000, "/hello", "none", 0, 0}
	err = serverConfig1.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	serverConfig2 := &ServerConfig{"server2", "localhost", 3100, "/hello", "none", 0, 0}
	err = serverConfig2.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	r := &RouteConfig{
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

	b := &BackendConfig{
		Name:        "hello-backend",
		ServerNames: []string{"server1", "server2"},
	}
	err = b.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}
}

func loadConfigTwoBackendsNoPluginNameSpecified(kvs kvstore.KVStore, t *testing.T) {
	ln := &ListenerConfig{"l1", []string{"r1"}, true}
	err := ln.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	serverConfig1 := &ServerConfig{"s1", "localhost", 3000, "/hello", "none", 0, 0}
	err = serverConfig1.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	serverConfig2 := &ServerConfig{"s2", "localhost", 3100, "/hello", "none", 0, 0}
	err = serverConfig2.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	r := &RouteConfig{
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

	be1 := &BackendConfig{
		Name:        "be1",
		ServerNames: []string{"server1", "server2"},
	}
	err = be1.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	be2 := &BackendConfig{
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
	ln := &ListenerConfig{"l2", []string{"r2"}, true}
	err := ln.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	r := &RouteConfig{
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
	ln := &ListenerConfig{"l2", []string{"r2"}, true}
	err := ln.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	r := &RouteConfig{
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
	be1 := &BackendConfig{
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
	be1 := &BackendConfig{
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
