package statsd

import (
	"github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/env"
	"net"
	"os"
	"sync"
	"testing"
	"time"
)

func TestFormatSingleComponentPath(t *testing.T) {
	formatted := FormatServiceName("/hello")
	assert.Equal(t, "hello", formatted)
}

func TestFormatMultiComponentPath(t *testing.T) {
	formatted := FormatServiceName("/hello/world")
	assert.Equal(t, "hello.world", formatted)
}

func TestEnvConnect(t *testing.T) {

	//If we spin up the statsd library from the address we read from the envrionment, we'll see a UDP
	//connection established.
	var connected bool
	var wg sync.WaitGroup

	//Get a UDP address
	addr, err := net.ResolveUDPAddr("udp", "localhost:0")
	assert.Nil(t, err)
	t.Log("address: ", addr)

	//Wait for it...
	wg.Add(1)

	//Spin up something listening on the address
	go func() {

		//Don't wait more than 5 seconds
		go func() {
			time.Sleep(5 * time.Second)
			wg.Done()
		}()

		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			t.Fatal(err)
		}

		connected = true

		conn.Close()

		wg.Done()
	}()

	//Set the statsd address via an environment variable
	os.Setenv(env.StatsdEndpoint, addr.String())

	//Fire up the statsd library
	initializeFromEnvironmentSettings()

	//Validate the connection happened
	wg.Wait()
	assert.True(t, connected)
}
