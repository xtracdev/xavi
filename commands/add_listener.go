package commands

import (
	"flag"
	"github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/mitchellh/cli"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"strings"
)

//AddListener command
type AddListener struct {
	UI      cli.Ui
	KVStore kvstore.KVStore
}

//Help provides details on the expected arguments for the AddListener command
func (al *AddListener) Help() string {
	helpText := `
	Usage: xavi add-listener [options]

	Options:
		-name Listener name
		-routes List of routes, comma separated, no spaces
	`

	return strings.TrimSpace(helpText)
}

//Run executes the AddListener command with the given arguments
func (al *AddListener) Run(args []string) int {
	var name, routes string
	cmdFlags := flag.NewFlagSet("add-listener", flag.ContinueOnError)
	cmdFlags.Usage = func() { al.UI.Output(al.Help()) }
	cmdFlags.StringVar(&name, "name", "", "")
	cmdFlags.StringVar(&routes, "routes", "", "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	argErr := false

	//Check name
	if name == "" {
		al.UI.Error("Name must be specified")
		argErr = true
	}

	//Check routes
	if routes == "" {
		al.UI.Error("Routes must be specified")
		argErr = true
	}

	if argErr {
		al.UI.Error("")
		al.UI.Error(al.Help())
		return 1
	}

	//TODO - validate routes
	listenerDef := &config.ListenerConfig{
		Name:       name,
		RouteNames: strings.Split(routes, ","),
	}

	if err := listenerDef.Store(al.KVStore); err != nil {
		al.UI.Error(err.Error())
		return 1
	}

	if err := al.KVStore.Flush(); err != nil {
		al.UI.Error(err.Error())
		return 1
	}

	return 0
}

//Synopsis gives the synopsis of the AddListener command
func (al *AddListener) Synopsis() string {
	return "Add a listener"
}
