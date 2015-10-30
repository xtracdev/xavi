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

var wrappedRouteFn = wrap(testKVStore, NewAPIService(RouteDefCmd))

func TestRouteGetEmptyList(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(wrappedRouteFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/v1/routes/", ts.URL)
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

func TestRoutePutMissingResource(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedRouteFn))
	defer ts.Close()

	testPayload := `
	{"Name":"test-route","URIRoot":"/hello","Backend":"demo-backend","Plugins":null,"MsgProps":""}
	`

	testURL := fmt.Sprintf("%s/v1/routes/", ts.URL)
	request, err := http.NewRequest("PUT", testURL, strings.NewReader(testPayload))
	assert.Nil(t, err)
	client := &http.Client{}
	response, err := client.Do(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestRoutePutMalformedBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedRouteFn))
	defer ts.Close()

	testPayload := `
	{"Name":"test-route","URIRoot":"/hello","Backend":"demo-backend","Plugins":null,"MsgProps":""
	`

	testURL := fmt.Sprintf("%s/v1/routes/test-route", ts.URL)
	request, err := http.NewRequest("PUT", testURL, strings.NewReader(testPayload))
	assert.Nil(t, err)
	client := &http.Client{}
	response, err := client.Do(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestRoutePutWIthKVSFault(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedRouteFn))
	defer ts.Close()

	testKVStore.InjectFaults()
	defer testKVStore.ClearFaults()

	testPayload := `
	{"Name":"test-route","URIRoot":"/hello","Backend":"demo-backend","Plugins":null,"MsgProps":""}
	`

	testURL := fmt.Sprintf("%s/v1/routes/test-route", ts.URL)
	request, err := http.NewRequest("PUT", testURL, strings.NewReader(testPayload))
	assert.Nil(t, err)
	client := &http.Client{}
	response, err := client.Do(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestRoutePut(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedRouteFn))
	defer ts.Close()

	testPayload := `
	{"Name":"test-route","URIRoot":"/hello","Backend":"demo-backend","Plugins":null,"MsgProps":""}
	`

	testURL := fmt.Sprintf("%s/v1/routes/test-route", ts.URL)
	request, err := http.NewRequest("PUT", testURL, strings.NewReader(testPayload))
	assert.Nil(t, err)
	client := &http.Client{}
	response, err := client.Do(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestRouteGet(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(wrappedRouteFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/v1/routes/test-route", ts.URL)
	res, err := http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)

	rs, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.Nil(t, err)

	expected := `{"Name":"test-route","URIRoot":"/hello","Backend":"demo-backend","Plugins":null,"MsgProps":""}`

	responseString := string(rs)
	assert.Equal(t, expected, responseString)

}

func TestRouteGeKVSFault(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(wrappedRouteFn))
	defer ts.Close()

	testKVStore.InjectFaults()
	defer testKVStore.ClearFaults()

	testURL := fmt.Sprintf("%s/v1/routes/test-route", ts.URL)
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

func TestRouteGetList(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(wrappedRouteFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/v1/routes/", ts.URL)
	res, err := http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)

	rs, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.Nil(t, err)

	expected := `[{"Name":"test-route","URIRoot":"/hello","Backend":"demo-backend","Plugins":null,"MsgProps":""}]`

	responseString := string(rs)
	assert.Equal(t, expected, responseString)

}

func TestRouteGetListKVStoreFault(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(wrappedRouteFn))
	defer ts.Close()

	testKVStore.InjectFaults()
	defer testKVStore.ClearFaults()

	testURL := fmt.Sprintf("%s/v1/routes/", ts.URL)
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

func TestRouteNotHandled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedRouteFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/routes/test-route", ts.URL)
	res, err := http.Head(testURL)
	assert.Nil(t, err)
	assert.Equal(t, 405, res.StatusCode)
}

func TestRouteNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedRouteFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/routes/test-nobody-here", ts.URL)
	res, err := http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, 404, res.StatusCode)
}
