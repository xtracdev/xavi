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

	r := &config.RouteConfig{"route1", "/hello", "hello-backend", []string{"Logging"}, ""}
	err = r.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	b := &config.BackendConfig{"hello-backend", []string{"server1", "server2"}, ""}
	err = b.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

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
