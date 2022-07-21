package connection

import (
	"net"
	"os"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/matryer/is"
)

func TestTcp(t *testing.T) {
	is := is.New(t)

	connChan := make(chan net.Conn)
	listeningChann := make(chan bool)
	go func() { NewConnectionListener(ConnTCP, ":1112", connChan, listeningChann) }()

	listening, ok := <-listeningChann
	is.True(listening)
	is.True(ok)

	// Channel should be closed.
	listening, ok = <-listeningChann
	is.Equal(listening, false)
	is.Equal(ok, false)

	conn, err := net.Dial("tcp", ":1112")
	is.NoErr(err)
	defer conn.Close()

	_, ok = <-connChan
	is.True(ok)

	// Should not be accepting connections anymore.
	conn, err = net.Dial("tcp", ":1112")
	if err == nil {
		defer conn.Close()
		t.Error("Expected 2nd dial to error")
	}

	_, ok = <-connChan
	is.Equal(ok, false)
}

func TestWs(t *testing.T) {
	is := is.New(t)

	// Need to use os.hostname for github actions to pass.
	hostname, err := os.Hostname()
	is.NoErr(err)

	connChan := make(chan net.Conn)
	listeningChann := make(chan bool)
	go func() { NewConnectionListener(ConnWs, hostname+":1113", connChan, listeningChann) }()

	listening, ok := <-listeningChann
	is.True(listening)
	is.True(ok)

	// Channel should be closed.
	listening, ok = <-listeningChann
	is.Equal(listening, false)
	is.Equal(ok, false)

	uri := "ws://" + hostname + ":1113"

	conn, _, err := websocket.DefaultDialer.Dial(uri, nil)
	is.NoErr(err)
	defer conn.Close()

	_, ok = <-connChan
	is.True(ok)

	conn, _, err = websocket.DefaultDialer.Dial(uri, nil)
	if err == nil {
		defer conn.Close()
		t.Error("Expected 2nd dial to error")
	}

	_, ok = <-connChan
	is.Equal(ok, false)
}
