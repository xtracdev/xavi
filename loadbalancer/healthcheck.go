package loadbalancer

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"crypto/tls"
	"crypto/x509"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/config"
)

//DefaultHealthCheckInterval is the time between health checks if no value is specified by the configuration
const DefaultHealthCheckInterval = 30 * 1000 //30 seconds

//DefaultHealthCheckTimeout is the timeout for the health check if no value is specified by the configuration
const DefaultHealthCheckTimeout = 10 * 1000 //10 seconds

//IsKnownHealthCheck returns true for the health check types supported by the toolkit
func IsKnownHealthCheck(healthcheck string) bool {
	switch healthcheck {
	case "none":
		return true
	case "http-get":
		return true
	case "https-get":
		return true
	case "custom-http":
		return true
	case "custom-https":
		return true
	default:
		return false
	}
}

//KnownHealthChecks returns the names of the health checks supported bt the toolkit
func KnownHealthChecks() string {
	return "none, http-get, https-get, custom-http, custom-https"
}

func createHealthCheckFnWithTimeout(healthCheckTimeout time.Duration) config.HealthCheckFn {
	return func(endpoint string, transport *http.Transport) <-chan bool {
		statusChannel := make(chan bool)

		client := &http.Client{
			Transport: transport,
			Timeout:   healthCheckTimeout,
		}

		go func() {

			resp, err := client.Get(endpoint)
			if err != nil {
				log.Warn("Error doing get on healthcheck endpoint ", endpoint, " : ", err.Error())

				//Check to see if there's a non-nil response: drain it if present
				if resp != nil {
					log.Info("clean up non-nil response delivered with health check client error")
					defer resp.Body.Close()
					b, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Infof("Error reading resp while cleaning up after error: %v\n", err)
					} else {
						log.Infof("Discarded response body after handling error on healtcheck get: %s\n", b)
					}
				}

				statusChannel <- false
				return
			}

			//Read the entire response and close the body to ensure proper connection hygiene. On the mac you
			//can use something like lsof | grep xavi|wc -l  (and check/timeout values
			//of 5000/2000 ms respectively) to see file handles in use - without the close and read the
			//connections in grow without being released.
			defer resp.Body.Close()
			_, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Warnf("Error reading health check response: %v", err)
				statusChannel <- false
				return
			}

			statusChannel <- resp.StatusCode == 200
		}()

		return statusChannel
	}
}

func makeCertPool(caCertPath string) *x509.CertPool {
	pool := x509.NewCertPool()

	pemData, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Warn("Error creating CA Cert Poll for health check: ", err.Error())
		return nil
	}

	ok := pool.AppendCertsFromPEM(pemData)
	if !ok {
		log.Warn("Error append pem data to cert pool for health check: ", err.Error())
		return nil
	}

	return pool
}

func makeTransportForHealthCheck(https bool, caCertPath string) *http.Transport {
	defaultTransport := &http.Transport{DisableKeepAlives: false, DisableCompression: false, Proxy: http.ProxyFromEnvironment}
	//Non-https case
	if https == false {
		return defaultTransport
	}

	if caCertPath == "" {
		log.Info("Using default transport for https health check - will work only for known CAs")
		log.Info("For self signed certs specify -cacert-path in your backend configuration.")
		return defaultTransport
	}

	pool := makeCertPool(caCertPath)
	if pool == nil {
		log.Warn("Unable to create cert pool based on configuration - using default transport")
		return defaultTransport
	}

	log.Info("using custom transport for health check")
	tlsConfig := &tls.Config{RootCAs: pool}
	return &http.Transport{DisableKeepAlives: false, DisableCompression: false, Proxy: http.ProxyFromEnvironment, TLSClientConfig: tlsConfig}

}

func httpGet(lbEndpoint *LoadBalancerEndpoint, serverConfig config.ServerConfig, loop bool, https bool, hcfn config.HealthCheckFn) func() {

	var url string
	transport := makeTransportForHealthCheck(https, lbEndpoint.CACertPath)
	if https {
		url = fmt.Sprintf("https://%s:%d%s", serverConfig.Address, serverConfig.Port, serverConfig.PingURI)
	} else {
		url = fmt.Sprintf("http://%s:%d%s", serverConfig.Address, serverConfig.Port, serverConfig.PingURI)
	}

	log.Info("Setting healthcheck url to ", url)
	healthCheckInterval := time.Duration(serverConfig.HealthCheckInterval) * time.Millisecond

	return func() {
		for {
			time.Sleep(healthCheckInterval)
			log.Debug("checking health")
			select {
			case healthStatus := <-hcfn(url, transport):
				if !healthStatus {
					log.Warn("Endpoint ", serverConfig.Address, ":", serverConfig.Port, " is not healthy")
					lbEndpoint.MarkLoadBalancerEndpointUp(false)
				} else {
					log.Debug("Endpoint is up: ", serverConfig.Address, ":", serverConfig.Port)
					lbEndpoint.MarkLoadBalancerEndpointUp(true)
				}

			case <-time.After(time.Duration(serverConfig.HealthCheckTimeout) * time.Millisecond):
				log.Warn("Health check timed out - marking endpoint ", serverConfig.Address, ":", serverConfig.Port, " as not healthy")
				lbEndpoint.MarkLoadBalancerEndpointUp(false)
			}

			if loop == false {
				return
			}
		}
	}
}

func noop() {}

//MakeHealthCheck returns a health check function based on the server configuration and load balancer endpoint. The
//loop arguement is meant to enable testability - normal health check functions run until the listener is shutdown,
//unit test health checks run once typically.
func MakeHealthCheck(lbEndpoint *LoadBalancerEndpoint, serverConfig config.ServerConfig, loop bool) func() {
	log.Debugf("Making health check for %s", serverConfig.Name)
	switch serverConfig.HealthCheck {
	default:
		log.Debug("returning no-op health check")
		return noop
	case "http-get":
		log.Debug("returning http-get health check")
		healthCheckTimeout := time.Duration(DefaultHealthCheckTimeout)
		if serverConfig.HealthCheckTimeout > 0 {
			healthCheckTimeout = time.Duration(serverConfig.HealthCheckTimeout) * time.Millisecond
		}
		return httpGet(lbEndpoint, serverConfig, loop, false,
			createHealthCheckFnWithTimeout(healthCheckTimeout))
	case "https-get":
		log.Debug("returning https-get health check")
		healthCheckTimeout := time.Duration(DefaultHealthCheckTimeout)
		if serverConfig.HealthCheckTimeout > 0 {
			healthCheckTimeout = time.Duration(serverConfig.HealthCheckTimeout) * time.Millisecond
		}
		return httpGet(lbEndpoint, serverConfig, loop, true,
			createHealthCheckFnWithTimeout(healthCheckTimeout))
	case "custom-http":
		log.Debug("returning custom http-get health check")
		hcfn := config.HealthCheckForServer(serverConfig.Name)
		if hcfn == nil {
			log.Fatalf("No custom health check registered for %s - add code to register healthcheck or change config",
				serverConfig.Name)
		}
		return httpGet(lbEndpoint, serverConfig, loop, false, hcfn)
	case "custom-https":
		log.Debug("returning custom https-get health check")
		hcfn := config.HealthCheckForServer(serverConfig.Name)
		if hcfn == nil {
			log.Fatalf("No custom health check registered for %s - add code to register healthcheck or change config",
				serverConfig.Name)
		}
		return httpGet(lbEndpoint, serverConfig, loop, true, hcfn)
	}
}
