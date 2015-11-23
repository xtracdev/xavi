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

func (th *TestMRHandler) MultiRouteServeHTTP(bhMap backendHandlerMap, w http.ResponseWriter, r *http.Request) {
	aHandler := bhMap["A"]
	aHandler.ServeHTTP(w, r)
}

func TestMultiRoutePlugin(t *testing.T) {

	var handlerMap backendHandlerMap = make(backendHandlerMap)
	handlerMap["A"] = http.HandlerFunc(handleAStuff)
	adapter := &MultiRouteAdapter{
		Ctx:     handlerMap,
		Handler: &TestMRHandler{},
	}

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

	var handlerMap backendHandlerMap = make(backendHandlerMap)
	handlerMap["A"] = http.HandlerFunc(handleAStuff)
	adapter := &MultiRouteAdapter{
		Ctx:     handlerMap,
		Handler: &TestMRHandler{},
	}

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
