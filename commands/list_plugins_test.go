package commands

import (
	"bytes"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/kvstore"
	"github.com/xtracdev/xavi/plugin"
	"golang.org/x/net/context"
	"net/http"
	"testing"
)

//NewAWrapper instantiates AWrapper
func NewAWrapper(args ...interface{}) plugin.Wrapper {
	return new(AWrapper)
}

//AWrapper can wrap http handlers
type AWrapper struct{}

//Wrap wraps http.Handlers with A stuff
func (aw AWrapper) Wrap(h plugin.ContextHandler) plugin.ContextHandler {
	return plugin.ContextHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		h.ServeHTTPContext(ctx, w, r)
		w.Write([]byte("A wrapper wrote this\n"))
	})
}

func testMakeListPlugins(faultyStore bool, withPlugin bool) (*bytes.Buffer, *PluginList) {

	var kvs, _ = kvstore.NewHashKVStore("")
	if faultyStore {
		kvs.InjectFaults()
	}
	var writer = new(bytes.Buffer)
	var ui = &cli.BasicUi{Writer: writer, ErrorWriter: writer}
	var listPlugins = &PluginList{
		UI:      ui,
		KVStore: kvs,
	}

	if withPlugin {
		plugin.RegisterWrapperFactory("AWrapper", NewAWrapper)
	}

	return writer, listPlugins
}

func TestListPluginsNoneRegistered(t *testing.T) {
	writer, listPlugins := testMakeListPlugins(false, false)
	var args []string
	status := listPlugins.Run(args)
	assert.Equal(t, 0, status)
	out := string(writer.Bytes())
	assert.Contains(t, out, "No plugins registered")
}

func TestListPluginsOneRegistered(t *testing.T) {
	writer, listPlugins := testMakeListPlugins(false, true)
	var args []string
	status := listPlugins.Run(args)
	assert.Equal(t, 0, status)
	out := string(writer.Bytes())
	assert.Contains(t, out, "AWrapper")
}

func TestListPluginHelp(t *testing.T) {
	_, listPlugins := testMakeListPlugins(false, false)
	assert.NotEmpty(t, listPlugins.Help())
}

func TestListPluginSynopsis(t *testing.T) {
	_, listPlugins := testMakeListPlugins(false, false)
	assert.NotEmpty(t, listPlugins.Synopsis())
}
