package recovery

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func handleBar(rw http.ResponseWriter, req *http.Request) {
	panic("Kaboom")
}

func TestSuppliedContextHandler(t *testing.T) {

	logged := false
	errorMsg := false

	rc := RecoveryContext{
		LogFn: func(r interface{}) { logged = true },
		ErrorMessageFn: func(r interface{}) string {
			errorMsg = true
			return ""
		},
	}

	recoveryWrapper := RecoveryWrapper{RecoveryContext: rc}

	handler := recoveryWrapper.Wrap(http.HandlerFunc(handleBar))

	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.True(t, logged)
	assert.True(t, errorMsg)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestDefaultContextHandler(t *testing.T) {
	recoveryWrapper := NewRecoveryWrapper()

	handler := recoveryWrapper.Wrap(http.HandlerFunc(handleBar))

	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
