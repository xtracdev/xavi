package loadbalancer

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/config"
	"net/http"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"golang.org/x/net/context/ctxhttp"
	"golang.org/x/net/context"
)

type BackendLoadBalancer struct {
	LoadBalancer LoadBalancer
	BackendConfig *config.BackendConfig
	CertPool       *x509.CertPool
}

var ErrBackendNotFound = errors.New("Given backed end not found in active listener config")

func findBackend(backend string) (*config.ServiceBackend, error) {
	for _, listenerName := range config.ActiveListenerNames() {
		sc := config.ActiveConfigForListener(listenerName)

		for _, r := range sc.Routes {
			log.Infof("route config for %s:", r.Route.Name)
			for _, b := range r.Backends {
				if b.Backend.Name == backend {
					return b, nil
				}
			}
		}
	}

	return nil, ErrBackendNotFound
}

func serversForBackend(backend *config.ServiceBackend) []config.ServerConfig {
	servers := make([]config.ServerConfig, 0)

	if backend == nil {
		return servers
	}

	for _, s := range backend.Servers {
		log.Infof("server config for %s:", s.Name)
		servers = append(servers, *s)
	}

	return servers
}

func createCertPool(backendConfig *config.BackendConfig) (*x509.CertPool, error) {
	if backendConfig.CACertPath == "" {
		return nil, nil
	}

	log.Debug("Creating cert pool for backend ", backendConfig.Name)

	pool := x509.NewCertPool()

	pemData, err := ioutil.ReadFile(backendConfig.CACertPath)
	if err != nil {
		return nil, err
	}

	ok := pool.AppendCertsFromPEM(pemData)
	if !ok {
		return nil, errors.New("Error appending certs from pem data")
	}

	return pool, nil
}

// NewLoadBalancer instantiates a load balancer based on the named backend configuration. Backend
// names are scoped to routes, thus the route is given to ensure the correct backend is returned
// if multiple backend definitions with the same name are given.
func NewBackendLoadBalancer(backendName string) (*BackendLoadBalancer, error) {
	backend, err := findBackend(backendName)
	if err != nil {
		return nil, err
	}

	servers := serversForBackend(backend)

	backendConfig := backend.Backend
	factory := ObtainFactoryForLoadBalancer(backendConfig.LoadBalancerPolicy)
	if factory == nil {
		factory = new(RoundRobinLoadBalancerFactory)
	}

	certPool, err := createCertPool(backendConfig)
	if err != nil {
		return nil,err
	}

	lb, err := factory.NewLoadBalancer(backendConfig.Name, backendConfig.CACertPath, servers)

	return &BackendLoadBalancer{LoadBalancer:lb,BackendConfig:backendConfig, CertPool: certPool}, err
}




func (lb *BackendLoadBalancer) DoWithLoadbalancer(ctx context.Context, req *http.Request, useTLS bool)(*http.Response,error) {
	connectString,err := lb.LoadBalancer.GetConnectAddress()
	if err != nil {
		return nil,err
	}


	log.Debug("connect string is ", connectString)
	req.URL.Host = connectString
	req.Host = connectString

	var transport *http.Transport
	if useTLS == true {
		tlsConfig := &tls.Config{RootCAs: lb.CertPool}
		transport = &http.Transport{DisableKeepAlives: false, DisableCompression: false, TLSClientConfig: tlsConfig}
		req.URL.Scheme = "https"
	} else {
		transport = &http.Transport{DisableKeepAlives: false, DisableCompression: false}
		req.URL.Scheme = "http"
	}

	client := &http.Client{
		Transport: transport,
	}

	req.RequestURI = "" //Must clear when using http.Client
	return ctxhttp.Do(ctx, client, req)
}
