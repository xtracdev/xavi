package service

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/plugin"
	"github.com/xtracdev/xavi/plugin/timing"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func postHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Println("handling post")
	bb, _ := ioutil.ReadAll(req.Body)

	req.Body.Close()
	rw.WriteHeader(200)
	rw.Write([]byte(bb))
}

func makeTestWrapper() plugin.Wrapper {
	return new(testWrapper)
}

type testWrapper struct{}

func (aw testWrapper) Wrap(h plugin.ContextHandler) plugin.ContextHandler {
	return plugin.ContextHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		rec := httptest.NewRecorder()
		h.ServeHTTPContext(ctx, rec, r)

		upperOut := strings.ToUpper(string(rec.Body.Bytes()))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(upperOut))

	})
}

func makeTestBackend(t *testing.T, testServerURL string, loadBalancerPolicyName string) *backend {

	testURL, err := url.Parse(testServerURL)
	assert.Nil(t, err)

	hostAndPort := strings.Split(testURL.Host, ":")
	port, _ := strconv.Atoi(hostAndPort[1])

	serverConfig := config.ServerConfig{
		Name:    "s1",
		Address: hostAndPort[0],
		Port:    port,
		PingURI: "/xtracrulesok",
	}

	servers := []config.ServerConfig{serverConfig}

	var b backend
	b.Name = "test-backend"
	loadBalancer, err := instantiateLoadBalancer(loadBalancerPolicyName, b.Name, servers)
	if err != nil {
		t.Log("Error instantiating test load balancer ", err)
		t.FailNow()
	}
	b.LoadBalancer = loadBalancer

	return &b
}

func TestPostRequest(t *testing.T) {
	t.Log("Given a server that echos back the post body")
	ts := httptest.NewServer(http.HandlerFunc(postHandler))
	defer ts.Close()

	backend := makeTestBackend(t, ts.URL, "")

	requestHandler := &requestHandler{
		Transport: &http.Transport{DisableKeepAlives: false, DisableCompression: false},
		Backend:   backend,
	}

	t.Log("When the echo server is proxied")
	handlerFn := requestHandler.toContextHandlerFunc()

	adapter := &plugin.ContextAdapter{
		Ctx:     context.Background(),
		Handler: timing.RequestTimerMiddleware(plugin.ContextHandlerFunc(handlerFn)),
	}

	ts2 := httptest.NewServer(adapter)
	defer ts2.Close()

	payload := `
	{
	"field1","val1",
	"field2","field2"
	}
	`

	req, _ := http.NewRequest("POST", ts2.URL+"/foo", bytes.NewBuffer([]byte(payload)))

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.Nil(t, err)

	defer resp.Body.Close()

	t.Log("Then the response is the request body that was posted")
	assert.Equal(t, 200, resp.StatusCode)
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, payload, string(body))

}

func TestPostRequestWithPlugin(t *testing.T) {
	t.Log("Given a server that echos back the post body")
	ts := httptest.NewServer(http.HandlerFunc(postHandler))
	defer ts.Close()

	backend := makeTestBackend(t, ts.URL, "round-robin")

	requestHandler := &requestHandler{
		Transport: &http.Transport{DisableKeepAlives: false, DisableCompression: false},
		Backend:   backend,
	}

	handlerFn := requestHandler.toContextHandlerFunc()

	wrapper := makeTestWrapper()
	wrappedHandler := (wrapper.Wrap(plugin.ContextHandlerFunc(handlerFn)))
	wrappedHandler = timing.RequestTimerMiddleware(wrappedHandler)

	t.Log("When the echo server is proxied with a wrapped handler")

	adapter := &plugin.ContextAdapter{
		Ctx:     context.Background(),
		Handler: wrappedHandler,
	}
	ts2 := httptest.NewServer(adapter)
	defer ts2.Close()

	payload := `
	{
	"field1","val1",
	"field2","field2"
	}
	`

	req, _ := http.NewRequest("POST", ts2.URL+"/foo", bytes.NewBuffer([]byte(payload)))

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.Nil(t, err)

	defer resp.Body.Close()

	t.Log("Then the response is the request body that was posted as transformed by the handler")
	assert.Equal(t, 200, resp.StatusCode)
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, strings.ToUpper(payload), string(body))

}

func TestExpVarsHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(expvarHandler))
	defer ts.Close()

	client := &http.Client{}
	resp, err := client.Get(ts.URL)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func makeListenerWithRoutesForTest(t *testing.T, loadBalancerPolicyName string) *managedService {

	testBackend := makeTestBackend(t, "http://localhost:666", loadBalancerPolicyName)
	backends := []*backend{testBackend}

	var r1 = route{
		Name:     "route1",
		URIRoot:  "/foo",
		Backends: backends,
		MsgProps: "Foo=bar",
	}

	var r2 = route{
		Name:     "route2",
		URIRoot:  "/foo",
		Backends: backends,
		MsgProps: "Foo=baz",
	}

	var r3 = route{
		Name:     "route3",
		URIRoot:  "/bar",
		Backends: backends,
	}

	var ms = managedService{
		Address:      "localhost:1234",
		ListenerName: "test listener",
		Routes:       []route{r1, r2, r3},
	}

	return &ms

}

func makeTestBackends(t *testing.T, testServerURL string, loadBalancerPolicyName string) []*backend {
	b := makeTestBackend(t, testServerURL, loadBalancerPolicyName)
	return []*backend{b, b}
}

func makeListenerWithMultiRoutesForTest(t *testing.T, loadBalancerPolicyName string) *managedService {

	var bHandler plugin.MultiBackendHandlerFunc = func(m plugin.BackendHandlerMap, ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("b stuff"))

		_, ok := m["A"]
		if ok == true {
			w.Write([]byte("context stuff"))
		}
	}

	var BMRAFactory = func(bhMap plugin.BackendHandlerMap) *plugin.MultiBackendAdapter {
		return &plugin.MultiBackendAdapter{
			BackendHandlerCtx: bhMap,
			Handler:           bHandler,
		}
	}

	plugin.RegisterMultiBackendAdapterFactory("test-multiroute-plugin", BMRAFactory)

	testBackends := makeTestBackends(t, "http://localhost:666", "")
	var r1 = route{
		Name:                   "route2",
		URIRoot:                "/foo2",
		Backends:               testBackends,
		MultiBackendPluginName: "test-multiroute-plugin",
	}

	var ms = managedService{
		Address:      "localhost:1234",
		ListenerName: "test listener",
		Routes:       []route{r1},
	}

	return &ms
}

func makeListenerWithBrokenMsgPropForTest(t *testing.T) *managedService {

	testBackend := makeTestBackend(t, "http://localhost:666", "")
	backends := []*backend{testBackend}

	var r1 = route{
		Name:     "route1",
		URIRoot:  "/foo",
		Backends: backends,
		MsgProps: "xxxxx",
	}

	var r2 = route{
		Name:     "route2",
		URIRoot:  "/foo",
		Backends: backends,
		MsgProps: "Foo",
	}

	var r3 = route{
		Name:     "route3",
		URIRoot:  "/bar",
		Backends: backends,
	}

	var ms = managedService{
		Address:      "localhost:1234",
		ListenerName: "test listener",
		Routes:       []route{r1, r2, r3},
	}

	return &ms

}

func makePanickyServiceConfig(t *testing.T) *managedService {

	testBackend := makeTestBackend(t, "http://localhost:666", "round-robin")
	backends := []*backend{testBackend}

	var r1 = route{
		Name:     "route1",
		URIRoot:  "/foo",
		Backends: backends,
	}

	var r2 = route{
		Name:     "route2",
		URIRoot:  "/foo",
		Backends: backends,
	}

	var ms = managedService{
		Address:      "localhost:1234",
		ListenerName: "test listener",
		Routes:       []route{r1, r2},
	}

	return &ms

}

func validateURIRoutesMap(uriRoutesMap map[string][]route, t *testing.T) {

	assert.Equal(t, 2, len(uriRoutesMap))

	fooRoutes := uriRoutesMap["/foo"]
	assert.NotNil(t, fooRoutes)
	assert.Equal(t, 2, len(fooRoutes))

	barRoutes := uriRoutesMap["/bar"]
	assert.NotNil(t, barRoutes)
	assert.Equal(t, 1, len(barRoutes))
}

