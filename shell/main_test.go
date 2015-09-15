package shell

import (
	"bytes"
	"github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/kvstore"
	"strings"
	"testing"
)

func TestMainNoArgs(t *testing.T) {
	kvs, _ := kvstore.NewHashKVStore("")
	var args []string

	buf := new(bytes.Buffer)

	status := DoMain(args, kvs, buf)
	assert.Equal(t, 1, status)
	out := buf.String()
	assert.True(t, strings.Contains(out, "add-server"), "Missing add-server command.")
	assert.True(t, strings.Contains(out, "ping-server"), "Missing ping-server command.")
	assert.True(t, strings.Contains(out, "add-backend"), "Missing add-backend command.")
	assert.True(t, strings.Contains(out, "add-route"), "Missing add-route command.")
	assert.True(t, strings.Contains(out, "add-listener"), "Missing add-listener command.")
	assert.True(t, strings.Contains(out, "listen"), "Missing listen command.")
	assert.True(t, strings.Contains(out, "boot-rest-agent"), "Missing boot-rest-agent command.")
	assert.True(t, strings.Contains(out, "list-servers"), "Missing list-servers command.")
	assert.True(t, strings.Contains(out, "list-backends"), "Missing list-backends command.")
	assert.True(t, strings.Contains(out, "list-routes"), "Missing list-routes command.")
	assert.True(t, strings.Contains(out, "list-listeners"), "Missing list-listeners command.")
	assert.True(t, strings.Contains(out, "list-plugins"), "Missing list-plugins command.")
}

func TestSetupError(t *testing.T) {
	var args []string
	status := DoMain(args, nil, nil)
	assert.Equal(t, 1, status)
}
