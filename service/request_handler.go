package service

import (
	"container/list"
	"expvar"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/armon/go-metrics"
	"github.com/xtracdev/xavi/plugin"
	"github.com/xtracdev/xavi/plugin/timing"
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
	"io"
	"net/http"
	"strings"
)

var contextCounts = expvar.NewMap("contextCounts")

func incCounter(counterName string) {
	contextCounts.Add(counterName, 1)
	metrics.IncrCounter([]string{counterName}, 1.0)
}

func incrementErrorCounts(err error) {
	if err == context.Canceled {
		incCounter("cancelled-count")
	} else if err == context.DeadlineExceeded {
		incCounter("timeout-count")
	}
}

//Service represents a runnable service
type Service interface {
	Run()
}

//Request handler has the configuration needed to build an http.Handler for a route and its chained plugins
type requestHandler struct {
	Transport    *http.Transport
	TLSTransport *http.Transport
	Backend      *backend
	PluginChain  *list.List
}

func backendName(name string) string {
	if strings.Contains(name, "backend") {
		return name
	} else {
		return name + "-backend"
	}
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

		timingContributor := rt.StartContributor(backendName(rh.Backend.Name))

		connectString, err := rh.Backend.getConnectAddress()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			timingContributor.End(err)
			return
		}

		log.Debug("connect string for ", rh.Backend.Name, "is ", connectString)
		r.URL.Host = connectString
		r.Host = connectString

		log.Debug("invoke backend service")
		serviceName := timing.GetServiceNameFromContext(ctx)
		if serviceName == "" {
			serviceName = "backend-call"
		}

		r.URL.Scheme = "http"
		log.Debug("Determine transport")
		var transport = rh.getTransportForBackend(ctx)
		if transport == rh.TLSTransport {
			log.Debug("https transport")
			r.URL.Scheme = "https"
		}

		log.Debug(r.URL.Scheme, " transport for backend ", rh.Backend.Name)

		beTimer := timingContributor.StartServiceCall(serviceName, connectString)
		log.Debug("call service ", serviceName, " for backend ", rh.Backend.Name)

		//resp, err := transport.RoundTrip(r)
		client := &http.Client{
			Transport: transport,
		}

		r.RequestURI = "" //Must clear when using http.Client
		resp, err := ctxhttp.Do(ctx, client, r)

		beTimer.End(err)
		if err != nil {
			incrementErrorCounts(err)
			log.Info(err.Error())
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

func (rh *requestHandler) getTransportForBackend(ctx context.Context) *http.Transport {
	//If we always use TLS use the TLS transport
	if rh.Backend.TLSOnly {
		log.Debug("tlsonly transport for backend ", rh.Backend.Name)
		return rh.TLSTransport
	}

	//Does the context indicate that his particular call will be https?
	useHttps := plugin.GetUseHttpsContext(ctx)
	switch useHttps {
	case true:
		log.Debug("https transport from context for backend ", rh.Backend.Name)
		return rh.TLSTransport
	default:
		log.Debug("Non-TLS transport for backend ", rh.Backend.Name)
		return rh.Transport
	}
}
