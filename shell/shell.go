package shell

import (
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/xtracdev/xavi/commands"
	"github.com/xtracdev/xavi/kvstore"
	"io"
	"os"
)

var xaviCommands map[string]cli.CommandFactory

func commandSetup(kvs kvstore.KVStore, writer io.Writer) error {
	if kvs == nil || writer == nil {
		return fmt.Errorf("Must supply non-nil KVS and Writer arguments")
	}

	ui := &cli.BasicUi{Writer: writer, ErrorWriter: writer}

	xaviCommands = map[string]cli.CommandFactory{
		"add-server": func() (cli.Command, error) {
			return &commands.AddServer{ui, kvs}, nil
		},
		"ping-server": func() (cli.Command, error) {
			return &commands.PingServer{ui, kvs}, nil
		},
		"add-backend": func() (cli.Command, error) {
			return &commands.AddBackend{ui, kvs}, nil
		},
		"add-route": func() (cli.Command, error) {
			return &commands.AddRoute{ui, kvs}, nil
		},
		"add-listener": func() (cli.Command, error) {
			return &commands.AddListener{ui, kvs}, nil
		},
		"listen": func() (cli.Command, error) {
			return &commands.Listen{ui, kvs}, nil
		},
		"boot-rest-agent": func() (cli.Command, error) {
			return &commands.RESTAgent{ui, kvs}, nil
		},
		"list-servers": func() (cli.Command, error) {
			return &commands.ServerList{ui, kvs}, nil
		},
		"list-backends": func() (cli.Command, error) {
			return &commands.BackendList{ui, kvs}, nil
		},
		"list-routes": func() (cli.Command, error) {
			return &commands.RouteList{ui, kvs}, nil
		},
		"list-listeners": func() (cli.Command, error) {
			return &commands.ListenerList{ui, kvs}, nil
		},
		"list-plugins": func() (cli.Command, error) {
			return &commands.PluginList{ui, kvs}, nil
		},
	}

	return nil
}

//DoMain runs the Xavi command line. This is packed as a reusable piece of code
//to allow directives to be registered via import into a main program
func DoMain(args []string, kvs kvstore.KVStore, writer io.Writer) int {

	if err := commandSetup(kvs, writer); err != nil {
		fmt.Println("Error setting up command line interface: ", err.Error())
		return 1
	}

	cli := &cli.CLI{
		Args:       args,
		Commands:   xaviCommands,
		HelpFunc:   cli.BasicHelpFunc("xavi"),
		HelpWriter: writer,
	}

	exitCode, err := cli.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		return 1
	}

	return exitCode
}
