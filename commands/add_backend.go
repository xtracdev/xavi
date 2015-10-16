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

//AddBackend provides a CLI compatible command
type AddBackend struct {
	UI      cli.Ui
	KVStore kvstore.KVStore
}

//Help provides detailed command help
func (ab *AddBackend) Help() string {
	helpText := `
	Usage: xavi add-backend [options]

		Options:
			-name Backend name
			-servers List of servers to add to backend, e.g. server1,server2,server3 no spaces
			-load-balancer-policy Load balancer policy name

	Known load balancers:`

	helpText = fmt.Sprintf("%s\n\t\t%s", helpText, loadbalancer.RegisteredLoadBalancers())

	return strings.TrimSpace(helpText)
}

//TODO - need to use flags Var to populate a slice of server names
//see https://golang.org/src/flag/example_test.go

//Run processed the command line argument passed in args for adding
//a backend configuration to the KV store assocaited with AddBackend
func (ab *AddBackend) Run(args []string) int {
	var name, serverList, loadBalancerPolicy string
	cmdFlags := flag.NewFlagSet("add-backend", flag.ContinueOnError)
	cmdFlags.Usage = func() { ab.UI.Output(ab.Help()) }
	cmdFlags.StringVar(&name, "name", "", "")
	cmdFlags.StringVar(&serverList, "servers", "", "")
	cmdFlags.StringVar(&loadBalancerPolicy, "load-balancer-policy", "", "")
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

	//TODO - validate server names
	backend := &config.BackendConfig{
		Name:               name,
		ServerNames:        strings.Split(serverList, ","),
		LoadBalancerPolicy: loadBalancerPolicy,
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
