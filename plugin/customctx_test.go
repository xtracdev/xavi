package plugin

import (
	"net/http"
	"golang.org/x/net/context"
"testing"
	"net/http/httptest"
	"github.com/stretchr/testify/assert")

const requestIDKey = -1

func newContextWithRequestID(ctx context.Context, req *http.Request) context.Context {
	return context.WithValue(ctx, requestIDKey, req.Header.Get("X-Request-ID"))
}

func requestIDFromContext(ctx context.Context) string {
	return ctx.Value(requestIDKey).(string)
}

func requestIdMiddleware(h ContextHandler) ContextHandler {
	return ContextHandlerFunc(func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
		ctx = newContextWithRequestID(ctx, req)
		println("-->Request id serve http")
		h.ServeHTTPContext(ctx, rw, req)
		println("<--Request id http served")
	})
}

var requestId = "NotFromContext"

func handler(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	requestId = requestIDFromContext(ctx)
}

func TestWithXRequestID(t *testing.T) {
	h := &ContextAdapter{
		Ctx:     context.Background(),
		Handler: requestIdMiddleware(ContextHandlerFunc(handler)),
	}

	ts := httptest.NewServer(h)
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL, nil)
	assert.Nil(t, err)

	req.Header.Set("X-Request-ID", "request-id")

	client := http.Client{}

	_, err = client.Do(req)
	assert.Nil(t, err)

	assert.Equal(t, "request-id", requestId)
}