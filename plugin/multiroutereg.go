package plugin

import "fmt"

var registeredMRAdaptorFactories map[string]MultiRouteAdapterFactory

func init() {
	registeredMRAdaptorFactories = make(map[string]MultiRouteAdapterFactory)
}

func ListMultirouteAdapters() []string {
	var adapters []string
	for key := range registeredMRAdaptorFactories {
		adapters = append(adapters, key)
	}
	return adapters
}

func MRARegistryContains(name string) bool {
	_, ok := registeredMRAdaptorFactories[name]
	return ok
}

func RegisterMRAFactory(name string, factory MultiRouteAdapterFactory) error {
	if name == "" {
		return fmt.Errorf("Empty name passed to RegisterMRAFactory")
	}

	registeredMRAdaptorFactories[name] = factory
	return nil
}

func LookupMRAFactory(name string) (MultiRouteAdapterFactory, error) {
	factory, ok := registeredMRAdaptorFactories[name]
	if !ok {
		return nil, fmt.Errorf("Factory %s not registered", name)
	}

	return factory, nil
}
