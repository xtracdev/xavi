package commands

import (
	"flag"
	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"github.com/xtracdev/xavi/plugin"
	"strings"
)

//AddRoute command
type AddRoute struct {
	UI      cli.Ui
	KVStore kvstore.KVStore
}

//Help provides details on command line options for AddRoute
func (ar *AddRoute) Help() string {
	helpText := `
	Usage: xavi [options]

	Options
		-name Route name
		-backends Backend name
		-base-uri Base uri to match
		-plugins Optional list of plugin names
		-msgprop Message properties for matching route
		`

	return strings.TrimSpace(helpText)
}

func (ar *AddRoute) validateBackend(name string) (bool, error) {
	key := "backends/" + name
	log.Info("Read key " + key)
	backend, err := ar.KVStore.Get(key)
	if err != nil {
		return false, err
	}

	return backend != nil, nil
}

func pluginssRegistered(plugins []string) (string, bool) {
	if len(plugins) > 0 && plugins[0] != "" {
		for _, f := range plugins {
			if !plugin.RegistryContains(f) {
				return f, false
			}
		}
	}

	return "", true
}

//Run executes the AddRoute command using the provided arguments
func (ar *AddRoute) Run(args []string) int {
	var name, backends, baseuri, pluginList, msgprop string
	cmdFlags := flag.NewFlagSet("add-route", flag.ContinueOnError)
	cmdFlags.Usage = func() { ar.UI.Output(ar.Help()) }
	cmdFlags.StringVar(&name, "name", "", "")
	cmdFlags.StringVar(&backends, "backends", "", "")
	cmdFlags.StringVar(&baseuri, "base-uri", "", "")
	cmdFlags.StringVar(&pluginList, "plugins", "", "")
	cmdFlags.StringVar(&msgprop, "msgprop", "", "")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	argErr := false

	if name == "" {
		ar.UI.Error("Name must be specified")
		argErr = true
	}

	if backends == "" {
		ar.UI.Error("Backends must be specified")
		argErr = true
	}

	if baseuri == "" {
		ar.UI.Error("Base uri must be specified")
		argErr = true
	}

	if argErr {
		ar.UI.Error("")
		ar.UI.Error(ar.Help())
		return 1
	}

	cmdBackends := strings.Split(backends, ",")
	for _, beName := range cmdBackends {
		validName, err := ar.validateBackend(beName)
		if err != nil || !validName {
			ar.UI.Error("backend not found: " + beName)
			return 1
		}
	}

	var plugins []string
	if pluginList != "" {
		plugins = strings.Split(pluginList, ",")
		unregistered, pluginsRegistered := pluginssRegistered(plugins)
		if !pluginsRegistered {
			ar.UI.Error("Error: plugin list contains unregistered plugin: '" + unregistered + "'")
			return 1
		}
	}

	route := &config.RouteConfig{
		Name:     name,
		Backends: cmdBackends,
		URIRoot:  baseuri,
		Plugins:  plugins,
		MsgProps: msgprop,
	}

	if err := route.Store(ar.KVStore); err != nil {
		ar.UI.Error(err.Error())
		return 1
	}

	if err := ar.KVStore.Flush(); err != nil {
		ar.UI.Error(err.Error())
		return 1
	}

	return 0
}

//Synopsis provides a concise description of the AddRoute command.
func (ar *AddRoute) Synopsis() string {
	return "Create a route linking a uri pattern to a backend"
}
