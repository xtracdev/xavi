package recovery

import (
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/plugin"
	"golang.org/x/net/context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func handleBar(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	panic("Kaboom")
}

func TestSuppliedContextHandler(t *testing.T) {

	logged := false
	errorMsg := false

	rc := &RecoveryContext{
		LogFn: func(r interface{}) { logged = true },
		ErrorMessageFn: func(r interface{}) string {
			errorMsg = true
			return ""
		},
	}

	handler := GlobalPanicRecoveryMiddleware(rc, plugin.ContextHandlerFunc(handleBar))

	adapter := &plugin.ContextAdapter{
		Ctx:     context.Background(),
		Handler: handler,
	}

	ts := httptest.NewServer(adapter)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.True(t, logged)
	assert.True(t, errorMsg)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestDefaultContextHandler(t *testing.T) {
	handler := GlobalPanicRecoveryMiddleware(nil, plugin.ContextHandlerFunc(handleBar))

	adapter := &plugin.ContextAdapter{
		Ctx:     context.Background(),
		Handler: handler,
	}

	ts := httptest.NewServer(adapter)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
