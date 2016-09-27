package plugin

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var contextVal bool

func httpsHandler(rw http.ResponseWriter, req *http.Request) {
	contextVal = GetUseHttpsContext(req.Context())
	rw.WriteHeader(http.StatusNoContent)
}

func requestIdMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		req = req.WithContext(AddUseHttpsToContext(req.Context(), true))
		println("-->Request id serve http")
		h.ServeHTTP(rw, req)
		println("<--Request id http served")
	})
}

func TestWithHttpsCtx(t *testing.T) {
	ts := httptest.NewServer(requestIdMiddleware(http.HandlerFunc(httpsHandler)))
	defer ts.Close()

	resp, err := http.Get(ts.URL)

	assert.Nil(t, err)
	assert.True(t, contextVal)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestAddUseContext(t *testing.T) {
	ctx := context.Background()
	ctx = AddUseHttpsToContext(ctx, true)
	useHttpFromCtx := GetUseHttpsContext(ctx)
	assert.True(t, useHttpFromCtx)
}
