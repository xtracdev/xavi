package timing

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/plugin/logging"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func handleBar(rw http.ResponseWriter, req *http.Request) {
	timerCtx := TimerFromContext(req.Context())
	if timerCtx == nil {
		println("damn no context")
		http.Error(rw, "No context", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(200)
	rw.Write([]byte("foo"))
}

func TestServiceNameContextPresent(t *testing.T) {
	ctx := context.Background()
	ctx = AddServiceNameToContext(ctx, "foo")
	assert.Equal(t, "foo", GetServiceNameFromContext(ctx))
}

func TestServiceNameContextNotPresent(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, "", GetServiceNameFromContext(ctx))
}

func TestContextPresent(t *testing.T) {
	wrapperFactory := logging.NewLoggingWrapper()
	assert.NotNil(t, wrapperFactory)
	handler := wrapperFactory.Wrap(http.HandlerFunc(handleBar))

	timerWrapper := NewTimingWrapper()

	handler = timerWrapper.Wrap(handler)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	res, err := http.Post(ts.URL, "application/json", bytes.NewBuffer([]byte("Some stuff")))
	if !assert.NoError(t, err) {
		t.Log(err)
		t.Fail()
		return
	}

	resBytes, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	res.Body.Close()

	assert.Equal(t, "foo", string(resBytes))
	assert.Equal(t, http.StatusOK, res.StatusCode)

	//Delay to see the log output and to pick it up for test coverage
	time.Sleep(1 * time.Second)
}
