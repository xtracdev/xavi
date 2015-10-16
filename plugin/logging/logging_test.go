package logging

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var requestBytes []byte

func handleFoo(rw http.ResponseWriter, req *http.Request) {

	requestBytes, _ = ioutil.ReadAll(req.Body)
	req.Body.Close()

	rw.WriteHeader(200)
	rw.Write([]byte("foo"))

}

func TestLoggingFilterPreservesIO(t *testing.T) {

	wrapperFactory := NewLoggingWrapper()
	assert.NotNil(t, wrapperFactory)
	handler := wrapperFactory.Wrap(http.HandlerFunc(handleFoo))

	ts := httptest.NewServer(handler)
	defer ts.Close()

	res, err := http.Post(ts.URL, "application/json", bytes.NewBuffer([]byte("Some stuff")))
	assert.NoError(t, err)

	assert.Equal(t, []byte("Some stuff"), requestBytes)

	resBytes, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	res.Body.Close()

	assert.Equal(t, "foo", string(resBytes))
}
