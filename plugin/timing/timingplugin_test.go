package timing

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/plugin"
	"github.com/xtracdev/xavi/plugin/logging"
	"github.com/xtracdev/xavi/timer"
	"golang.org/x/net/context"
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

func TestContextPresentConcurrently(t *testing.T) {
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

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
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
		}()
	}
	wg.Wait()
	//Delay to see the log output and to pick it up for test coverage
	time.Sleep(1 * time.Second)
}

func TestWriteTimingsToStatsdRaceCondition(t *testing.T) {
	eet := timer.NewEndToEndTimer("racy")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c := eet.StartContributor("fst")
		sc := c.StartServiceCall("foo call", "/foo/service")
		wg.Add(3)
		time.AfterFunc(1e9, func() { defer wg.Done(); sc.End(nil) })
		time.AfterFunc(2e9, func() { defer wg.Done(); c.End(fmt.Errorf("Error num %v", 1)) })
		time.AfterFunc(3e9, func() { defer wg.Done(); eet.Stop(nil) })
	}()
	wg.Add(1)
	go func() { defer wg.Done(); writeTimingsToStatsd(eet) }()
	wg.Wait()
	t.Logf("%s\n", eet.ToJSONString())
}
