package connection_test

import (
	"net"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/laytan/elephp/pkg/connection"
	"github.com/matryer/is"
)

func TestTcp(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	connChan := make(chan net.Conn)
	listeningChann := make(chan bool)
	go func() {
		connection.NewConnectionListener(connection.ConnTCP, ":1112", connChan, listeningChann)
	}()

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
	t.Parallel()
	// TODO: fix this test.
	t.Skip("This does not consistently succeed")

	is := is.New(t)

	connChan := make(chan net.Conn)
	listeningChann := make(chan bool)
	go func() {
		connection.NewConnectionListener(connection.ConnWs, ":1113", connChan, listeningChann)
	}()

	listening, ok := <-listeningChann
	is.True(listening)
	is.True(ok)

	// Channel should be closed.
	listening, ok = <-listeningChann
	is.Equal(listening, false)
	is.Equal(ok, false)

	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:1113", nil)
	is.NoErr(err)
	defer conn.Close()

	_, ok = <-connChan
	is.True(ok)

	conn, _, err = websocket.DefaultDialer.Dial("ws://127.0.0.1:1113", nil)
	if err == nil {
		defer conn.Close()
		t.Error("Expected 2nd dial to error")
	}

	_, ok = <-connChan
	is.Equal(ok, false)
}
