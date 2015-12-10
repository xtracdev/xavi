package service

import (
	"container/list"
	"expvar"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/armon/go-metrics"
	"github.com/xtracdev/xavi/statsd"
	"golang.org/x/net/context"
	"io"
	"net/http"
)

var (
	counts = expvar.NewMap("counters")
)

func counterName(method string, path string) string {
	return fmt.Sprintf("%s::%s", method, path)
}

func incCounter(method string, path string) {
	counter := counterName(method, path)
	counts.Add(counter, 1)
}

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

//Increment service counter
func incServiceCounter(name string) {
	metrics.IncrCounter([]string{statsd.FormatServiceName(name)}, 1.0)
}

//Create a handler function from a requestHandler
func (rh *requestHandler) toContextHandlerFunc() func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		st := NewServiceTimer(r)

		r.URL.Scheme = "http"

		connectString, err := rh.Backend.getConnectAddress()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			st.ConnectFail(err)
			go st.EndService(http.StatusServiceUnavailable)
			return
		}

		log.Debug("connect string is ", connectString)
		r.URL.Host = connectString
		r.Host = connectString

		incCounter(r.Method, r.RequestURI)

		log.Debug("invoke backend service")
		st.BackendCallStart()
		resp, err := rh.Transport.RoundTrip(r)
		//TODO - timing context wrapper with better timer name support.
		metrics.MeasureSince([]string{"timing_" + r.RequestURI}, st.backendStartTime)
		st.BackendCallEnd(err)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "Error: %v", err)
			metrics.IncrCounter([]string{"error_" + r.RequestURI}, 1.0)
			go st.EndService(http.StatusServiceUnavailable)
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

		go st.EndService(resp.StatusCode)
		go incServiceCounter(r.URL.String())
	}
}
