package statsd

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	"github.com/armon/go-metrics"
	"github.com/xtracdev/xavi/env"
	"os"
	"strings"
	"time"
)


func init() {
	initializeFromEnvironmentSettings()
}

func initializeFromEnvironmentSettings() {
	envSettings := os.Getenv(env.StatsdEndpoint)
	if envSettings != "" {
		log.Info("Using statsd client to send telemetry to ", envSettings)
		sink, err := metrics.NewStatsdSink(envSettings)
		if err != nil{
			log.Warn("Unable to configure statds sink", err.Error())
			return
		}

		metrics.NewGlobal(metrics.DefaultConfig("xavi"), sink)
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
