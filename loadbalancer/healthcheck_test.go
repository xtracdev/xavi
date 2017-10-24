package loadbalancer

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"io/ioutil"
	"os"
)

var standardTransport = &http.Transport{DisableKeepAlives: false, DisableCompression: false}

var healthy = createHealthCheckFnWithTimeout(500 * time.Millisecond)

func TestHCIsKnownHealthCheck(t *testing.T) {
	assert.True(t, IsKnownHealthCheck("none"))
	assert.True(t, IsKnownHealthCheck("http-get"))
	assert.False(t, IsKnownHealthCheck("code-monkey"))
}

func TestHCKnownHealthChecks(t *testing.T) {
	healthChecks := KnownHealthChecks()
	assert.NotEmpty(t, healthChecks)
	assert.True(t, strings.Contains(healthChecks, "none"))
	assert.True(t, strings.Contains(healthChecks, "http-get"))
}

func TestHCHealthy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	healthChan := healthy(ts.URL+"/foo", standardTransport)

	select {
	case status := <-healthChan:
		assert.True(t, status)
	case <-time.After(100 * time.Millisecond):
		t.Fail()
	}

	healthChan = healthy("http://localhost:666/foo", standardTransport)
	select {
	case status := <-healthChan:
		assert.False(t, status)
	case <-time.After(100 * time.Millisecond):
		t.Fail()
	}

}

func TestHCCustomHealthy(t *testing.T) {
	var called bool
	var customerHeaderPresent bool

	//Test server ping handler
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
		called = true
		headerVal := r.Header.Get("CustomerHeader")
		if headerVal == "XXX" {
			customerHeaderPresent = true
		}
	}))

	lbEndpoint := new(LoadBalancerEndpoint)
	lbEndpoint.Address = ts.URL
	lbEndpoint.PingURI = "/foo"
	lbEndpoint.Up = false

	testURL, err := url.Parse(ts.URL)
	assert.Nil(t, err)
	_, portStr, err := net.SplitHostPort(testURL.Host)
	assert.Nil(t, err)

	port, err := strconv.Atoi(portStr)

	kvs, _ := kvstore.NewHashKVStore("")

	serverConfig := config.ServerConfig{
		Name:                "server1",
		Address:             "localhost",
		Port:                port,
		PingURI:             "/foo",
		HealthCheck:         "custom-http",
		HealthCheckInterval: 200,
		HealthCheckTimeout:  100,
	}

	err = serverConfig.Store(kvs)
	if err != nil {
		t.Fatal(err)
	}

	//Register custom health check
	hcfn := func(endpoint string, transport *http.Transport) <-chan bool {
		statusChannel := make(chan bool)

		client := &http.Client{
			Transport: transport,
		}

		go func() {
			logrus.Infof("Custom health check alive, endpoint %s", endpoint)

			req, _ := http.NewRequest("GET", endpoint, nil)
			req.Header.Add("CustomerHeader", "XXX")
			resp, err := client.Do(req)
			if err != nil {
				logrus.Warnf("Error getting endpoint: %s", err.Error())
				statusChannel <- false
				return
			}

			defer resp.Body.Close()
			ioutil.ReadAll(resp.Body)

			if resp == nil {
				statusChannel <- false
				return
			}

			statusChannel <- resp.StatusCode == 200
		}()

		return statusChannel
	}

	config.ListenContext = true
	config.RegisterHealthCheckForServer(kvs, "server1", hcfn)
	config.ListenContext = false

	//Create the healthcheck function and invoke it
	healthcheckFn := MakeHealthCheck(lbEndpoint, serverConfig, false)
	healthcheckFn()

	assert.True(t, called)
	assert.True(t, lbEndpoint.Up)

}

func TestHCHealthyTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	healthChan := healthy(ts.URL+"/foo", standardTransport)

	select {
	case status := <-healthChan:
		t.Log("Got something from the healthy channel")
		assert.False(t, status)
		t.Fail()
	case <-time.After(100 * time.Millisecond):
		t.Log("timed out as expected")
	}
}

func TestHCMakeHealthCheck(t *testing.T) {
	var called = false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	lbEndpoint := new(LoadBalancerEndpoint)
	lbEndpoint.Address = ts.URL
	lbEndpoint.PingURI = "/foo"
	lbEndpoint.Up = false

	testURL, err := url.Parse(ts.URL)
	assert.Nil(t, err)
	_, portStr, err := net.SplitHostPort(testURL.Host)
	assert.Nil(t, err)

	port, err := strconv.Atoi(portStr)

	serverConfig := config.ServerConfig{
		Name:                "testcfg",
		Address:             "localhost",
		Port:                port,
		PingURI:             "/foo",
		HealthCheck:         "http-get",
		HealthCheckInterval: 200,
		HealthCheckTimeout:  100,
	}

	healthcheckFn := MakeHealthCheck(lbEndpoint, serverConfig, false)
	healthcheckFn()

	assert.True(t, called)
	assert.True(t, lbEndpoint.Up)

}

