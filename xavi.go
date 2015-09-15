/*
Ok - here's the deal on Mac OS X for getting log messages into /var/log/system

First, by default it doesn't look like you can write via udp to localhost:514 - see
http://stackoverflow.com/questions/5510563/how-to-start-syslogd-server-on-mac-to-accept-remote-logging-messages

Without enabling the above, specify "" as the network and raddr args to NewSyslogHook. You can specify the
sender via the last arg (tag)

To catch info messages in the log, you need to edit the /etc/asl.conf file:

 Change the ? [<= Level notice] store line to ? [<= Level info] store, and likewise
 change the ? [<= Level notice] file system.log line to ? [<= Level info] file system.log

 After changing the config file (you did make a backup first yeah), restart syslogd via

 sudo launchctl unload /System/Library/LaunchDaemons/com.apple.syslogd.plist
 sudo launchctl load /System/Library/LaunchDaemons/com.apple.syslogd.plist

 After that you'll get your info message for shizzle.

 Note: syslog references removed as they do not build on windows. The above is left for historical context.
*/

package main

import (
	log "github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/Sirupsen/logrus"
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
