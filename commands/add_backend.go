package commands

import (
	"errors"
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"github.com/xtracdev/xavi/loadbalancer"
	"os"
	"strings"
)

//AddBackend provides a CLI compatible command
type AddBackend struct {
	UI      cli.Ui
	KVStore kvstore.KVStore
}

//ErrBadPathSPec indicates the given CACertPath is invalid. Note at config time we just check
//that the path is valid. When the listen command is run if the backend is part of a processing path
//we load the file, check its format and content, and so on.
var ErrBadPathSpec = errors.New("CA certificate path is invalid or inaccessible")

//Help provides detailed command help
func (ab *AddBackend) Help() string {
	helpText := `
	Usage: xavi add-backend [options]

		Options:
			-name Backend name
			-servers List of servers to add to backend, e.g. server1,server2,server3 no spaces
			-load-balancer-policy Load balancer policy name
			-cacert-path Path to PEM file containing CA cert for backend servers
			-tls-only Use TSL/HTTPS only when calling server.

	Known load balancers:`

	helpText = fmt.Sprintf("%s\n\t\t%s", helpText, loadbalancer.RegisteredLoadBalancers())

	return strings.TrimSpace(helpText)
}

//Run processed the command line argument passed in args for adding
//a backend configuration to the KV store assocaited with AddBackend
func (ab *AddBackend) Run(args []string) int {
	log.Debug("AddBackend run commands ", args)
	var name, serverList, loadBalancerPolicy, caCertPath string
	var tlsOnly bool
	cmdFlags := flag.NewFlagSet("add-backend", flag.ContinueOnError)
	cmdFlags.Usage = func() { ab.UI.Output(ab.Help()) }
	cmdFlags.StringVar(&name, "name", "", "")
	cmdFlags.StringVar(&serverList, "servers", "", "")
	cmdFlags.StringVar(&loadBalancerPolicy, "load-balancer-policy", "", "")
	cmdFlags.StringVar(&caCertPath, "cacert-path", "", "")
	cmdFlags.BoolVar(&tlsOnly, "tls-only", false, "")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	argErr := false

	if name == "" {
		ab.UI.Error("Name must be specified")
		argErr = true
	}

	if serverList == "" {
		ab.UI.Error("Server list must be specified")
		argErr = true
	}

	if argErr {
		ab.UI.Error("")
		ab.UI.Error(ab.Help())
		return 1
	}

	//Check load balancer policy
	if loadBalancerPolicy != "" && !loadbalancer.IsKnownLoadBalancerPolicy(loadBalancerPolicy) {
		ab.UI.Error(fmt.Sprintf("Unknown load balancer policy %s", loadBalancerPolicy))
		ab.UI.Error(fmt.Sprintf("  Known policies: %s", loadbalancer.RegisteredLoadBalancers()))
		return 1
	}

	//Check cacert
	if err := ab.validCertPath(caCertPath); err != nil {
		ab.UI.Error(err.Error())
		return 1
	}

	backend := &config.BackendConfig{
		Name:               name,
		ServerNames:        strings.Split(serverList, ","),
		LoadBalancerPolicy: loadBalancerPolicy,
		TLSOnly:            tlsOnly,
		CACertPath:         caCertPath,
	}

	if err := backend.Store(ab.KVStore); err != nil {
		ab.UI.Error(err.Error())
		return 1
	}

	if err := ab.KVStore.Flush(); err != nil {
		ab.UI.Error(err.Error())
		return 1
	}

	return 0

}

//Synopsis gives the synopsis of the AddBackend command
func (ab *AddBackend) Synopsis() string {
	return "Define a backend as a collection of servers"
}

func (ab *AddBackend) validCertPath(caCertPath string) error {
	log.Debugf("validCertPath %s", caCertPath)
	if caCertPath != "" {
		if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
			return ErrBadPathSpec
		}
	}
	return nil
}
