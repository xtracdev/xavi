package config

var activeConfig map[string]*ServiceConfig

func init() {
	activeConfig = make(map[string]*ServiceConfig)
}

func RecordActiveConfig(serviceConfig *ServiceConfig) {
	if serviceConfig == nil || serviceConfig.Listener == nil {
		return
	}

	activeConfig[serviceConfig.Listener.Name] = serviceConfig
}

func ActiveConfigForListener(listenerName string) *ServiceConfig {
	return activeConfig[listenerName]
}

func ActiveListenerNames() []string {
	keys := make([]string, len(activeConfig))

	i := 0
	for k := range activeConfig {
		keys[i] = k
		i++
	}

	return keys
}