func TestHCMakeHealthCheckUnhealthy(t *testing.T) {
	var called = false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	lbEndpoint := new(LoadBalancerEndpoint)
	lbEndpoint.Address = ts.URL
	lbEndpoint.PingURI = "/foo"
	lbEndpoint.Up = false

	testURL, err := url.Parse(ts.URL)
	assert.Nil(t, err)
	_, portStr, err := net.SplitHostPort(testURL.Host)
	assert.Nil(t, err)

	port, err := strconv.Atoi(portStr)

	serverConfig := config.ServerConfig{
		Name:                "testcfg",
		Address:             "localhost",
		Port:                port,
		PingURI:             "/foo",
		HealthCheck:         "http-get",
		HealthCheckInterval: 200,
		HealthCheckTimeout:  100,
	}

	healthcheckFn := MakeHealthCheck(lbEndpoint, serverConfig, false)
	healthcheckFn()

	assert.True(t, called)
	assert.False(t, lbEndpoint.Up)

}

func TestHCMakeHealthCheckTimeout(t *testing.T) {
	var called = false
	var wg sync.WaitGroup

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wg.Add(1)
		go func() {
			called = true
			defer wg.Done()
		}()
		time.Sleep(200 * time.Millisecond)
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	lbEndpoint := new(LoadBalancerEndpoint)
	lbEndpoint.Address = ts.URL
	lbEndpoint.PingURI = "/foo"
	lbEndpoint.Up = false

	testURL, err := url.Parse(ts.URL)
	assert.Nil(t, err)
	_, portStr, err := net.SplitHostPort(testURL.Host)
	assert.Nil(t, err)

	port, err := strconv.Atoi(portStr)

	serverConfig := config.ServerConfig{
		Name:                "testcfg",
		Address:             "localhost",
		Port:                port,
		PingURI:             "/foo",
		HealthCheck:         "http-get",
		HealthCheckInterval: 100,
		HealthCheckTimeout:  100,
	}

	healthcheckFn := MakeHealthCheck(lbEndpoint, serverConfig, false)
	healthcheckFn()

	wg.Wait()
	assert.True(t, called)

}

func TestMakeTransportForHealthCheckNoProxySet(t *testing.T) {
	defer os.Setenv("http_proxy", os.Getenv("http_proxy"))
	defer os.Setenv("https_proxy", os.Getenv("https_proxy"))
	defer os.Setenv("no_proxy", os.Getenv("no_proxy"))
	defer os.Setenv("HTTP_PROXY", os.Getenv("HTTP_PROXY"))
	defer os.Setenv("HTTPS_PROXY", os.Getenv("HTTPS_PROXY"))
	defer os.Setenv("NO_PROXY", os.Getenv("NO_PROXY"))

	os.Unsetenv("http_proxy")
	os.Unsetenv("https_proxy")
	os.Unsetenv("no_proxy")
	os.Unsetenv("HTTP_PROXY")
	os.Unsetenv("HTTPS_PROXY")
	os.Unsetenv("NO_PROXY")

	req, err := http.NewRequest(http.MethodGet, "https://www.test.com", nil)
	assert.Nil(t, err)

	transport := makeTransportForHealthCheck(false, "cert")
	proxy, err := transport.Proxy(req)
	assert.Nil(t, err)
	assert.Nil(t, proxy)
}

func TestMakeTransportForHealthCheckProxyIgnored(t *testing.T) {
	defer os.Setenv("http_proxy", os.Getenv("http_proxy"))
	defer os.Setenv("https_proxy", os.Getenv("https_proxy"))
	defer os.Setenv("no_proxy", os.Getenv("no_proxy"))
	defer os.Setenv("HTTP_PROXY", os.Getenv("HTTP_PROXY"))
	defer os.Setenv("HTTPS_PROXY", os.Getenv("HTTPS_PROXY"))
	defer os.Setenv("NO_PROXY", os.Getenv("NO_PROXY"))

	os.Unsetenv("http_proxy")
	os.Unsetenv("https_proxy")
	os.Unsetenv("no_proxy")
	os.Unsetenv("HTTP_PROXY")
	os.Unsetenv("HTTPS_PROXY")
	os.Unsetenv("NO_PROXY")

	os.Setenv("HTTPS_PROXY", "https://proxy.com")
	os.Setenv("no_proxy", "test.com")

	req, err := http.NewRequest(http.MethodGet, "https://test.com", nil)
	assert.Nil(t, err)

	transport := makeTransportForHealthCheck(false, "cert")
	proxy, err := transport.Proxy(req)
	assert.Nil(t, err)
	assert.Nil(t, proxy)
}

func TestMakeTransportForHealthCheckProxyUsed(t *testing.T) {
	defer os.Setenv("http_proxy", os.Getenv("http_proxy"))
	defer os.Setenv("https_proxy", os.Getenv("https_proxy"))
	defer os.Setenv("no_proxy", os.Getenv("no_proxy"))
	defer os.Setenv("HTTP_PROXY", os.Getenv("HTTP_PROXY"))
	defer os.Setenv("HTTPS_PROXY", os.Getenv("HTTPS_PROXY"))
	defer os.Setenv("NO_PROXY", os.Getenv("NO_PROXY"))

	os.Unsetenv("http_proxy")
	os.Unsetenv("https_proxy")
	os.Unsetenv("no_proxy")
	os.Unsetenv("HTTP_PROXY")
	os.Unsetenv("HTTPS_PROXY")
	os.Unsetenv("NO_PROXY")

	os.Setenv("http_proxy", "http://proxy.com")
	os.Setenv("no_proxy", "other.com")

	req, err := http.NewRequest(http.MethodGet, "http://www.test.com", nil)
	assert.Nil(t, err)

	transport := makeTransportForHealthCheck(false, "cert")
	proxy, err := transport.Proxy(req)
	assert.Nil(t, err)
	assert.Equal(t, "proxy.com", proxy.Host)
}
