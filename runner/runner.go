package runner

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/env"
	"github.com/xtracdev/xavi/kvstore"
	"github.com/xtracdev/xavi/shell"
	"net/http"

	//Needed to pickup the package imports for the profiler
	"github.com/xtracdev/xavi/info"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strings"
	"github.com/xtracdev/xavi/config"
)

//Build version is set via the command line, e.g.
//go build -ldflags "-X github.com/xtracdev/xavi/runner.BuildVersion=20160129.1"
var BuildVersion string

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	setLoggingLevel()
}

func setLoggingLevel() {

	logLevel := strings.ToLower(os.Getenv(env.LoggingLevel))
	switch logLevel {
	default:
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
		//Note - makes no sense to set the default log levels to fatal or to panic
	}

	log.Info("log level set: ", log.GetLevel())
}

func getKVStoreEndpoint() string {
	endpoint := os.Getenv(env.KVStoreURL)
	if endpoint == "" {
		log.Fatal(fmt.Sprintf("Required environment variable %s for configuration KV store must be specified", env.KVStoreURL))
	}
	return endpoint
}

func setupXAVIEnvironment(pluginRegistrationFn func()) kvstore.KVStore {
	log.Info("GOMAXPROCS: ", runtime.GOMAXPROCS(-1))
	if pluginRegistrationFn != nil {
		log.Info("Registering plugins")
		pluginRegistrationFn()
	}

	log.Info("Obtaining handle to KV store")
	kvs, err := kvstore.NewKVStore(getKVStoreEndpoint())
	if err != nil {
		log.Fatal(err.Error())
	}

	return kvs
}

//fire up the profiler endpoint if indicated by the environment. Return true we attempt to fire it up,
//false otherwise.
func fireUpPProf() bool {
	pprofEndpoint := os.Getenv(env.PProfEndpoint)
	if pprofEndpoint == "" {
		log.Info("Profiling port not enabled - ", env.PProfEndpoint, " not specified")
		return false
	}

	log.Info("Attempting to start pprof listener on  ", pprofEndpoint)

	go func() {
		log.Println(http.ListenAndServe(pprofEndpoint, nil))
	}()

	return true

}

const versionNotSpecified = `%s%s: no build version specified. A version can be set on the
command line using the –X –ldflags option, for example
go build -ldflags "-X github.com/xtracdev/xavi/runner.BuildVersion=20160129.1"`

func dumpVersionAndExit(args []string) (string, bool) {
	var versionFormat string

	switch BuildVersion {
	case "":
		versionFormat = versionNotSpecified
	default:
		versionFormat = "%s: build version %s"
	}

	output := fmt.Sprintf(versionFormat, args[0], BuildVersion)
	return output, len(args) == 2 && args[1] == "-version"
}

//KVSCallbackFn defines a function type that can be registered with
//the runner package. On startup, Run will call all registered functions with
//the key value store used for the config
type KVSCallbackFn func(kvs kvstore.KVStore) error

var initKVSFuncs []KVSCallbackFn

func AddKVSCallbackFunction(f KVSCallbackFn) {
	if f != nil {
		initKVSFuncs = append(initKVSFuncs, f)
	}
}

//Run starts a process delegating to the shell.DoMain function
func Run(args []string, pluginRegistrationFn func()) {
	version, exit := dumpVersionAndExit(os.Args)
	if exit == true {
		fmt.Println(version)
		os.Exit(0)
	}

	info.BuildVersion = BuildVersion

	log.Info(version)
	fireUpPProf()
	kvs := setupXAVIEnvironment(pluginRegistrationFn)

	if len(os.Args) > 1 && os.Args[1] == "listen" {
		config.ListenContext = true
	}

	for _, f := range initKVSFuncs {
		err := f(kvs)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

	os.Exit(shell.DoMain(args, kvs, os.Stdout))
}
