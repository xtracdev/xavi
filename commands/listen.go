package commands

import (
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/xtracdev/xavi/kvstore"
	"github.com/xtracdev/xavi/service"
	"os"
	"os/signal"
	"runtime/pprof"
	"strings"
)

//Listen command, which starts up a proxy listener based on the configuration associated
//with the given listener name
type Listen struct {
	UI      cli.Ui
	KVStore kvstore.KVStore
}

//Help provides details on the Listen command
func (l *Listen) Help() string {
	helpText := `
		Usage: xavi listen [options]
		
		Options:
			-ln Listener name - name of listener definition to use
			-address host:port to listen on
			-cpuprofile Enable Go lang profiling and write to the file named in the argument
			`

	return strings.TrimSpace(helpText)
}

//Run executes the Listen command with the supplied arguments
func (l *Listen) Run(args []string) int {
	var listener, address, cpuprofile string
	cmdFlags := flag.NewFlagSet("listen", flag.ContinueOnError)
	cmdFlags.Usage = func() { l.UI.Error(l.Help()) }
	cmdFlags.StringVar(&listener, "ln", "", "")
	cmdFlags.StringVar(&address, "address", "", "")
	cmdFlags.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	argErr := false

	//Check listener
	if listener == "" {
		l.UI.Error("Listener name must be specified")
		argErr = true
	}

	//Check address
	if address == "" {
		l.UI.Error("Address must be specified")
		argErr = true
	}

	if argErr {
		l.UI.Error("")
		l.UI.Error(l.Help())
		return 1
	}

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			l.UI.Error(err.Error())
			return 1
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	s, err := service.BuildServiceForListener(listener, address, l.KVStore)
	if err != nil {
		l.UI.Error(err.Error())
		return 1
	}

	l.UI.Info(fmt.Sprintf("***Service:\n%s", s))

	exitChannel := make(chan int)
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)

	go func() {
		for _ = range signalChannel {
			exitChannel <- 0
		}
	}()

	go func(service service.Service) {
		service.Run()
		//Run can return if it can't open ports, etc.
		exitChannel <- 1
	}(s)

	exitStatus := <-exitChannel
	fmt.Printf("exiting with status %d\n", exitStatus)
	return exitStatus
}

//Synopsis provides a concise description of the Listen command
func (l *Listen) Synopsis() string {
	return "Listen on an address using a listener definition"
}
