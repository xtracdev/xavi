/*
Package timing provides a plugin Xavi wires in as the enclosing wrapper for the user specified
plugin chain. This plugin creates and puts an EndToEndTimer into the context that downstream
components may annotate with the service name and contributors of note. The JSON representation
of the timing is logged on completion of the wrapped call chain.
*/
package timing

import (
	"expvar"
	"fmt"
	"github.com/armon/go-metrics"
	"github.com/xtracdev/xavi/plugin"
	_ "github.com/xtracdev/xavi/statsd"
	"github.com/xtracdev/xavi/timer"
	"golang.org/x/net/context"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"
)

type key int

const timerKey key = -22132
const serviceNameKey key = -22133

var counts = expvar.NewMap("counters")

//NewContextWithTimer adds a new timer to the request context
func NewContextWithTimer(ctx context.Context, req *http.Request) context.Context {
	timer := timer.NewEndToEndTimer("unspecified timer")
	return context.WithValue(ctx, timerKey, timer)
}

//AddServiceNameToContext adds the name of the service the backend handler will invoke. This provides
//a service name in the output timing log to allow the latency of different backend services to be
//assessed.
func AddServiceNameToContext(ctx context.Context, serviceName string) context.Context {
	return context.WithValue(ctx, serviceNameKey, serviceName)
}

//GetServiceNameFromContext pulls the service name from the context.
func GetServiceNameFromContext(ctx context.Context) string {
	serviceName, ok := ctx.Value(serviceNameKey).(string)
	if !ok {
		return ""
	}

	return serviceName
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

type TimingWrapper struct{}

func NewTimingWrapper() TimingWrapper {
	return TimingWrapper{}
}

//Wrap implements the plugin Wrapper interface, and is used
//to wrap a handler to put a EndToEndTimer instance into the call context
func (tw TimingWrapper) Wrap(h plugin.ContextHandler) plugin.ContextHandler {
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
	//We add a timestamp to the JSON to allow indexing in elasticsearch
	t.LoggingTimestamp = time.Now()

	fmt.Fprintln(os.Stderr, t.ToJSONString())

	go func(t *timer.EndToEndTimer) {
		updateCounters(t)
	}(t)
}

//Function to modify epvar counters
func updateCounters(t *timer.EndToEndTimer) {
	if t.ErrorFree {
		countName := spaceMap(t.Name + "-count")
		counts.Add(countName, 1)
		metrics.IncrCounter([]string{countName}, 1.0)
		writeTimingsToStatsd(t)
	} else {
		counts.Add(t.Name+"-errors", 1)
	}
}

func spaceMap(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

//Send timing data to statsd
func writeTimingsToStatsd(t *timer.EndToEndTimer) {
	metrics.AddSample([]string{spaceMap(t.Name)}, float32(t.Duration))
	for _, c := range t.Contributors {
		metrics.AddSample([]string{spaceMap(t.Name + "." + c.Name)}, float32(c.Duration))
		for _, sc := range c.ServiceCalls {
			metrics.AddSample([]string{spaceMap(t.Name + "." + c.Name + "." + sc.Name)}, float32(sc.Duration))
		}
	}
}
