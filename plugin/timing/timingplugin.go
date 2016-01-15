/*
Package timing provides a plugin Xavi wires in as the enclosing wrapper for the user specified
plugin chain. This plugin creates and puts an EndToEndTimer into the context that downstream
components may annotate with the service name and contributors of note. The JSON representation
of the timing is logged on completion of the wrapped call chain.
*/
package timing

import (
	"github.com/Sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"github.com/xtracdev/xavi/plugin"
	"github.com/xtracdev/xavi/timer"
	"golang.org/x/net/context"
	"net/http"
	"time"
	"expvar"
	"github.com/armon/go-metrics"
	_ "github.com/xtracdev/xavi/statsd"
)

type key int

const timerKey key = -22132

var counts = expvar.NewMap("counters")

//Use a separate timer to avoid escaping the JSON timing data - we want it to appear as
//JSON In the logfile.
var timerLog = logrus.New()

func init() {
	pf := new(prefixed.TextFormatter)
	pf.TimestampFormat = time.RFC3339
	timerLog.Formatter = pf
}

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

//Function to log timing data for later analysis
func logTiming(t *timer.EndToEndTimer) {
	timerLog.WithFields(logrus.Fields{
		"prefix": "timing-data",
	}).Info(t.ToJSONString())

	go func(t *timer.EndToEndTimer) {
		updateCounters(t)
	}(t)
}

//Function to modify epvar counters
func updateCounters(t *timer.EndToEndTimer) {
	if(t.ErrorFree) {
		counts.Add(t.Name, 1)
		metrics.IncrCounter([]string{t.Name},1.0)
		writeTimingsToStatsd(t)
	} else {
		counts.Add(t.Name + "-errors", 1)
	}
}

//Send timing data to statsd
func writeTimingsToStatsd(t *timer.EndToEndTimer) {
	metrics.AddSample([]string{t.Name}, float32(t.Time))
	for _, c := range t.Contributors {
		metrics.AddSample([]string{t.Name + ":" + c.Name}, float32(c.Time))
		for _,sc := range c.ServiceCalls {
			metrics.AddSample([]string{t.Name + ":" + c.Name + ":" + sc.Name }, float32(sc.Time))
		}
	}
}
