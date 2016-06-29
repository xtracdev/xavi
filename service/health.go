package service

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/info"
	"net/http"
)

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
//health check handler
type HealthCheckContext struct {
	ListenerName string
	BuildNumber  string
	routes       []*route
}

//HealthHandler returns the handler for health checks
func (hcc *HealthCheckContext) HealthHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		var hr HealthResponse
		hr.ListenerName = hcc.ListenerName
		hr.BuildNumber = info.BuildVersion

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

		b, err := json.Marshal(&hr)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.Write(b)
	}
}

//AddRouteContext adds the needed route context for performing the healthcheck
func (hcc *HealthCheckContext) AddRouteContext(r *route) {
	hcc.routes = append(hcc.routes, r)
}

//HealthResponse is used to structure the json response payload for the health check
type HealthResponse struct {
	ListenerName string         `json:"listenerName"`
	BuildNumber  string         `json:"buildNumber"`
	Routes       []routeContext `json:"routes"`
}
