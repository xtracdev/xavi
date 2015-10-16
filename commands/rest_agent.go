package commands

import (
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/xtracdev/xavi/commands/agent"
	"github.com/xtracdev/xavi/kvstore"
	"os"
	"os/signal"
	"strings"
)

//RESTAgent command, which starts a REST endpoint for executing commands supporting a
//REST interface
type RESTAgent struct {
	UI      cli.Ui
	KVStore kvstore.KVStore
}

//Help provides details on the RESTAgent command
func (a *RESTAgent) Help() string {
	helpText := `
		Usage: xavi agent [options]
		
		Options:
			-address host:port to listen on
			`

	return strings.TrimSpace(helpText)
}

//Run executes the RESTAgent command with the supplied args
func (a *RESTAgent) Run(args []string) int {
	var address string
	cmdFlags := flag.NewFlagSet("agent", flag.ContinueOnError)
	cmdFlags.Usage = func() { a.UI.Error(a.Help()) }
	cmdFlags.StringVar(&address, "address", "", "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	//Check address
	if address == "" {
		a.UI.Error("Address must be specified")
		a.UI.Error("")
		a.UI.Error(a.Help())
		return 1
	}

	exitChannel := make(chan int)
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)

	go func() {
		for _ = range signalChannel {
			exitChannel <- 0
		}
	}()

	go func(addr string, kvs kvstore.KVStore) {
		agent := agent.NewAgent(addr, kvs)
		agent.Start()
		exitChannel <- 1
	}(address, a.KVStore)

	exitStatus := <-exitChannel
	fmt.Printf("exiting with status %d\n", exitStatus)
	return exitStatus
}

//Synopsis provides a concise description of the RESTAgent command
func (a *RESTAgent) Synopsis() string {
	return "Boot REST API agent"
}
