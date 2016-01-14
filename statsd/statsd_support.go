package statsd

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	"github.com/armon/go-metrics"
	"github.com/xtracdev/xavi/env"
	"os"
	"strings"
	"time"
	"github.com/armon/go-metrics/datadog"
)


func init() {
	initializeFromEnvironmentSettings()
}


func configureStatsD(endpoint string) {
	namespace := os.Getenv(env.StatsdNamespace)
	if namespace == "" {
		namespace = "xavi"
	}

	useDataDog := os.Getenv(env.UseDataDogStatsD)
	if useDataDog == "" {
		log.Info("Using vanilla statsd")
		configureVanillaStatsD(endpoint, namespace)
	} else {
		log.Info("Using datadog statsd implementation")
		configureDatadogStatsd(endpoint, namespace)
	}
}

func configureDatadogStatsd(endpoint string, namespace string) {
	log.Info("Using datadog statsd client to send telemetry to ", endpoint, " using namespace ", namespace)
	ddhost := os.Getenv(env.DatadogHost)
	if ddhost == "" {
		ddhost = "xavi-host-with-the-most"
	}
	sink, err := datadog.NewDogStatsdSink(endpoint, ddhost)
	if err != nil {
		log.Warn("Unable to configure statds sink", err.Error())
		return
	}
	metrics.NewGlobal(metrics.DefaultConfig(namespace), sink)
}

func configureVanillaStatsD(envEndpoint string, namespace string) {
	log.Info("Using vanilla statsd client to send telemetry to ", envEndpoint)
	sink, err := metrics.NewStatsdSink(envEndpoint)
	if err != nil{
		log.Warn("Unable to configure statds sink", err.Error())
		return
	}
	metrics.NewGlobal(metrics.DefaultConfig(namespace), sink)
}

func initializeFromEnvironmentSettings() {
	log.Info("Configuring go metrics sink")
	envSettings := os.Getenv(env.StatsdEndpoint)
	if envSettings != "" {
		log.Info("configure statsd sink")
		configureStatsD(envSettings)
	} else {
		log.Info("Using in memory metrics accumulator - dump via USR1 signal")
		inm := metrics.NewInmemSink(10*time.Second, 5 * time.Minute)
		metrics.DefaultInmemSignal(inm)
		metrics.NewGlobal(metrics.DefaultConfig("xavi"), inm)
	}
}

//FormatServiceName removes url path separators from the service name. Not doing
//this seems to mess up the graphite/whisper storage path and defeats
//obtaining any metrics. Graphite/carbon/whisper writes a dash instead
//of the slash, so /hello becomes -hello, etc.
func FormatServiceName(name string) string {
	log.Info("formatting ", name)
	parts := strings.Split(name, "/")
	var buffer bytes.Buffer
	var firstPart = true
	for _, s := range parts {
		if s != "" {
			if !firstPart {
				buffer.WriteString(".")
			} else {
				firstPart = false
			}
			buffer.WriteString(s)
		}

	}
	return buffer.String()
}
