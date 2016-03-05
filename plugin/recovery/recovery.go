package recovery

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/plugin"
	"golang.org/x/net/context"
	"net/http"
)

//RecoveryContext defines a structure to allow a logging function and error message formulation
//function to be used when handling a panic produced when servicing an http route.
type RecoveryContext struct {
	LogFn          func(interface{})
	ErrorMessageFn func(interface{}) string
}


//defaultGlobalRecoveryContext defines the default logging and error message when handling a panic
//produced servicing an http route
var defaultGlobalRecoveryContext = &RecoveryContext{
	LogFn: func(r interface{}) {
		var err error
		switch t := r.(type) {
		case string:
			err = errors.New(t)
		case error:
			err = t
		default:
			err = errors.New("Unknown error")
		}
		log.Warn("Handled panic: ", err.Error())
	},
	ErrorMessageFn: func(r interface{}) string {
		return ""
	},
}

//GlobalPanicRecoveryMiddleware defines a middleware that xavi wraps the http call chain in, such that a
//panic that occurs in the service handling gets logged and an error status is returned to the client.
//If a recovery context is provided, the logging and error message used in the panic handling is derived
//using the functions passed in the recovery context. Otherwise, the default logging and error message
//is used.
func GlobalPanicRecoveryMiddleware(rc *RecoveryContext, h plugin.ContextHandler) plugin.ContextHandler {
	return plugin.ContextHandlerFunc(func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
		if rc == nil {
			rc = defaultGlobalRecoveryContext
		}

		defer func() {
			r := recover()
			if r != nil {
				rc.LogFn(r)
				http.Error(rw, rc.ErrorMessageFn(r), http.StatusInternalServerError)
			}
		}()
		h.ServeHTTPContext(ctx, rw, req)
	})
}
