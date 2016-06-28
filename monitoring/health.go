package monitoring

import (
	"encoding/json"
	"github.com/xtracdev/xavi/info"
	"net/http"
)

//HealthCheckContext is a type  that is used to supply the context needed to build the
//health check handler
type HealthCheckContext struct {
	ListenerName string
	BuildNumber  string
}

//HealthHandler returns the handler for health checks
func (hcc *HealthCheckContext) HealthHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		var hr HealthResponse
		hr.ListenerName = hcc.ListenerName
		hr.BuildNumber = info.BuildVersion

		b, err := json.Marshal(&hr)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.Write(b)
	}
}

//HealthResponse is used to structure the json response payload for the health check
type HealthResponse struct {
	ListenerName string `json:"listenerName"`
	BuildNumber  string `json:"buildNumber"`
}
