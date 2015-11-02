package logging


import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"net"
	"os"
	"strconv"
)

const (
	format = "Jan 2 15:04:05"
)

//TCPLoggingHook holds the context needed for installing a hook in the Sirupsen/logrus logging package.
type TCPLoggingHook struct {
	Host    string
	Port    int
	UDPConn net.Conn
}

//NewTCPLoggingHook creates a new TCPLoggingHook
func NewTCPLoggingHook(host, port string) (*TCPLoggingHook, error) {
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, portInt))
	return &TCPLoggingHook{host, portInt, conn}, err
}

//Fire is called when a log event is fired.
func (hook *TCPLoggingHook) Fire(entry *logrus.Entry) error {
	msg, _ := entry.String()

	bytesWritten, err := hook.UDPConn.Write([]byte(msg))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to send log line to via TCPLoggingHook UDP. Wrote %d bytes before error: %v", bytesWritten, err)
		return err
	}

	return nil
}

// Levels returns the available logging levels.
func (hook *TCPLoggingHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}
