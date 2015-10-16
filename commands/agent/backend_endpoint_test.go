package agent

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var wrappedBackendFn = wrap(testKVStore, NewAPIService(BackendDefCmd))

func TestBackendGetEmptyList(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(wrappedBackendFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/v1/backends/", ts.URL)
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

func TestBackendPutMissingResource(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedBackendFn))
	defer ts.Close()

	testPayload := `
	{
"ServerNames":["server1","server2"]
}
	`

	testURL := fmt.Sprintf("%s/v1/backends/", ts.URL)
	request, err := http.NewRequest("PUT", testURL, strings.NewReader(testPayload))
	assert.Nil(t, err)
	client := &http.Client{}
	response, err := client.Do(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestBackendPutMalformedBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedBackendFn))
	defer ts.Close()

	testPayload := `
	{
"ServerNames":["server1","server2"]
	`

	testURL := fmt.Sprintf("%s/v1/backends/test-name", ts.URL)
	request, err := http.NewRequest("PUT", testURL, strings.NewReader(testPayload))
	assert.Nil(t, err)
	client := &http.Client{}
	response, err := client.Do(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestBackendPutWIthKVSFault(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedBackendFn))
	defer ts.Close()

	testKVStore.InjectFaults()
	defer testKVStore.ClearFaults()

	testPayload := `
	{
"ServerNames":["server1","server2"]
}
	`

	testURL := fmt.Sprintf("%s/v1/backends/test-name", ts.URL)
	request, err := http.NewRequest("PUT", testURL, strings.NewReader(testPayload))
	assert.Nil(t, err)
	client := &http.Client{}
	response, err := client.Do(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestBackendPut(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedBackendFn))
	defer ts.Close()

	testPayload := `
	{
"ServerNames":["server1","server2"]
}
	`

	testURL := fmt.Sprintf("%s/v1/backends/test-name", ts.URL)
	request, err := http.NewRequest("PUT", testURL, strings.NewReader(testPayload))
	assert.Nil(t, err)
	client := &http.Client{}
	response, err := client.Do(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestBackendGet(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(wrappedBackendFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/v1/backends/test-name", ts.URL)
	res, err := http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)

	rs, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.Nil(t, err)

	expected := `{"Name":"test-name","ServerNames":["server1","server2"],"LoadBalancerPolicy":""}`

	responseString := string(rs)
	assert.Equal(t, expected, responseString)

}

func TestBackendGeKVSFault(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(wrappedBackendFn))
	defer ts.Close()

	testKVStore.InjectFaults()
	defer testKVStore.ClearFaults()

	testURL := fmt.Sprintf("%s/v1/backends/test-name", ts.URL)
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

func TestBackendGetList(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(wrappedBackendFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/v1/backends/", ts.URL)
	res, err := http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)

	rs, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.Nil(t, err)

	expected := `[{"Name":"test-name","ServerNames":["server1","server2"],"LoadBalancerPolicy":""}]`

	responseString := string(rs)
	assert.Equal(t, expected, responseString)

}

func TestBackendGetListKVStoreFault(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(wrappedBackendFn))
	defer ts.Close()

	testKVStore.InjectFaults()
	defer testKVStore.ClearFaults()

	testURL := fmt.Sprintf("%s/v1/backends/", ts.URL)
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

func TestBackendNotHandled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedBackendFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/backends/test-name", ts.URL)
	res, err := http.Head(testURL)
	assert.Nil(t, err)
	assert.Equal(t, 405, res.StatusCode)
}

func TestBackendNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedBackendFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/backends/test-nobody-here", ts.URL)
	res, err := http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, 404, res.StatusCode)
}
