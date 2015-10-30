package plugin

import (
	"fmt"
	"net/http"
)

var registeredWrapperFactories map[string]WrapperFactory

func init() {
	registeredWrapperFactories = make(map[string]WrapperFactory)
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
func RegisterWrapperFactory(name string, factory WrapperFactory) error {
	if name == "" {
		return fmt.Errorf("Empty name passed to RegisterWrapperFactory")
	}

	registeredWrapperFactories[name] = factory
	return nil
}

//LookupWrapperFactory looks up the named wrapper factory in the
//registry, returning an error if the factory is not registered.
func LookupWrapperFactory(name string) (WrapperFactory, error) {
	factory, ok := registeredWrapperFactories[name]
	if !ok {
		return nil, fmt.Errorf("Factory %s not registered", name)
	}

	return factory, nil
}

//Wrapper defines an interface for things that can wrap http Handlers
type Wrapper interface {
	Wrap(http.Handler) http.Handler
}

//WrapperFactory defines a function that can create something that
//implements Wrapper
type WrapperFactory func() Wrapper

//ChainWrappers wraps the given handler function with wrappers instantiated from
//the slice of wrapper factories. The order of factories in the slice is
//significant; the lowest indexed wrapper in the innermost, the highest
//the outermost.
func ChainWrappers(hf func(w http.ResponseWriter, r *http.Request), wrapperFactories []WrapperFactory) http.Handler {
	handler := http.HandlerFunc(hf)
	for _, factory := range wrapperFactories {
		if factory == nil {
			continue
		}
		wrapper := factory()
		handler = (wrapper.Wrap(handler)).(http.HandlerFunc)
	}

	return handler
}

func WrapHandlerFunc(hf http.HandlerFunc, wrapperFactories []WrapperFactory) http.HandlerFunc {
	handler := hf
	for _, factory := range wrapperFactories {
		if factory == nil {
			continue
		}
		wrapper := factory()
		handler = (wrapper.Wrap(handler)).(http.HandlerFunc)
	}

	return handler
}

//ChainWrappersAroundHandler wraps the given handler with wrappers created via
//the passed slice of wrapper factories.
func ChainWrappersAroundHandler(handler http.Handler, wrapperFactories []WrapperFactory) http.Handler {
	for _, factory := range wrapperFactories {
		if factory == nil {
			continue
		}
		wrapper := factory()
		handler = (wrapper.Wrap(handler))
	}

	return handler
}
