package agent

import (
	"fmt"
	"github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var wrappedListenerFn = wrap(testKVStore, NewAPIService(ListenerDefCmd))

func TestListenerGetEmptyList(t *testing.T) {

	t.Log("test get listener list with no listeners in KV store")

	ts := httptest.NewServer(http.HandlerFunc(wrappedListenerFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/v1/listeners/", ts.URL)
	res, err := http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)

	rs, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.Nil(t, err)

	expected := `[]`

	responseString := string(rs)
	assert.Equal(t, expected, responseString)

}

func TestListenerPutMissingResource(t *testing.T) {

	t.Log("test put listener with missing resource")

	ts := httptest.NewServer(http.HandlerFunc(wrappedListenerFn))
	defer ts.Close()

	testPayload := `
	{"Name":"test-listener","RouteNames":["demo-route"]}
	`

	testURL := fmt.Sprintf("%s/v1/listeners/", ts.URL)
	request, err := http.NewRequest("PUT", testURL, strings.NewReader(testPayload))
	assert.Nil(t, err)
	client := &http.Client{}
	response, err := client.Do(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestListenerPutMalformedBody(t *testing.T) {

	t.Log("test put listener with malformed body")

	ts := httptest.NewServer(http.HandlerFunc(wrappedListenerFn))
	defer ts.Close()

	testPayload := `
	{"Name":"test-listener","RouteNames":["demo-route"]
	`

	testURL := fmt.Sprintf("%s/v1/listeners/test-listener", ts.URL)
	request, err := http.NewRequest("PUT", testURL, strings.NewReader(testPayload))
	assert.Nil(t, err)
	client := &http.Client{}
	response, err := client.Do(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestListenerPutWIthKVSFault(t *testing.T) {

	t.Log("test put listener with KV Store fault")

	ts := httptest.NewServer(http.HandlerFunc(wrappedListenerFn))
	defer ts.Close()

	testKVStore.InjectFaults()
	defer testKVStore.ClearFaults()

	testPayload := `
	{"Name":"test-listener","RouteNames":["demo-route"]}
	`

	testURL := fmt.Sprintf("%s/v1/listeners/test-listener", ts.URL)
	request, err := http.NewRequest("PUT", testURL, strings.NewReader(testPayload))
	assert.Nil(t, err)
	client := &http.Client{}
	response, err := client.Do(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestListenerPut(t *testing.T) {

	t.Log("test put listener")
	ts := httptest.NewServer(http.HandlerFunc(wrappedListenerFn))
	defer ts.Close()

	testPayload := `
	{"Name":"test-listener","RouteNames":["demo-route"]}
	`

	testURL := fmt.Sprintf("%s/v1/listeners/test-listener", ts.URL)
	request, err := http.NewRequest("PUT", testURL, strings.NewReader(testPayload))
	assert.Nil(t, err)
	client := &http.Client{}
	response, err := client.Do(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestListenerGet(t *testing.T) {

	t.Log("test get listener")

	ts := httptest.NewServer(http.HandlerFunc(wrappedListenerFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/v1/listeners/test-listener", ts.URL)
	res, err := http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)

	rs, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.Nil(t, err)

	expected := `{"Name":"test-listener","RouteNames":["demo-route"]}`

	responseString := string(rs)
	assert.Equal(t, expected, responseString)

}

func TestListenerGeKVSFault(t *testing.T) {

	t.Log("test get listener kvstore fault")

	ts := httptest.NewServer(http.HandlerFunc(wrappedListenerFn))
	defer ts.Close()

	testKVStore.InjectFaults()
	defer testKVStore.ClearFaults()

	testURL := fmt.Sprintf("%s/v1/listeners/test-listener", ts.URL)
	res, err := http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	rs, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.Nil(t, err)

	expected := `Faulty store does not get ur key, ok?`

	responseString := string(rs)
	assert.Equal(t, expected, responseString)

}

func TestListenerGetList(t *testing.T) {

	t.Log("test listener get list")

	ts := httptest.NewServer(http.HandlerFunc(wrappedListenerFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/v1/listeners/", ts.URL)
	res, err := http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)

	rs, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.Nil(t, err)

	expected := `[{"Name":"test-listener","RouteNames":["demo-route"]}]`

	responseString := string(rs)
	assert.Equal(t, expected, responseString)

}

func TestListenerGetListKVStoreFault(t *testing.T) {

	t.Log("test listener get list with kv store fault")

	ts := httptest.NewServer(http.HandlerFunc(wrappedListenerFn))
	defer ts.Close()

	testKVStore.InjectFaults()
	defer testKVStore.ClearFaults()

	testURL := fmt.Sprintf("%s/v1/listeners/", ts.URL)
	res, err := http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	rs, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.Nil(t, err)

	expected := `You can haz list? Nope.`

	responseString := string(rs)
	assert.Equal(t, expected, responseString)

}

func TestListenerNotHandled(t *testing.T) {

	t.Log("test listener with non handled url")
	ts := httptest.NewServer(http.HandlerFunc(wrappedListenerFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/listeners/test-listener", ts.URL)
	res, err := http.Head(testURL)
	assert.Nil(t, err)
	assert.Equal(t, 405, res.StatusCode)
}

func TestListenerNotFound(t *testing.T) {
	t.Log("test get listener with no such resource")
	ts := httptest.NewServer(http.HandlerFunc(wrappedListenerFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/listeners/test-nobody-here", ts.URL)
	res, err := http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, 404, res.StatusCode)
}
