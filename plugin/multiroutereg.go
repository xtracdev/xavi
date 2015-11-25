package plugin

import "fmt"

var registeredMBAdaptorFactories map[string]MultiBackendAdapterFactory

func init() {
	registeredMBAdaptorFactories = make(map[string]MultiBackendAdapterFactory)
}

//ListMultiBackendAdapters lists the multi backend adapter factories that have been registered.
func ListMultiBackendAdapters() []string {
	var adapters []string
	for key := range registeredMBAdaptorFactories {
		adapters = append(adapters, key)
	}
	return adapters
}

//MultiBackendAdapterRegistryContains is a predicate that returns true if the named
//multi backend adapter factory has been registered
func MultiBackendAdapterRegistryContains(name string) bool {
	_, ok := registeredMBAdaptorFactories[name]
	return ok
}

//RegisterMultiBackendAdapterFactory registers the given factory function using the given name.
func RegisterMultiBackendAdapterFactory(name string, factory MultiBackendAdapterFactory) error {
	if name == "" {
		return fmt.Errorf("Empty name passed to RegisterMRAFactory")
	}

	registeredMBAdaptorFactories[name] = factory
	return nil
}

//LookupMultiBackendAdapterFactory returns the factory function registered for the given
//name.
func LookupMultiBackendAdapterFactory(name string) (MultiBackendAdapterFactory, error) {
	factory, ok := registeredMBAdaptorFactories[name]
	if !ok {
		return nil, fmt.Errorf("Factory %s not registered", name)
	}

	return factory, nil
}
