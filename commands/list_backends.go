package commands

import (
	"bytes"
	"encoding/json"
	"github.com/mitchellh/cli"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"strings"
)

//BackendList command
type BackendList struct {
	UI      cli.Ui
	KVStore kvstore.KVStore
}

//Help provides details on the BackendList command
func (sl *BackendList) Help() string {
	helpText := `
		Usage: xavi list-backends
		`

	return strings.TrimSpace(helpText)
}

//Run executes the BackendList command with the supplied args
func (sl *BackendList) Run(args []string) int {
	backends, err := config.ListBackendConfigs(sl.KVStore)
	if err != nil {
		sl.UI.Error(err.Error())
		return 1
	}

	jsonRep, err := json.Marshal(backends)
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

//Synopsis provides a concise description of the BackendList command
func (sl *BackendList) Synopsis() string {
	return "List backend definitions"
}
