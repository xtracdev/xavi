package commands

import (
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"github.com/xtracdev/xavi/loadbalancer"
	"strings"
)

//AddServer command
type AddServer struct {
	UI      cli.Ui
	KVStore kvstore.KVStore
}

//Help provides details on using the AddServer command
func (as *AddServer) Help() string {
	helpText := `
	Usage: xavi add-server [options]

		Adds a server to the system, which can then be associated with one or more backends.

	Options:
		-address Server address (e.g. 127.0.0.1, server1, etc.)strings
		-port Server port
		-name Name of server for reference in backend configuration.
		-ping-uri Uri resource used to assess health via an HTTP GET (e.g. /hello)
		-health-check Health check type (optional)
		-health-check-interval (optional) duration in milliseconds at which health is checked
		-health-check-timeout (optional) time in killiseconds for healthcheck timeout

	Known health checks:
	`

	helpText = fmt.Sprintf("%s\n\t\t%s", helpText, loadbalancer.KnownHealthChecks())

	return strings.TrimSpace(helpText)
}

//Run executes the AddServer command using the supplied args
func (as *AddServer) Run(args []string) int {
	var address, name, pinguri string
	var port int
	var healthCheck string
	var healthCheckInterval, healthCheckTimeout int

	cmdFlags := flag.NewFlagSet("add-server", flag.ContinueOnError)
	cmdFlags.Usage = func() { as.UI.Output(as.Help()) }
	cmdFlags.StringVar(&address, "address", "", "")
	cmdFlags.IntVar(&port, "port", -1, "")
	cmdFlags.StringVar(&name, "name", "", "")
	cmdFlags.StringVar(&pinguri, "ping-uri", "", "")
	cmdFlags.StringVar(&healthCheck, "health-check", "none", "")
	cmdFlags.IntVar(&healthCheckInterval, "health-check-interval", loadbalancer.DefaultHealthCheckInterval, "")
	cmdFlags.IntVar(&healthCheckTimeout, "health-check-timeout", loadbalancer.DefaultHealthCheckTimeout, "")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	argErr := false

	//Check address
	if address == "" {
		as.UI.Error("Address name must be specified")
		argErr = true
	}

	//Check  port
	if port == -1 {
		as.UI.Error("Port must be specified")
		argErr = true
	}

	//Check name
	if name == "" {
		as.UI.Error("Name must be specified")
		argErr = true
	}

	//Check health check
	if !loadbalancer.IsKnownHealthCheck(healthCheck) {
		as.UI.Error("Invalid health check specified")
		argErr = true
	}

	if healthCheckTimeout >= healthCheckInterval {
		as.UI.Error("Health check timeout must be less than health check interval")
		argErr = true
	}

	if argErr {
		as.UI.Error("")
		as.UI.Error(as.Help())
		return 1
	}

	serverDef := &config.ServerConfig{
		Name:                name,
		Port:                port,
		Address:             address,
		PingURI:             pinguri,
		HealthCheck:         healthCheck,
		HealthCheckInterval: healthCheckInterval,
		HealthCheckTimeout:  healthCheckTimeout,
	}

	err := serverDef.Store(as.KVStore)
	if err != nil {
		as.UI.Error(err.Error())
		return 1
	}

	if err := as.KVStore.Flush(); err != nil {
		as.UI.Error(err.Error())
		return 1
	}

	return 0
}

//Synopsis gives a concise description of the AddServer command
func (as *AddServer) Synopsis() string {
	return "Add a server definition"
}
