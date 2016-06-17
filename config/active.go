package config

var activeConfig map[string]*ServiceConfig

func init() {
	activeConfig = make(map[string]*ServiceConfig)
}

//RecordActiveConfig allows a service config constructed from
//a listener name to be recorded and made available to all
//code using the xavi framework
func RecordActiveConfig(serviceConfig *ServiceConfig) {
	if serviceConfig == nil || serviceConfig.Listener == nil {
		return
	}

	activeConfig[serviceConfig.Listener.Name] = serviceConfig
}

//ActiveConfigForListener returns the ServiceConfig associated with
//the provided listener names
func ActiveConfigForListener(listenerName string) *ServiceConfig {
	return activeConfig[listenerName]
}

//ActiveListenerNames reurns the names of the listeners for whom
//the associated ServiceConfig has been recorded.
func ActiveListenerNames() []string {
	keys := make([]string, len(activeConfig))

	i := 0
	for k := range activeConfig {
		keys[i] = k
		i++
	}

	return keys
}
