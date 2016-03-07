package service

import (
	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/plugin"
	"github.com/xtracdev/xavi/plugin/timing"
	"golang.org/x/net/context"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

const (
	aBackendResponse           = "a backend stuff\n"
	bBackendResponse           = "b backend stuff\n"
	bHandlerStuff              = "b stuff\n"
	backendA                   = "backendA"
	backendB                   = "backendB"
	fooURI                     = "/foo"
	multiBackendAdapterFactory = "test-multiroute-plugin"
)

func handleAStuff(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(aBackendResponse))
}

func handleBStuff(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(bBackendResponse))
}

func TestMRConfigListener(t *testing.T) {
	log.SetLevel(log.InfoLevel)

	var bHandler plugin.MultiBackendHandlerFunc = func(m plugin.BackendHandlerMap, ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(bHandlerStuff))

		ah := m[backendA]
		ar := httptest.NewRecorder()
		ah.ServeHTTPContext(ctx, ar, r)
		assert.Equal(t, aBackendResponse, ar.Body.String())

		bh := m[backendB]
		br := httptest.NewRecorder()
		bh.ServeHTTPContext(ctx, br, r)
		assert.Equal(t, bBackendResponse, br.Body.String())
	}

	var BMBAFactory = func(bhMap plugin.BackendHandlerMap) *plugin.MultiBackendAdapter {
		return &plugin.MultiBackendAdapter{
			BackendHandlerCtx: bhMap,
			Handler:           bHandler,
		}
	}

	plugin.RegisterMultiBackendAdapterFactory(multiBackendAdapterFactory, BMBAFactory)

	AServer := httptest.NewServer(http.HandlerFunc(handleAStuff))
	BServer := httptest.NewServer(http.HandlerFunc(handleBStuff))

	defer AServer.Close()
	defer BServer.Close()

	ms := mrtBuildListener(AServer.URL, BServer.URL)

	uriToRoutesMap := ms.organizeRoutesByUri()
	uriToGuardAndHandlerMap := mapRoutesToGuardAndHandler(uriToRoutesMap)
	uriHandlerMap := makeURIHandlerMap(uriToGuardAndHandlerMap)

	assert.Equal(t, 1, len(uriHandlerMap))

	adapter := &plugin.ContextAdapter{
		Ctx:     context.Background(),
		Handler: timing.NewTimerWrapper().Wrap(uriHandlerMap[fooURI]),
	}
	ls := httptest.NewServer(adapter)
	defer ls.Close()

	resp, err := http.Get(ls.URL + fooURI)
	assert.Nil(t, err)
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	assert.True(t, strings.Contains(string(body), bHandlerStuff))
}

func makeServerConfig(name string, theURL string) config.ServerConfig {
	parseUrl, _ := url.Parse(theURL)
	host, port, _ := net.SplitHostPort(parseUrl.Host)
	portVal, _ := strconv.Atoi(port)

	println(host)
	println(portVal)

	return config.ServerConfig{
		Name:    name,
		Address: host,
		Port:    portVal,
		PingURI: "/xtracrulesok",
	}
}

func makeBackend(name string, serverConfig config.ServerConfig) *backend {
	servers := []config.ServerConfig{serverConfig}
	var b backend
	b.Name = name
	loadBalancer, err := instantiateLoadBalancer("round-robin", b.Name, servers)
	if err != nil {
		panic(err.Error())
	}
	b.LoadBalancer = loadBalancer

	return &b

}

func mrtBuildListener(urlA string, urlB string) *managedService {
	serverA := makeServerConfig("server1", urlA)
	serverB := makeServerConfig("server2", urlB)

	backEndA := makeBackend(backendA, serverA)
	backEndB := makeBackend(backendB, serverB)

	var r1 = route{
		Name:                   "route1",
		URIRoot:                fooURI,
		Backends:               []*backend{backEndA, backEndB},
		MultiBackendPluginName: multiBackendAdapterFactory,
	}

	var ms = managedService{
		Address:      "localhost:23456", //Ignored - we use a testserver with a dyn addr for testing
		ListenerName: "test listener",
		Routes:       []route{r1},
	}

	return &ms

}
