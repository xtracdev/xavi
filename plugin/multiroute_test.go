package plugin

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type TestMRHandler struct{}

const testCtxKey = 100

func handleAStuff(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("a stuff"))

	val, ok := ctx.Value(testCtxKey).(string)
	if ok {
		w.Write([]byte(val))
	}
}

var bHandler MultiBackendHandlerFunc = func(m BackendHandlerMap, ctx context.Context, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("b stuff"))

	_, ok := m["A"]
	if ok == true {
		w.Write([]byte("backend context A"))
	}

	val, ok := ctx.Value(testCtxKey).(string)
	if ok {
		w.Write([]byte(val))
	}
}

func (th *TestMRHandler) MultiBackendServeHTTP(bhMap BackendHandlerMap, ctx context.Context, w http.ResponseWriter, r *http.Request) {
	aHandler := bhMap["A"]
	aHandler.ServeHTTPContext(ctx, w, r)
}

func BMRAFactory(bhMap BackendHandlerMap) *MultiBackendAdapter {
	return &MultiBackendAdapter{
		BackendHandlerCtx: bhMap,
		Ctx:               context.Background(),
		Handler:           bHandler,
	}
}

func ATestMRHandlerFactory(bhMap BackendHandlerMap, mrHandler MultiBackendHandler) *MultiBackendAdapter {
	adapter := &MultiBackendAdapter{
		BackendHandlerCtx: bhMap,
		Ctx:               context.Background(),
		Handler:           mrHandler,
	}

	adapter.Ctx = context.WithValue(adapter.Ctx, testCtxKey, "foo")

	return adapter

}

func adaptAWithFooContext() *MultiBackendAdapter {
	var handlerMap = make(BackendHandlerMap)
	handlerMap["A"] = ContextHandlerFunc(handleAStuff)

	adapter := &MultiBackendAdapter{
		BackendHandlerCtx: handlerMap,
		Ctx:               context.Background(),
		Handler:           bHandler,
	}

	adapter.Ctx = context.WithValue(adapter.Ctx, testCtxKey, "foo")

	return adapter
}

func TestMultiBackendHandlerFunc(t *testing.T) {

	adapter := adaptAWithFooContext()

	ts := httptest.NewServer(adapter)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	assert.Nil(t, err)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	assert.True(t, strings.Contains(string(body), "b stuff"))

}

func TestMultiBackendAdapter(t *testing.T) {

	var handlerMap = make(BackendHandlerMap)
	handlerMap["A"] = ContextHandlerFunc(handleAStuff)
	adapter := ATestMRHandlerFactory(handlerMap, &TestMRHandler{})

	ts := httptest.NewServer(adapter)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
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
	handlerMap["A"] = ContextHandlerFunc(handleAStuff)
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

func TestWrappedPlugin(t *testing.T) {
	wrapper := NewAWrapper()

	var handlerMap = make(BackendHandlerMap)
	handlerMap["A"] = ContextHandlerFunc(handleAStuff)
	adapter := ATestMRHandlerFactory(handlerMap, &TestMRHandler{})

	wrapped := wrapper.Wrap(adapter)

	ctxAdapter := &ContextAdapter{
		Ctx:     context.Background(),
		Handler: wrapped,
	}
	ts := httptest.NewServer(ctxAdapter)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
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
