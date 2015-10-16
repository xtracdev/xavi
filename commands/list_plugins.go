package commands

import (
	"strings"

	"github.com/mitchellh/cli"
	"github.com/xtracdev/xavi/kvstore"
	"github.com/xtracdev/xavi/plugin"
)

//PluginList command
type PluginList struct {
	UI      cli.Ui
	KVStore kvstore.KVStore
}

//Help provides details on the PluginList command
func (sl *PluginList) Help() string {
	helpText := `
		Usage: xavi list-plugins
		`

	return strings.TrimSpace(helpText)
}

//Run executes the PluginList command with the supplied args
func (sl *PluginList) Run(args []string) int {

	plugins := plugin.ListPlugins()
	if len(plugins) > 0 {
		for _, p := range plugins {
			sl.UI.Output("Plugin name: '" + p + "'")
		}
	} else {
		sl.UI.Output("No plugins registered")
	}

	return 0
}

//Synopsis provides a concise description of the PluginList command
func (sl *PluginList) Synopsis() string {
	return "List backend definitions"
}
