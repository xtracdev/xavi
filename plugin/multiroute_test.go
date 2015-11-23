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

func (th *TestMRHandler) MultiRouteServeHTTP(bhMap BackendHandlerMap, w http.ResponseWriter, r *http.Request) {
	aHandler := bhMap["A"]
	aHandler.ServeHTTP(w, r)
}

func ATestMRHandlerFactory(bhMap BackendHandlerMap, mrHandler MultirouteHandler) *MultiRouteAdapter {
	return &MultiRouteAdapter{
		Ctx:     bhMap,
		Handler: mrHandler,
	}
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
