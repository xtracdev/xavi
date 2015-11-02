package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/env"
	"github.com/xtracdev/xavi/logging"
	"net"
	"os"
	"strings"
)

func addXapHook() error {
	loggingOpts := os.Getenv(env.LoggingOpts)
	if !strings.Contains(loggingOpts, env.Tcplog) {
		return nil
	}

	loggingAddress := os.Getenv(env.TcplogAddress)
	if loggingAddress == "" {
		log.Info("TCPLOG specified, but not address available using TCPLOG_ADDRESS - defaulting to localhost:5000")
		loggingAddress = "localhost:5000"
	}

	host, port, err := net.SplitHostPort(loggingAddress)
	if err != nil {
		log.Error("Unable to parse loggng address", loggingAddress, err)
		return err
	}

	hook, err := logging.NewTCPLoggingHook(host, port)
	if err != nil {
		log.Error("Unable to load the xaphook hook", err)
		return err
	}

	log.AddHook(hook)
	return nil
}

func init() {
	addXapHook()
}
