package plugin

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sync"

	"context"
	"github.com/stretchr/testify/assert"
)

type TestMRHandler struct{}

const testCtxKey = 100

func handleAStuff(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("a stuff"))

	val, ok := r.Context().Value(testCtxKey).(string)
	if ok {
		w.Write([]byte(val))
	}
}

var bHandler MultiBackendHandlerFunc = func(m BackendHandlerMap, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("b stuff"))

	_, ok := m["A"]
	if ok == true {
		w.Write([]byte("backend context A"))
	}

	val, ok := r.Context().Value(testCtxKey).(string)
	if ok {
		w.Write([]byte(val))
	}
}

func (th *TestMRHandler) MultiBackendServeHTTP(bhMap BackendHandlerMap, w http.ResponseWriter, r *http.Request) {
	aHandler := bhMap["A"]
	aHandler.ServeHTTP(w, r)
}

func BMRAFactory(bhMap BackendHandlerMap) *MultiBackendAdapter {
	return &MultiBackendAdapter{
		BackendHandlerCtx: bhMap,
		Handler:           bHandler,
	}
}

func ATestMRHandlerFactory(bhMap BackendHandlerMap, mrHandler MultiBackendHandler) *MultiBackendAdapter {
	return &MultiBackendAdapter{
		BackendHandlerCtx: bhMap,
		Handler:           mrHandler,
	}
}

func TestMultiBackendHandlerFunc(t *testing.T) {

	var handlerMap = make(BackendHandlerMap)
	handlerMap["A"] = http.HandlerFunc(handleAStuff)

	adapter := &MultiBackendAdapter{
		BackendHandlerCtx: handlerMap,
		Handler:           bHandler,
	}

	ts := httptest.NewServer(adapter)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	req = req.WithContext(context.WithValue(req.Context(), testCtxKey, "foo"))

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	assert.True(t, strings.Contains(string(body), "b stuff"))

}

func TestMultiBackendAdapter(t *testing.T) {

	var handlerMap = make(BackendHandlerMap)
	handlerMap["A"] = http.HandlerFunc(handleAStuff)
	adapter := ATestMRHandlerFactory(handlerMap, &TestMRHandler{})

	ts := httptest.NewServer(adapter)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	req = req.WithContext(context.WithValue(req.Context(), testCtxKey, "foo"))

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	assert.True(t, strings.Contains(string(body), "a stuff"))
}

func TestMBAWithFactory(t *testing.T) {
	assert.False(t, MultiBackendAdapterRegistryContains("b-plugin"))
	var factory MultiBackendAdapterFactory = BMRAFactory
	RegisterMultiBackendAdapterFactory("b-plugin", factory)
	assert.True(t, MultiBackendAdapterRegistryContains("b-plugin"))

	registeredAdapters := ListMultiBackendAdapters()
	assert.Equal(t, 1, len(registeredAdapters))
	assert.Equal(t, "b-plugin", registeredAdapters[0])

	factoryFromReg, err := LookupMultiBackendAdapterFactory("b-plugin")
	assert.Nil(t, err)
	assert.NotNil(t, factoryFromReg)

	var handlerMap = make(BackendHandlerMap)
	handlerMap["A"] = http.HandlerFunc(handleAStuff)
	adapter := factoryFromReg(handlerMap)

	ts := httptest.NewServer(adapter)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	assert.Nil(t, err)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	println("--->", string(body))
	assert.True(t, strings.Contains(string(body), "b stuff"))
	assert.True(t, strings.Contains(string(body), "backend context A"))

}

func TestMBAWithFactoryConcurrently(t *testing.T) {
	var factory MultiBackendAdapterFactory = BMRAFactory
	RegisterMultiBackendAdapterFactory("b-plugin", factory)
	assert.True(t, MultiBackendAdapterRegistryContains("b-plugin"))

	registeredAdapters := ListMultiBackendAdapters()
	assert.Equal(t, 1, len(registeredAdapters))
	assert.Equal(t, "b-plugin", registeredAdapters[0])

	factoryFromReg, err := LookupMultiBackendAdapterFactory("b-plugin")
	assert.Nil(t, err)
	assert.NotNil(t, factoryFromReg)

	var handlerMap = make(BackendHandlerMap)
	handlerMap["A"] = http.HandlerFunc(handleAStuff)
	adapter := factoryFromReg(handlerMap)

	ts := httptest.NewServer(adapter)
	defer ts.Close()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			resp, err := http.Get(ts.URL)
			assert.Nil(t, err)
			body, _ := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			assert.True(t, strings.Contains(string(body), "b stuff"))
			assert.True(t, strings.Contains(string(body), "backend context A"))
		}()
	}
	wg.Wait()
}

func TestWrappedPlugin(t *testing.T) {
	wrapper := NewAWrapper()

	var handlerMap = make(BackendHandlerMap)
	handlerMap["A"] = http.HandlerFunc(handleAStuff)
	adapter := ATestMRHandlerFactory(handlerMap, &TestMRHandler{})

	wrapped := wrapper.Wrap(adapter.ToHandlerFunc())

	ts := httptest.NewServer(wrapped)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	req = req.WithContext(context.WithValue(req.Context(), testCtxKey, "foo"))

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	assert.True(t, strings.Contains(string(body), "a stuff"))
	assert.True(t, strings.Contains(string(body), "A wrapper wrote this"))

}

func TestEmptyAdapterNameRegistration(t *testing.T) {
	err := RegisterMultiBackendAdapterFactory("", nil)
	assert.NotNil(t, err)
}

func TestLookupOfNonRegisteredMBAFactory(t *testing.T) {
	_, err := LookupMultiBackendAdapterFactory("foobar")
	assert.NotNil(t, err)
}
