package loadbalancer

import (
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestLBUtilsBuildFromConfig(t *testing.T) {
	kvs := config.BuildKVStoreTestConfig(t)
	assert.NotNil(t, kvs)
	sc, err := config.ReadServiceConfig("listener", kvs)
	assert.Nil(t, err)
	assert.NotNil(t, sc)

	config.RecordActiveConfig(sc)

	lb, err := NewBackendLoadBalancer("hello-backend")
	assert.Nil(t, err)
	assert.NotNil(t, lb)

	assert.Equal(t, "hello-backend", lb.BackendConfig.Name)
	assert.Equal(t, "", lb.BackendConfig.CACertPath)
	assert.Equal(t, 2, len(lb.BackendConfig.ServerNames))

	h, _ := lb.LoadBalancer.GetEndpoints()
	if assert.True(t, len(h) == 2) {
		assert.Equal(t, "localhost:3000", h[0])
		assert.Equal(t, "localhost:3100", h[1])
	}
}

func TestLBUtilsNoSuchBackend(t *testing.T) {
	kvs := config.BuildKVStoreTestConfig(t)
	assert.NotNil(t, kvs)
	sc, err := config.ReadServiceConfig("listener", kvs)
	assert.Nil(t, err)
	assert.NotNil(t, sc)

	config.RecordActiveConfig(sc)

	lb, err := NewBackendLoadBalancer("no-such-backed")
	assert.Nil(t, lb)
	assert.NotNil(t, err)
	assert.Equal(t, ErrBackendNotFound, err)
}

func buildTestConfigForLBCall(t *testing.T, server1Url, server2Url string) kvstore.KVStore {
	kvs, _ := kvstore.NewHashKVStore("")

	//Define listener
	ln := &config.ListenerConfig{"lbclistener", []string{"lbcroute1"}, true}
	err := ln.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	//Define server 1
	url, _ := url.Parse(server1Url)
	host, port, err := net.SplitHostPort(url.Host)
	assert.Nil(t, err)

	portVal, err := strconv.Atoi(port)
	assert.Nil(t, err)

	serverConfig1 := &config.ServerConfig{"lbcserver1", host, portVal, "/hello", "none", 0, 0}
	err = serverConfig1.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	//Define server 2
	url, _ = url.Parse(server2Url)
	host, port, err = net.SplitHostPort(url.Host)
	assert.Nil(t, err)

	portVal, err = strconv.Atoi(port)
	assert.Nil(t, err)

	serverConfig2 := &config.ServerConfig{"lbcserver2", host, portVal, "/hello", "none", 0, 0}
	err = serverConfig2.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	//Define route
	r := &config.RouteConfig{
		Name:     "lbcroute1",
		URIRoot:  "/hello",
		Backends: []string{"lbcbackend"},
		Plugins:  []string{"Logging"},
		MsgProps: "",
	}
	err = r.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	b := &config.BackendConfig{
		Name:        "lbcbackend",
		ServerNames: []string{"lbcserver1", "lbcserver2"},
	}
	err = b.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	return kvs
}

func TestLBUtilsCallSvc(t *testing.T) {

	serverResp := "Hello, client"
	var server1Called, server2Called bool

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server1Called = true
		w.Write([]byte(serverResp))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server2Called = true
		w.Write([]byte(serverResp))
	}))
	defer server2.Close()

	kvs := buildTestConfigForLBCall(t, server1.URL, server2.URL)
	sc, err := config.ReadServiceConfig("lbclistener", kvs)
	assert.Nil(t, err)
	assert.NotNil(t, sc)

	config.RecordActiveConfig(sc)

	lb, err := NewBackendLoadBalancer("lbcbackend")
	assert.Nil(t, err)

	req, err := http.NewRequest("GET", "/foo", nil)
	assert.Nil(t, err)

	//Call 1
	resp, err := lb.DoWithLoadBalancer(req, false)
	if assert.Nil(t, err) {
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)
		assert.Equal(t, serverResp, string(b))
	}

	//Call 2
	resp, err = lb.DoWithLoadBalancer(req, false)
	if assert.Nil(t, err) {
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)
		assert.Equal(t, serverResp, string(b))
	}

	//Make sure both servers were called
	assert.True(t, server1Called, "Expected server 1 to be called")
	assert.True(t, server2Called, "Expected server 2 to be called")
}
