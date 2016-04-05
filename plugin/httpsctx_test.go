package plugin

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"net/http"
	"net/http/httptest"
	"testing"
)

var contextVal bool

func httpsContextHandler(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	contextVal = GetUseHttpsContext(ctx)
	rw.WriteHeader(http.StatusOK)
}

func TestWithHttpsCtx(t *testing.T) {
	ctx := context.Background()
	ctx = AddUseHttpsToContext(ctx, true)

	h := &ContextAdapter{
		Ctx:     ctx,
		Handler: requestIdMiddleware(ContextHandlerFunc(httpsContextHandler)),
	}

	ts := httptest.NewServer(h)
	defer ts.Close()

	_, err := http.Get(ts.URL)

	if assert.Nil(t, err) {
		assert.True(t, contextVal)
	}

}

func TestAddUseContext(t *testing.T) {
	ctx := context.Background()
	ctx = AddUseHttpsToContext(ctx, true)
	useHttpFromCtx := GetUseHttpsContext(ctx)
	assert.True(t, useHttpFromCtx)
}
