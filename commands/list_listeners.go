package commands

import (
	"bytes"
	"encoding/json"
	"github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/mitchellh/cli"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"strings"
)

//ListenerList command
type ListenerList struct {
	UI      cli.Ui
	KVStore kvstore.KVStore
}

//Help provides details on the ListenerList command
func (sl *ListenerList) Help() string {
	helpText := `
		Usage: xavi list-listeners
		`

	return strings.TrimSpace(helpText)
}

//Run executes the ListenerList command with the supplied args
func (sl *ListenerList) Run(args []string) int {
	listeners, err := config.ListListenerConfigs(sl.KVStore)
	if err != nil {
		sl.UI.Error(err.Error())
		return 1
	}

	jsonRep, err := json.Marshal(listeners)
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

//Synopsis provides a concise description of the ListenerList command
func (sl *ListenerList) Synopsis() string {
	return "List listener definitions"
}
