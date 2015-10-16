package commands

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func testMakePingServer(faultKVStore bool, writeServerDefs bool, port int) (*bytes.Buffer, *PingServer) {

	var kvs, _ = kvstore.NewHashKVStore("")

	var writer = new(bytes.Buffer)
	var ui = &cli.BasicUi{Writer: writer, ErrorWriter: writer}
	var pingServer = &PingServer{
		UI:      ui,
		KVStore: kvs,
	}

	if writeServerDefs {
		if port == -1 {
			port = 124
			s := &config.ServerConfig{"unpingable", "host", port, "", "none", 0, 0}
			s.Store(kvs)
		} else {
			s := &config.ServerConfig{"pingable", "0.0.0.0", port, "/pingme", "none", 0, 0}
			s.Store(kvs)
		}
	}

	if faultKVStore {
		kvs.InjectFaults()
	}
	return writer, pingServer
}

func TestPingServerHelp(t *testing.T) {
	_, ps := testMakePingServer(false, false, -1)
	assert.NotEmpty(t, ps.Help())
}

func TestPingServerSynopsis(t *testing.T) {
	_, ps := testMakePingServer(false, false, -1)
	assert.NotEmpty(t, ps.Synopsis())
}

func TestPingServerWithNoArgs(t *testing.T) {
	_, ps := testMakePingServer(false, false, -1)
	status := ps.Run([]string{})
	assert.Equal(t, 1, status)
}

func TestPingServerWithNoSuchServer(t *testing.T) {
	_, ps := testMakePingServer(false, false, -1)
	status := ps.Run([]string{"foo"})
	assert.Equal(t, 1, status)
}

func TestPingServerWithFaultyKVStore(t *testing.T) {
	_, ps := testMakePingServer(true, false, -1)
	status := ps.Run([]string{"foo"})
	assert.Equal(t, 1, status)
}

func TestPingServerWithNoPingUri(t *testing.T) {
	_, ps := testMakePingServer(false, true, -1)
	status := ps.Run([]string{"unpingable"})
	assert.Equal(t, 1, status)
}

func TestPingServerWithPingable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	urlParts := strings.Split(ts.URL, ":")
	port, _ := strconv.Atoi(urlParts[2])

	_, ps := testMakePingServer(false, true, port)
	status := ps.Run([]string{"pingable"})
	assert.Equal(t, 0, status)

}
