/*
Package timing provides a plugin Xavi wires in as the enclosing wrapper for the user specified
plugin chain. This plugin creates and puts an EndToEndTimer into the context that downstream
components may annotate with the service name and contributors of note. The JSON representation
of the timing is logged on completion of the wrapped call chain.
*/
package timing

import (
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/plugin"
	"github.com/xtracdev/xavi/timer"
	"golang.org/x/net/context"
	"net/http"
)

type key int

const timerKey key = -22132

//NewContextWithTimer adds a new timer to the request context
func NewContextWithTimer(ctx context.Context, req *http.Request) context.Context {
	timer := timer.NewEndToEndTimer("unspecified timer")
	return context.WithValue(ctx, timerKey, timer)
}

//TimerFromContext returns an EndToEndTimer from the given context if one
//is present, otherwise nil is returned
func TimerFromContext(ctx context.Context) *timer.EndToEndTimer {
	newCtx, ok := ctx.Value(timerKey).(*timer.EndToEndTimer)
	if !ok {
		return nil
	}

	return newCtx
}

//RequestTimerMiddleware implements the plugin Wrapper interface, and is used
//to wrap a handler to put a EndToEndTimer instance into the call context
func RequestTimerMiddleware(h plugin.ContextHandler) plugin.ContextHandler {
	return plugin.ContextHandlerFunc(func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
		ctx = NewContextWithTimer(ctx, req)
		h.ServeHTTPContext(ctx, rw, req)
		ctxTimer := TimerFromContext(ctx)
		ctxTimer.Stop(nil)
		go func(t *timer.EndToEndTimer) {
			logTiming(t)
		}(ctxTimer)
	})
}

//Method to log the timing data to standard out.
func logTiming(t *timer.EndToEndTimer) {
	log.Info(t.ToJSONString())
}
