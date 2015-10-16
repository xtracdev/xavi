package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/plugin"
	"github.com/xtracdev/xavi/plugin/logging"
	"github.com/xtracdev/xavi/runner"
	"os"
)

func registerPlugins() {
	log.Info("Register plugins")
	log.Info("Registering logging wrapper plugin factory")

	err := plugin.RegisterWrapperFactory("Logging", logging.NewLoggingWrapper)
	if err != nil {
		log.Warn("Error registering logging wrapper plugin factory")
	}
	log.Info("logging wrapper plugin registered")

	log.Info("Plugins registered")
}

func grabCommandLineArgs() []string {
	log.Info("starting with log level ", log.GetLevel())
	log.Info("parse command line arguments")
	args := os.Args[1:]
	logMsg := fmt.Sprintf("Invoking runner.Run with command line args: %v", args)
	log.Info(logMsg)
	log.Info("finished parsing command line arguments")
	return args
}

func main() {
	args := grabCommandLineArgs()
	runner.Run(args, registerPlugins)
}
