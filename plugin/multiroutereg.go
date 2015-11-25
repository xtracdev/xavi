package plugin

import "fmt"

var registeredMBAdaptorFactories map[string]MultiBackendAdapterFactory

func init() {
	registeredMBAdaptorFactories = make(map[string]MultiBackendAdapterFactory)
}

func ListMultiBackendAdapters() []string {
	var adapters []string
	for key := range registeredMBAdaptorFactories {
		adapters = append(adapters, key)
	}
	return adapters
}

func MultiBackendAdapterRegistryContains(name string) bool {
	_, ok := registeredMBAdaptorFactories[name]
	return ok
}

func RegisterMultiBackendAdapterFactory(name string, factory MultiBackendAdapterFactory) error {
	if name == "" {
		return fmt.Errorf("Empty name passed to RegisterMRAFactory")
	}

	registeredMBAdaptorFactories[name] = factory
	return nil
}

func LookupMultiBackendAdapterFactory(name string) (MultiBackendAdapterFactory, error) {
	factory, ok := registeredMBAdaptorFactories[name]
	if !ok {
		return nil, fmt.Errorf("Factory %s not registered", name)
	}

	return factory, nil
}
