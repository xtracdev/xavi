package logging

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/plugin"
)

//NewLoggingWrapper is a wrapper factory function that returns
//a new instance of a LoggingWrapper
func NewLoggingWrapper() plugin.Wrapper {
	return new(LoggingWrapper)
}

//LoggingWrapper defines a directive for capturing HTTP requests and responses in the logs
type LoggingWrapper struct{}

//Wrap wraps the given handler with some logging functionality. This should
//be update to log async in a go routine if it is to be used in production
//configuration.
func (lw LoggingWrapper) Wrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
		log.Info(fmt.Sprintf("request for %v with method %v", r.RequestURI,
			r.Method))
	})
}
