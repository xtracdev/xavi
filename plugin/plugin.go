package plugin

import (
	"fmt"
	"net/http"
)

var registeredWrapperFactories map[string]*WrapperFactoryContext

type WrapperFactoryContext struct {
	factory WrapperFactory
	args    []interface{}
}

func init() {
	registeredWrapperFactories = make(map[string]*WrapperFactoryContext)
}

//ListPlugins lists the plugins currently registered with
//the package.
func ListPlugins() []string {
	var plugins []string
	for key := range registeredWrapperFactories {
		plugins = append(plugins, key)
	}
	return plugins
}

//RegistryContains is a predicate that indicates if the named
//plugin is registered with this package
func RegistryContains(name string) bool {
	_, ok := registeredWrapperFactories[name]
	return ok
}

//RegisterWrapperFactory is a method for registering wrapper factories
//with the package.
func RegisterWrapperFactory(name string, factory WrapperFactory, args ...interface{}) error {
	if name == "" {
		return fmt.Errorf("Empty name passed to RegisterWrapperFactory")
	}

	context := &WrapperFactoryContext{
		factory: factory,
		args:    args,
	}

	registeredWrapperFactories[name] = context
	return nil
}

//LookupWrapperFactoryCtx looks up the named wrapper factory in the
//registry, returning an error if the factory is not registered.
func LookupWrapperFactoryCtx(name string) (*WrapperFactoryContext, error) {
	factoryCtx, ok := registeredWrapperFactories[name]
	if !ok {
		return nil, fmt.Errorf("Factory %s not registered", name)
	}

	return factoryCtx, nil
}

//Wrapper defines an interface for things that can wrap http Handlers
type Wrapper interface {
	Wrap(http.Handler) http.Handler
}

//WrapperFactory defines a function that can create something that
//implements Wrapper
type WrapperFactory func(...interface{}) Wrapper

//WrapHandlerFunc wraps a handler function, which is instantiated using the wrapper
//factory and arguments to the wrapper factory stored in the wrapper factory context
func WrapHandlerFunc(hf http.HandlerFunc, wrapperFactories []*WrapperFactoryContext) http.HandlerFunc {
	handler := hf
	for _, factoryCtx := range wrapperFactories {
		if factoryCtx == nil {
			continue
		}
		wrapper := factoryCtx.factory(factoryCtx.args...)
		handler = (wrapper.Wrap(handler)).(http.HandlerFunc)
	}

	return handler
}
