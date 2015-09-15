package logging

import (
	"bufio"
	"fmt"
	log "github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"net"
	"strings"
	"testing"
)

func newLocalListener(t *testing.T) net.Listener {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	return ln
}

func TestLogHook(t *testing.T) {
	listener := newLocalListener(t)
	fmt.Println(listener.Addr())

	host, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Error(err)
	}

	hook, err := NewTCPLoggingHook(host, port)
	if err != nil {
		t.Error(err)
	}

	go func() {

		conn, err := listener.Accept()
		if err != nil {
			t.Fatal(err)
		}
		reader := bufio.NewReader(conn)
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err.Error())
			t.Fatal(err)
		}
		fmt.Printf(">>%s<<\n", msg)
		if !strings.Contains(msg, "This is a test") {
			t.Error("Expected message to contain 'This is a test', got " + msg)
		}
		conn.Close()
		listener.Close()

	}()

	log.AddHook(hook)
	log.Info("This is a test")
}

func TestBreakOfLogHook(t *testing.T) {
	listener := newLocalListener(t)
	host, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Error(err)
	}

	hook, err := NewTCPLoggingHook(host, port)
	if err != nil {
		t.Error(err)
	}

	log.AddHook(hook)
	listener.Close()
	log.Info("This is a test")
}
