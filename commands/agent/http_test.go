package agent

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net"
	"net/http"
	"strconv"
	"testing"
)

func handleFoo(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(200)
	rw.Write([]byte("foo"))
}

func handleBar(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(200)
	rw.Write([]byte("bar"))
}

func findFreePort() int {
	l, _ := net.Listen("tcp", "localhost:0")
	defer l.Close()
	_, port, _ := net.SplitHostPort(l.Addr().String())
	portInt, _ := strconv.Atoi(port)
	return portInt
}

func TestUriHandling(t *testing.T) {
	freePort := findFreePort()
	baseAddr := fmt.Sprintf("localhost:%d", freePort)
	a := NewAgent(baseAddr, nil)
	a.addHandler("/foo", handleFoo)
	a.addHandler("/bar", handleBar)
	go func(a *Agent) { a.Start() }(a)

	res, err := http.Get("http://" + baseAddr + "/foo")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.StatusCode, 200)
	res.Body.Close()

	res, err = http.Get("http://" + baseAddr + "/bar")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.StatusCode, 200)
	res.Body.Close()

	res, err = http.Get("http://" + baseAddr + "/xxxx")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.StatusCode, 404)
	res.Body.Close()

}
