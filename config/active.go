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
