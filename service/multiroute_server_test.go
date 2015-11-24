package service

import (
	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/plugin"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func handleAStuff(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("a backend stuff\n"))
}

func handleBStuff(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("b backend stuff\n"))
}

func TestMRConfigListener(t *testing.T) {
	log.SetLevel(log.InfoLevel)

	var bHandler plugin.MultiRouteHandlerFunc = func(m plugin.BackendHandlerMap, w http.ResponseWriter, r *http.Request) {
		println("handler called")
		w.Write([]byte("b stuff\n"))

		ah := m["backendA"]
		ar := httptest.NewRecorder()
		ah.ServeHTTP(ar, r)
		assert.Equal(t, "a backend stuff\n", ar.Body.String())

		bh := m["backendB"]
		br := httptest.NewRecorder()
		bh.ServeHTTP(br, r)
		assert.Equal(t, "b backend stuff\n", br.Body.String())
	}

	var BMRAFactory = func(bhMap plugin.BackendHandlerMap) *plugin.MultiRouteAdapter {
		return &plugin.MultiRouteAdapter{
			Ctx:     bhMap,
			Handler: bHandler,
		}
	}

	plugin.RegisterMRAFactory("test-multiroute-plugin", BMRAFactory)

	AServer := httptest.NewServer(http.HandlerFunc(handleAStuff))
	BServer := httptest.NewServer(http.HandlerFunc(handleBStuff))

	defer AServer.Close()
	defer BServer.Close()

	ms := mrtBuildListener(AServer.URL, BServer.URL)

	uriToRoutesMap := ms.mapUrisToRoutes()
	uriToGuardAndHandlerMap := mapRoutesToGuardAndHandler(uriToRoutesMap)
	uriHandlerMap := makeURIHandlerMap(uriToGuardAndHandlerMap)

	assert.Equal(t, 1, len(uriHandlerMap))

	ls := httptest.NewServer(uriHandlerMap["/foo"])
	defer ls.Close()

	resp, err := http.Get(ls.URL + "/foo")
	assert.Nil(t, err)
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	assert.True(t, strings.Contains(string(body), "b stuff"))
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

	backEndA := makeBackend("backendA", serverA)
	backEndB := makeBackend("backendB", serverB)

	var r1 = route{
		Name:                 "route1",
		URIRoot:              "/foo",
		Backends:             []*backend{backEndA, backEndB},
		MultiRoutePluginName: "test-multiroute-plugin",
	}

	var ms = managedService{
		Address:      "localhost:23456",
		ListenerName: "test listener",
		Routes:       []route{r1},
	}

	return &ms

}
