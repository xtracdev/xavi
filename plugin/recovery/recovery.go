package recovery

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/plugin"
	"net/http"
)

//RecoveryContext defines a structure to allow a logging function and error message formulation
//function to be used when handling a panic produced when servicing an http route.
type RecoveryContext struct {
	LogFn          func(interface{})
	ErrorMessageFn func(interface{}) string
}

//RecoveryWrapper is a plugin type used for adding recovery handling as a plugin
type RecoveryWrapper struct {
	RecoveryContext RecoveryContext
}

//NewRecoveryWrapper is the factory function for instantiating a RecoveryWrapper. Note that anyone wanting
//to customize the RecoveryWrapper with their own logging and error message functions will need to provide
//their own factory method
func NewRecoveryWrapper(args ...interface{}) plugin.Wrapper {
	return &RecoveryWrapper{
		RecoveryContext: defaultRecoveryContext,
	}
}

//defaultGlobalRecoveryContext defines the default logging and error message when handling a panic
//produced servicing an http route
var defaultRecoveryContext = RecoveryContext{
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

//Wrap wraps the context handler with panic recovery capability.
func (rcw RecoveryWrapper) Wrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rc := rcw.RecoveryContext

		defer func() {
			r := recover()
			if r != nil {
				rc.LogFn(r)
				http.Error(rw, rc.ErrorMessageFn(r), http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(rw, req)
	})
}
