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
var ErrNoSuchBackend = errors.New("Backend definition not found")

func init() {
	customHealthChecks = make(map[string]HealthCheckFn)
}

//HealthCheckForServer returns the custom health check function associated with a server
func HealthCheckForServer(server string) HealthCheckFn {
	return customHealthChecks[server]
}

//RegisterHealthCheckForServer registers a custom health check function for a given server. The
//configuration store is check for the existance of the specified server definition prior to
//storing the health check function.
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

//RegisterHealthCheckForBackend registers the given health check for every server associated with the
//given backend
func RegisterHealthCheckForBackend(kvs kvstore.KVStore, backend string, hcfn HealthCheckFn) error {
	//Must register something if this is called.
	if hcfn == nil {
		return ErrNoHealthCheckFn
	}

	//Look up the backend
	be, err := ReadBackendConfig(backend, kvs)
	if err != nil {
		return err
	}

	if be == nil {
		return ErrNoSuchBackend
	}

	//Go through the server definitions associated with the backend
	for _, s := range be.ServerNames {
		err := RegisterHealthCheckForServer(kvs, s, hcfn)
		if err != nil {
			return err
		}
	}

	return nil
}
