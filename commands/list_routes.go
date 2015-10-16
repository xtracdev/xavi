package commands

import (
	"bytes"
	"encoding/json"
	"github.com/mitchellh/cli"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"strings"
)

//RouteList command
type RouteList struct {
	UI      cli.Ui
	KVStore kvstore.KVStore
}

//Help provides details on the RouteList command
func (sl *RouteList) Help() string {
	helpText := `
		Usage: xavi list-routes
		`

	return strings.TrimSpace(helpText)
}

//Run executes the RouteList command with the supplied args
func (sl *RouteList) Run(args []string) int {
	routes, err := config.ListRouteConfigs(sl.KVStore)
	if err != nil {
		sl.UI.Error(err.Error())
		return 1
	}

	jsonRep, err := json.Marshal(routes)
	if err != nil {
		sl.UI.Error(err.Error())
		return 1
	}

	if bytes.Equal(jsonRep, config.NullJSON) {
		sl.UI.Output("[]")
	} else {
		sl.UI.Output(string(jsonRep))
	}

	return 0
}

//Synopsis provides a concise description of the RouteList command
func (sl *RouteList) Synopsis() string {
	return "List route definitions"
}
