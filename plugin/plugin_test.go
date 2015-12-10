package plugin

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

//NewAWrapper instantiates AWrapper
func NewAWrapper() Wrapper {
	return new(AWrapper)
}

//AWrapper can wrap http handlers
type AWrapper struct{}

//Wrap wraps http.Handlers with A stuff
func (aw AWrapper) Wrap(h ContextHandler) ContextHandler {
	return ContextHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		h.ServeHTTPContext(ctx, w, r)
		w.Write([]byte("A wrapper wrote this\n"))
	})
}

func handleCall(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("handleCall wrote this stuff\n"))
}

func TestPluginRegisterPlugins(t *testing.T) {

	err := RegisterWrapperFactory("AWrapper", NewAWrapper)
	assert.Nil(t, err)

	var factories []WrapperFactory
	factory, err := LookupWrapperFactory("AWrapper")
	assert.Nil(t, err)

	plugins := ListPlugins()
	assert.Equal(t, 1, len(plugins))
	assert.True(t, RegistryContains("AWrapper"))

	factories = append(factories, factory)
	assert.Equal(t, 1, len(factories))
	handler := WrapHandlerFunc(handleCall, factories)

	adaptedHandler := &ContextAdapter{
		Ctx:     context.Background(),
		Handler: handler,
	}

	ts := httptest.NewServer(adaptedHandler)
	defer ts.Close()

	testURL := fmt.Sprintf("%s/foo", ts.URL)
	res, err := http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)

	rs, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.Nil(t, err)
	assert.True(t, strings.Contains(string(rs), "A wrapper wrote this"))
}

func TestPluginRegisterWrapperFactoryWithNoName(t *testing.T) {
	err := RegisterWrapperFactory("", NewAWrapper)
	assert.NotNil(t, err)
}

func TestPluginLookupUnregisteredWrapperFactory(t *testing.T) {
	_, err := LookupWrapperFactory("huh?")
	assert.NotNil(t, err)
}

func TestPluginWrapHandlerFunc(t *testing.T) {
	var factories []WrapperFactory
	factory, err := LookupWrapperFactory("AWrapper")
	assert.Nil(t, err)
	factories = append(factories, factory)

	hf := WrapHandlerFunc(handleCall, factories)
	assert.NotNil(t, hf)

	adaptedHandler := &ContextAdapter{
		Ctx:     context.Background(),
		Handler: hf,
	}

	ts := httptest.NewServer(adaptedHandler)
	defer ts.Close()

	testURL := fmt.Sprintf("%s/foo", ts.URL)
	res, err := http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)

	rs, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.Nil(t, err)
	assert.True(t, strings.Contains(string(rs), "A wrapper wrote this"))
}
