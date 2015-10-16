package commands

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"net/http"
	"strings"
)

//PingServer command, use to ping a server based on the named server definition.
type PingServer struct {
	UI      cli.Ui
	KVStore kvstore.KVStore
}

//Help provides details on the PingServer command
func (ps *PingServer) Help() string {
	helpText := `
	Usage: xavi ping-server server-name

		'Ping' a server by invoking an HTTP GET on the ping uri specified by the server.
	`

	return strings.TrimSpace(helpText)

}

//Run executes the PingServer command using the supplied args
func (ps *PingServer) Run(args []string) int {
	//Assume we get os.Args[1:] as the input via the CLI
	if len(args) != 1 {
		ps.UI.Error(ps.Help())
		return 1
	}

	log.Info("pinging ", args[0])

	//Read the definition from the key store
	key := "servers/" + args[0]
	log.Info("Read key " + key)
	bv, err := ps.KVStore.Get(key)
	if err != nil {
		ps.UI.Error("Error reading key: " + err.Error())
		return 1
	}

	if bv == nil {
		ps.UI.Error("No server definition exists named " + key)
		return 1
	}

	serverDef := config.JSONToServer(bv)
	if serverDef.PingURI == "" {
		ps.UI.Error("No ping uri specified for server")
		return 1
	}

	//Form the url
	url := fmt.Sprintf("http://%s:%d%s",
		serverDef.Address,
		serverDef.Port,
		serverDef.PingURI)

	//Ping the server
	_, err = http.Get(url)
	if err != nil {
		ps.UI.Error(err.Error())
		return 1
	}

	ps.UI.Info("server ping ok")
	return 0

}

//Synopsis provides a concise description of the PingServer command
func (ps *PingServer) Synopsis() string {
	return "ping a server"
}
