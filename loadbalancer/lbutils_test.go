package loadbalancer

import (
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"testing"
	"net/http/httptest"
	"net/http"
	"net/url"
	"github.com/xtracdev/xavi/kvstore"
	"net"
	"context"
	"io/ioutil"
	"strconv"
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

func buildTestConfigForLBCall(t *testing.T, urlStr string)kvstore.KVStore {
	kvs, _ := kvstore.NewHashKVStore("")

	url,_ := url.Parse(urlStr)
	host,port,err := net.SplitHostPort(url.Host)
	assert.Nil(t,err)

	portVal,err := strconv.Atoi(port)
	assert.Nil(t,err)

	ln := &config.ListenerConfig{"lbclistener", []string{"lbcroute1"}}
	err = ln.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	serverConfig1 := &config.ServerConfig{"lbcserver1", host, portVal, "/hello", "none", 0, 0}
	err = serverConfig1.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

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
		ServerNames: []string{"lbcserver1"},
	}
	err = b.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	return kvs
}

func TestLBUtilsCallSvc(t *testing.T) {


	serverResp := "Hello, client"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(serverResp))
	}))
	defer ts.Close()

	kvs := buildTestConfigForLBCall(t, ts.URL)
	sc, err := config.ReadServiceConfig("lbclistener", kvs)
	assert.Nil(t, err)
	assert.NotNil(t, sc)

	config.RecordActiveConfig(sc)

	lb, err := NewBackendLoadBalancer("lbcbackend")
	assert.Nil(t, err)

	req,err := http.NewRequest("GET","/foo",nil)
	assert.Nil(t,err)
	resp, err := lb.DoWithLoadbalancer(context.Background(), req, false)
	if assert.Nil(t, err) {
		defer resp.Body.Close()
		b,err := ioutil.ReadAll(resp.Body)
		assert.Nil(t,err)
		assert.Equal(t, serverResp,string(b))
	}
}
