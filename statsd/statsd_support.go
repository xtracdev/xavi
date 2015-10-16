package statsd

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	statsdlib "github.com/quipo/statsd"
	"github.com/xtracdev/xavi/env"
	"os"
	"strings"
	"time"
)

//Statsd provides the statsd interface
var Statsd statsdlib.Statsd

func init() {
	initializeFromEnvironmentSettings()
}

func initializeFromEnvironmentSettings() {
	env_setting := os.Getenv(env.StatsdEndpoint)
	if env_setting != "" {
		log.Info("Using buffered statsd client")
		client := statsdlib.NewStatsdClient(env_setting, "xavi.")
		client.CreateSocket()
		interval := time.Second * 10 // aggregate stats and flush every 10 seconds
		Statsd = statsdlib.NewStatsdBuffer(interval, client)
		//defer statsd.Close() TODO golang exit hook
	} else {
		log.Info("Using noop statsd interface")
		var noopClient statsdlib.NoopClient
		Statsd = noopClient
	}
}

//formatServiceName removes url path separators from the service name. Not doing
//this seems to mess up the graphite/whisper storage path and defeats
//obtaining any metrics. Graphite/carbon/whisper writes a dash instead
//of the slash, so /hello becomes -hello, etc.
func FormatServiceName(name string) string {
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
