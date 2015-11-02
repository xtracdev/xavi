package loadbalancer

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestIsKnownHealthCheck(t *testing.T) {
	assert.True(t, IsKnownHealthCheck("none"))
	assert.True(t, IsKnownHealthCheck("http-get"))
	assert.False(t, IsKnownHealthCheck("code-monkey"))
}

func TestKnownHealthChecks(t *testing.T) {
	healthChecks := KnownHealthChecks()
	assert.NotEmpty(t, healthChecks)
	assert.True(t, strings.Contains(healthChecks, "none"))
	assert.True(t, strings.Contains(healthChecks, "http-get"))
}

func TestHealthy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	healthChan := healthy(ts.URL + "/foo")

	select {
	case status := <-healthChan:
		assert.True(t, status)
	case <-time.After(100 * time.Millisecond):
		t.Fail()
	}

	healthChan = healthy("http://localhost:666/foo")
	select {
	case status := <-healthChan:
		assert.False(t, status)
	case <-time.After(100 * time.Millisecond):
		t.Fail()
	}

}

func TestHealthyTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	healthChan := healthy(ts.URL + "/foo")

	select {
	case status := <-healthChan:
		t.Log("Got something from the healthy channel")
		assert.False(t, status)
		t.Fail()
	case <-time.After(100 * time.Millisecond):
		t.Log("timed out as expected")
	}
}

func TestMakeHealthCheck(t *testing.T) {
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

func TestMakeHealthCheckUnhealthy(t *testing.T) {
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

func TestMakeHealthCheckTimeout(t *testing.T) {
	var called = false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
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

	assert.True(t, called)

}
