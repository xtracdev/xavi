package service

import (
	"container/list"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"github.com/xtracdev/xavi/plugin/timing"
)

//Service represents a runnable service
type Service interface {
	Run()
}

//Request handler has the configuration needed to build an http.Handler for a route and its chained plugins
type requestHandler struct {
	Transport   *http.Transport
	Backend     *backend
	PluginChain *list.List
}

//Create a handler function from a requestHandler
func (rh *requestHandler) toContextHandlerFunc() func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {

		//Record call time contribution
		rt := timing.TimerFromContext(ctx)
		if rt == nil {
			http.Error(w, "No EndToEndTimer found in call context", http.StatusInternalServerError)
			return
		}

		timingContributor := rt.StartContributor(rh.Backend.Name + " backend")

		r.URL.Scheme = "http"

		connectString, err := rh.Backend.getConnectAddress()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			timingContributor.End(err)
			return
		}

		log.Debug("connect string is ", connectString)
		r.URL.Host = connectString
		r.Host = connectString

		log.Debug("invoke backend service")
		beTimer := timingContributor.StartServiceCall("backend call " + r.RequestURI)
		resp, err := rh.Transport.RoundTrip(r)
		beTimer.End(err)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "Error: %v", err)
			timingContributor.End(err)
			return
		}

		log.Debug("backend service complete, copy backend response headers to response")
		for k, v := range resp.Header {
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}

		log.Debug("write status code to response")
		w.WriteHeader(resp.StatusCode)

		log.Debug("Copy body to response")
		io.Copy(w, resp.Body)
		resp.Body.Close()

		timingContributor.End(nil)
	}
}
