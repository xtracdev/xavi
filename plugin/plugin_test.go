package plugin

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

//NewAWrapper instantiates AWrapper
func NewAWrapper() Wrapper {
	return new(AWrapper)
}

//AWrapper can wrap http handlers
type AWrapper struct{}

//Wrap wraps http.Handlers with A stuff
func (aw AWrapper) Wrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
		w.Write([]byte("A wrapper wrote this\n"))
	})
}

func handleCall(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("handleCall wrote this stuff\n"))
}

func TestRegisterPlugins(t *testing.T) {

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
	handler := ChainWrappers(handleCall, factories)
	ts := httptest.NewServer(handler)
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

func TestRegisterWrapperFactoryWithNoName(t *testing.T) {
	err := RegisterWrapperFactory("", NewAWrapper)
	assert.NotNil(t, err)
}

func TestLookupUnregisteredWrapperFactory(t *testing.T) {
	_, err := LookupWrapperFactory("huh?")
	assert.NotNil(t, err)
}

func TestWrapHandlerFunc(t *testing.T) {
	var factories []WrapperFactory
	factory, err := LookupWrapperFactory("AWrapper")
	assert.Nil(t, err)
	factories = append(factories, factory)

	hf := WrapHandlerFunc(handleCall, factories)
	assert.NotNil(t, hf)

	ts := httptest.NewServer(hf)
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

func TestChainWrappersAroundHandler(t *testing.T) {
	var factories []WrapperFactory
	factory, err := LookupWrapperFactory("AWrapper")
	assert.Nil(t, err)
	factories = append(factories, factory)

	hf := ChainWrappersAroundHandler(http.HandlerFunc(handleCall), factories)
	assert.NotNil(t, hf)

	ts := httptest.NewServer(hf)
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
