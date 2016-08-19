package config

import (
	"errors"
	"github.com/xtracdev/xavi/kvstore"
	"net/http"
)

//HealthCheckFn defines the signature of custom health checks
type HealthCheckFn func(string, *http.Transport) <-chan bool

var customHealthChecks map[string]HealthCheckFn
var ErrNoSuchServer = errors.New("Server definition not found")
var ErrNoHealthCheckFn = errors.New("No health check function provided")

func init() {
	customHealthChecks = make(map[string]HealthCheckFn)
}

func HealthCheckForServer(server string) HealthCheckFn {
	return customHealthChecks[server]
}

func RegisterHealthCheckForServer(kvs kvstore.KVStore, server string, hcfn HealthCheckFn) error {
	//Must register something if this is called.
	if hcfn == nil {
		return ErrNoHealthCheckFn
	}

	//Look up the server
	sc, err := ReadServerConfig(server, kvs)
	if err != nil {
		return err
	}

	if sc == nil {
		return ErrNoSuchServer
	}

	customHealthChecks[server] = hcfn
	return nil
}

func RegisterHealthCheckForBackend(kvs kvstore.KVStore, backend string, hcfn HealthCheckFn) error {
	return nil
}
