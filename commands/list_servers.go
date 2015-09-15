package commands

import (
	"bytes"
	"encoding/json"
	"github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/mitchellh/cli"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"strings"
)

//ServerList command
type ServerList struct {
	UI      cli.Ui
	KVStore kvstore.KVStore
}

//Help provides details on the ServerList command
func (sl *ServerList) Help() string {
	helpText := `
		Usage: xavi list-servers
		`

	return strings.TrimSpace(helpText)
}

//Run executes the ServerList command with the supplied args
func (sl *ServerList) Run(args []string) int {
	servers, err := config.ListServerConfigs(sl.KVStore)
	if err != nil {
		sl.UI.Error(err.Error())
		return 1
	}

	jsonRep, err := json.Marshal(servers)
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

//Synopsis provides a concise description of the ServerList command
func (sl *ServerList) Synopsis() string {
	return "List server definitions"
}
