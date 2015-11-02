package loadbalancer

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/config"
	"net/http"
	"time"
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
	default:
		return false
	}
}

//KnownHealthChecks returns the names of the health checks supported bt the toolkit
func KnownHealthChecks() string {
	return "none, http-get"
}

func healthy(endpoint string) <-chan bool {
	statusChannel := make(chan bool)
	go func() {

		resp, err := http.Get(endpoint)
		if err != nil {
			log.Warn("Error doing get on healthcheck endpoint ", endpoint, " : ", err.Error())
			statusChannel <- false
			return
		}

		if resp == nil {
			log.Warn("nil response from health check endpoint")
			statusChannel <- false
			return
		}

		statusChannel <- resp.StatusCode == 200
	}()

	return statusChannel
}

func httpGet(lbEndpoint *LoadBalancerEndpoint, serverConfig config.ServerConfig, loop bool) func() {

	url := fmt.Sprintf("http://%s:%d%s", serverConfig.Address, serverConfig.Port, serverConfig.PingURI)
	log.Info("Setting healthcheck url to ", url)
	healthCheckInterval := time.Duration(serverConfig.HealthCheckInterval) * time.Millisecond

	return func() {
		for {
			time.Sleep(healthCheckInterval)
			log.Debug("checking health")
			select {
			case healthStatus := <-healthy(url):
				if !healthStatus {
					log.Warn("Endpoint ", serverConfig.Address, ":", serverConfig.Port, " is not healthy")
					lbEndpoint.Up = false
				} else {
					log.Debug("Endpoint is up: ", serverConfig.Address, ":", serverConfig.Port)
					lbEndpoint.Up = true
				}

			case <-time.After(time.Duration(serverConfig.HealthCheckTimeout) * time.Millisecond):
				log.Warn("Health check timed out - marking endpoint ", serverConfig.Address, ":", serverConfig.Port, " as not healthy")
				lbEndpoint.Up = false
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
	switch serverConfig.HealthCheck {
	default:
		log.Debug("returning no-op health check")
		return noop
	case "http-get":
		log.Debug("returning http-get health check")
		return httpGet(lbEndpoint, serverConfig, loop)
	}
}
