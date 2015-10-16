package agent

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/kvstore"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var testKVStore = testMakeAndInitializeKVStore()

func testMakeAndInitializeKVStore() *kvstore.HashKVStore {
	kvs, _ := kvstore.NewHashKVStore("")
	return kvs
}

var wrappedFn = wrap(testKVStore, NewAPIService(ServerDefCmd))

func TestServerGetEmptyList(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(wrappedFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/v1/servers/", ts.URL)
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

func TestServerPutMissingResource(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedFn))
	defer ts.Close()

	testPayload := `
	{
"Address":"localhost",
"Port":9876,
"PingURI":"/hello",
"healthCheck":"none",
"healthCheckInterval":30,
"healthCheckTimeout":10
}
	`

	testURL := fmt.Sprintf("%s/v1/servers/", ts.URL)
	request, err := http.NewRequest("PUT", testURL, strings.NewReader(testPayload))
	assert.Nil(t, err)
	client := &http.Client{}
	response, err := client.Do(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestServerPutMalformedBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedFn))
	defer ts.Close()

	testPayload := `
	{
"Address":"localhost",
"Port":9876
yo mama
"PingURI":"/hello","healthCheck":"none", "healthCheckInterval":30, "healthCheckTimeout":10
	`

	testURL := fmt.Sprintf("%s/v1/servers/test-name", ts.URL)
	request, err := http.NewRequest("PUT", testURL, strings.NewReader(testPayload))
	assert.Nil(t, err)
	client := &http.Client{}
	response, err := client.Do(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestServerPutWIthKVSFault(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedFn))
	defer ts.Close()

	testKVStore.InjectFaults()
	defer testKVStore.ClearFaults()

	testPayload := `
	{
"Address":"localhost",
"Port":9876,
"PingURI":"/hello","healthCheck":"none", "healthCheckInterval":30, "healthCheckTimeout":10
}
	`

	testURL := fmt.Sprintf("%s/v1/servers/test-name", ts.URL)
	request, err := http.NewRequest("PUT", testURL, strings.NewReader(testPayload))
	assert.Nil(t, err)
	client := &http.Client{}
	response, err := client.Do(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestServerPut(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedFn))
	defer ts.Close()

	testPayload := `
	{
"Address":"localhost",
"Port":9876,
"PingURI":"/hello","healthCheck":"none", "healthCheckInterval":30, "healthCheckTimeout":10
}
	`

	testURL := fmt.Sprintf("%s/v1/servers/test-name", ts.URL)
	request, err := http.NewRequest("PUT", testURL, strings.NewReader(testPayload))
	assert.Nil(t, err)
	client := &http.Client{}
	response, err := client.Do(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestServerGet(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(wrappedFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/v1/servers/test-name", ts.URL)
	res, err := http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)

	rs, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.Nil(t, err)

	expected := `{"Name":"test-name","Address":"localhost","Port":9876,"PingURI":"/hello","HealthCheck":"none","HealthCheckInterval":30,"HealthCheckTimeout":10}`

	responseString := string(rs)
	assert.Equal(t, expected, responseString)

}

func TestServerGeKVSFault(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(wrappedFn))
	defer ts.Close()

	testKVStore.InjectFaults()
	defer testKVStore.ClearFaults()

	testURL := fmt.Sprintf("%s/v1/servers/test-name", ts.URL)
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

func TestServerGetList(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(wrappedFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/v1/servers/", ts.URL)
	res, err := http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)

	rs, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.Nil(t, err)

	expected := `[{"Name":"test-name","Address":"localhost","Port":9876,"PingURI":"/hello","HealthCheck":"none","HealthCheckInterval":30,"HealthCheckTimeout":10}]`

	responseString := string(rs)
	assert.Equal(t, expected, responseString)

}

func TestServerGetListKVStoreFault(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(wrappedFn))
	defer ts.Close()

	testKVStore.InjectFaults()
	defer testKVStore.ClearFaults()

	testURL := fmt.Sprintf("%s/v1/servers/", ts.URL)
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

func TestNotHandled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/servers/test-name", ts.URL)
	res, err := http.Head(testURL)
	assert.Nil(t, err)
	assert.Equal(t, 405, res.StatusCode)
}

func TestServerNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/servers/test-nobody-here", ts.URL)
	res, err := http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, 404, res.StatusCode)
}
