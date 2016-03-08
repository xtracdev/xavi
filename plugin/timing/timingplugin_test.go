package timing

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/plugin"
	"github.com/xtracdev/xavi/plugin/logging"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func handleBar(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	timerCtx := TimerFromContext(ctx)
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
	handler := wrapperFactory.Wrap(plugin.ContextHandlerFunc(handleBar))

	timerWrapper := NewTimingWrapper()

	handler = timerWrapper.Wrap(handler)

	adapter := &plugin.ContextAdapter{
		Ctx:     context.Background(),
		Handler: handler,
	}

	ts := httptest.NewServer(adapter)
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
