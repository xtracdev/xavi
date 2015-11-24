package plugin

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type TestMRHandler struct{}

func handleAStuff(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("a stuff"))
}

var bHandler MultiRouteHandlerFunc = func(m BackendHandlerMap, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("b stuff"))

	_, ok := m["A"]
	if ok == true {
		w.Write([]byte("context stuff"))
	}
}

func (th *TestMRHandler) MultiRouteServeHTTP(bhMap BackendHandlerMap, w http.ResponseWriter, r *http.Request) {
	aHandler := bhMap["A"]
	aHandler.ServeHTTP(w, r)
}

func BMRAFactory(bhMap BackendHandlerMap) *MultiRouteAdapter {
	return &MultiRouteAdapter{
		Ctx:     bhMap,
		Handler: bHandler,
	}
}

func ATestMRHandlerFactory(bhMap BackendHandlerMap, mrHandler MultirouteHandler) *MultiRouteAdapter {
	return &MultiRouteAdapter{
		Ctx:     bhMap,
		Handler: mrHandler,
	}
}

func TestMultiRouteHandlerFunc(t *testing.T) {
	var handlerMap BackendHandlerMap = make(BackendHandlerMap)
	handlerMap["A"] = http.HandlerFunc(handleAStuff)

	adapter := &MultiRouteAdapter{
		Ctx:     handlerMap,
		Handler: bHandler,
	}

	ts := httptest.NewServer(adapter)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	assert.Nil(t, err)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	assert.True(t, strings.Contains(string(body), "b stuff"))

}

func TestMultiRoutePlugin(t *testing.T) {

	var handlerMap BackendHandlerMap = make(BackendHandlerMap)
	handlerMap["A"] = http.HandlerFunc(handleAStuff)
	adapter := ATestMRHandlerFactory(handlerMap, &TestMRHandler{})

	ts := httptest.NewServer(adapter)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	assert.Nil(t, err)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	assert.True(t, strings.Contains(string(body), "a stuff"))
}

func TestMRPluginWithFactory(t *testing.T) {
	assert.False(t, MRARegistryContains("b-plugin"))
	var factory MultiRouteAdapterFactory = BMRAFactory
	RegisterMRAFactory("b-plugin", factory)
	assert.True(t, MRARegistryContains("b-plugin"))

	registeredAdapters := ListMultirouteAdapters()
	assert.Equal(t, 1, len(registeredAdapters))
	assert.Equal(t, "b-plugin", registeredAdapters[0])

	factoryFromReg, err := LookupMRAFactory("b-plugin")
	assert.Nil(t, err)
	assert.NotNil(t, factoryFromReg)

	var handlerMap BackendHandlerMap = make(BackendHandlerMap)
	handlerMap["A"] = http.HandlerFunc(handleAStuff)
	adapter := factoryFromReg(handlerMap)

	ts := httptest.NewServer(adapter)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	assert.Nil(t, err)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	assert.True(t, strings.Contains(string(body), "b stuff"))
	assert.True(t, strings.Contains(string(body), "context stuff"))

}

func TestWrappedPlugin(t *testing.T) {
	wrapper := NewAWrapper()

	var handlerMap BackendHandlerMap = make(BackendHandlerMap)
	handlerMap["A"] = http.HandlerFunc(handleAStuff)
	adapter := ATestMRHandlerFactory(handlerMap, &TestMRHandler{})

	wrapped := wrapper.Wrap(adapter)

	ts := httptest.NewServer(wrapped)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	assert.Nil(t, err)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	assert.True(t, strings.Contains(string(body), "a stuff"))
	assert.True(t, strings.Contains(string(body), "A wrapper wrote this"))

}

func TestEmptyAdapterNameRegistration(t *testing.T) {
	err := RegisterMRAFactory("", nil)
	assert.NotNil(t, err)
}

func TestLookupOfNonRegisteredMRAFactory(t *testing.T) {
	_, err := LookupMRAFactory("foobar")
	assert.NotNil(t, err)
}
