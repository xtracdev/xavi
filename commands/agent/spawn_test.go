package agent

//Note the happy path test coverage is achieved in setting up the framework acceptance tests

import (
	"fmt"
	"github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var wrappedSpawnListenerFn = wrap(testKVStore, NewAPIService(SpawnListenerDefCmd))
var wrappedSpawnKillerFn = wrap(testKVStore, NewAPIService(SpawnKillerDefCmd))

func TestSpawnListenerMissingPayload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedSpawnListenerFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/v1/spawn-listener/", ts.URL)
	resp, err := http.Post(testURL, "", nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSpawnListenerMalformedPayload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedSpawnListenerFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/v1/spawn-listener/", ts.URL)
	resp, err := http.Post(testURL, "", strings.NewReader("{ack ack"))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSpawnKillerMissingPid(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedSpawnKillerFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/v1/spawn-killer/", ts.URL)
	resp, err := http.Post(testURL, "", nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSpawnKillerMalformedPid(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedSpawnKillerFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/v1/spawn-killer/xxx77x9qqq", ts.URL)
	resp, err := http.Post(testURL, "", nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSpawnKillerNotMyPid(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedSpawnKillerFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/v1/spawn-killer/0", ts.URL)
	resp, err := http.Post(testURL, "", nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSpawnListenerNotImplemented(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedSpawnListenerFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/v1/spawn-listener", ts.URL)
	resp, err := http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

	testURL = fmt.Sprintf("%s/v1/spawn-listener/", ts.URL)
	resp, err = http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

	request, err := http.NewRequest("PUT", testURL, strings.NewReader("xxx"))
	assert.Nil(t, err)
	client := &http.Client{}
	response, err := client.Do(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, response.StatusCode)

}

func TestSpawnKillerNotImplemented(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrappedSpawnKillerFn))
	defer ts.Close()

	testURL := fmt.Sprintf("%s/v1/spawn-killer", ts.URL)
	resp, err := http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

	testURL = fmt.Sprintf("%s/v1/spawn-killer/", ts.URL)
	resp, err = http.Get(testURL)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

	request, err := http.NewRequest("PUT", testURL, strings.NewReader("xxx"))
	assert.Nil(t, err)
	client := &http.Client{}
	response, err := client.Do(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, response.StatusCode)
}

func TestPidTracking(t *testing.T) {
	assert.False(t, isSpawnedPid(-1))
	addPid(-1)
	assert.True(t, isSpawnedPid(-1))
	removePid(-1)
	assert.False(t, isSpawnedPid(-1))
}