func validateURIToGuardAndHandlerMapping(ghMap map[string][]guardAndHandler, t *testing.T) {
	assert.Equal(t, 2, len(ghMap))

	fooRoutes := ghMap["/foo"]
	assert.NotNil(t, fooRoutes)
	assert.Equal(t, 2, len(fooRoutes))

	req, _ := http.NewRequest("GET", "http://localhost:123/foo", nil)
	req.Header.Set("Foo", "bar")

	//One of the two functions should match
	var fooMatches int
	for _, r := range fooRoutes {
		guardFn := r.Guard
		match, err := guardFn(req)
		assert.Nil(t, err)
		if match {
			fooMatches++
		}
	}

	//Neither request should match
	fooMatches = 0
	req.Header.Set("Foo", "no way, Jose")
	for _, r := range fooRoutes {
		guardFn := r.Guard
		match, err := guardFn(req)
		assert.Nil(t, err)
		if match {
			fooMatches++
		}
	}

	assert.Equal(t, 0, fooMatches)

	barRoutes := ghMap["/bar"]
	assert.NotNil(t, barRoutes)
	assert.Equal(t, 1, len(barRoutes))

	//No guard on bar route, so it should match the last request we made
	match, err := barRoutes[0].Guard(req)
	assert.Nil(t, err)
	assert.True(t, match)
}

func validateURIHandlerMap(handlers map[string]plugin.ContextHandler, t *testing.T) {
	assert.Equal(t, 2, len(handlers))
	assert.NotNil(t, handlers["/foo"])
	assert.NotNil(t, handlers["/bar"])
	assert.Nil(t, handlers["no way, Jose"])

	handler := handlers["/foo"]
	handler = timing.RequestTimerMiddleware(handler)

	adapter := &plugin.ContextAdapter{
		Ctx:     context.Background(),
		Handler: handler,
	}

	ts := httptest.NewServer(adapter)
	t.Log("test server url", ts.URL)
	defer ts.Close()

	client := &http.Client{}
	req, err := http.NewRequest("GET", ts.URL+"/foo", nil)
	req.Header.Set("Foo", "bar")
	resp, err := client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	req.Header.Set("Foo", "no match")
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	req.Header.Set("Foo", "no match")
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

}

func TestMakeOfHandlersFromConfig(t *testing.T) {
	ms := makeListenerWithRoutesForTest(t, "")
	uriRoutesMap := ms.organizeRoutesByUri()
	validateURIRoutesMap(uriRoutesMap, t)

	uriToGuardAndHandlerMap := mapRoutesToGuardAndHandler(uriRoutesMap)
	validateURIToGuardAndHandlerMapping(uriToGuardAndHandlerMap, t)

	t.Log("Validate handler creation")
	uriHandlerMap := makeURIHandlerMap(uriToGuardAndHandlerMap)
	validateURIHandlerMap(uriHandlerMap, t)
}

func TestMakeOfHandlersFromMultiRouteConfig(t *testing.T) {
	ms := makeListenerWithMultiRoutesForTest(t, "")
	uriRoutesMap := ms.organizeRoutesByUri()

	uriToGuardAndHandlerMap := mapRoutesToGuardAndHandler(uriRoutesMap)

	t.Log("Validate handler creation")
	uriHandlerMap := makeURIHandlerMap(uriToGuardAndHandlerMap)

	println(len(uriHandlerMap))
}

func TestGuardFnGenWithBrokerHeaderProp(t *testing.T) {
	ms := makeListenerWithBrokenMsgPropForTest(t)
	uriRoutesMap := ms.organizeRoutesByUri()
	uriToGuardAndHandlerMap := mapRoutesToGuardAndHandler(uriRoutesMap)
	uriHandlerMap := makeURIHandlerMap(uriToGuardAndHandlerMap)

	handler := uriHandlerMap["/foo"]
	handler = timing.RequestTimerMiddleware(handler)

	adapter := &plugin.ContextAdapter{
		Ctx:     context.Background(),
		Handler: handler,
	}

	ts := httptest.NewServer(adapter)
	t.Log("test server url", ts.URL)
	defer ts.Close()

	client := &http.Client{}
	req, err := http.NewRequest("GET", ts.URL+"/foo", nil)
	req.Header.Set("Foo", "bar")
	resp, err := client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

}

func TestPanicGuardConfig(t *testing.T) {
	ms := makePanickyServiceConfig(t)
	assert.Panics(t, func() {
		ms.organizeRoutesByUri()
	})
}
