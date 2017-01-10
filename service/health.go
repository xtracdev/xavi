package service

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/info"
	"net/http"
)

var activeHealthCheckContext map[string]*HealthCheckContext

func init() {
	activeHealthCheckContext = make(map[string]*HealthCheckContext)
}

func ActiveHealthCheckContextForListener(listenerName string) *HealthCheckContext {
	return activeHealthCheckContext[listenerName]
}

func RecordActiveHealthCheckContext(hcc *HealthCheckContext) {
	if hcc == nil || hcc.ListenerName == "" {
		return
	}

	activeHealthCheckContext[hcc.ListenerName] = hcc
}

//HealthResponse is used to structure the json response payload for the health check
type HealthResponse struct {
	ListenerName string         `json:"listenerName"`
	BuildNumber  string         `json:"buildNumber"`
	Routes       []routeContext `json:"routes"`
}

type routeContext struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Up       bool   `json:"up"`
	Backends []backendContext
}

type backendContext struct {
	Name                  string   `json:"name"`
	Up                    bool     `json:"up"`
	HealthyDependencies   []string `json:healthyDependencies`
	UnhealthyDependencies []string `json:unhealthyDependencies`
}

//HealthCheckContext is a type  that is used to supply the context needed to build the
//health check handler and to provide current health status
type HealthCheckContext struct {
	ListenerName         string
	BuildNumber          string
	EnableHealthEndpoint bool
	routes               []*route
}

//HealthHandler returns the handler for health checks
func (hcc *HealthCheckContext) HealthHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(hcc.GetHealthStatus())
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.Write(b)
	}
}

//Return current health status
func (hcc *HealthCheckContext) GetHealthStatus() *HealthResponse {

	hr := &HealthResponse{
		ListenerName: hcc.ListenerName,
		BuildNumber:  info.BuildVersion,
	}

	for _, r := range hcc.routes {

		rc := routeContext{
			Name: r.Name,
			URL:  r.URIRoot,
			Up:   true,
		}

		unhealthyBackends := 0
		for _, b := range r.Backends {
			bctx := backendContext{
				Name: b.Name,
			}
			log.Infof("Backend %s", b.Name)
			h, uh := b.LoadBalancer.GetEndpoints()
			log.Infof("Healthy: %s, unhealthy: %s", h, uh)
			bctx.Up = len(h) > 0
			bctx.HealthyDependencies = h
			bctx.UnhealthyDependencies = uh

			rc.Backends = append(rc.Backends, bctx)

			if !bctx.Up {
				unhealthyBackends++
			}
		}

		if unhealthyBackends > 0 {
			rc.Up = false
		}

		hr.Routes = append(hr.Routes, rc)
	}

	return hr
}

//AddRouteContext adds the needed route context for performing the healthcheck
func (hcc *HealthCheckContext) AddRouteContext(r *route) {
	hcc.routes = append(hcc.routes, r)
}
