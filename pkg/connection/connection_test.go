package connection_test

import (
	"net"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/laytan/phpls/pkg/connection"
	"github.com/stretchr/testify/require"
)

func TestTcp(t *testing.T) {
	t.Parallel()
	// TODO: fix this test.
	t.Skip("This does not consistently succeed")

	connChan := make(chan net.Conn)
	listeningChann := make(chan bool)
	go func() {
		connection.NewConnectionListener(connection.ConnTCP, ":1112", connChan, listeningChann)
	}()

	listening, ok := <-listeningChann
	require.True(t, listening, "server should be listening")
	require.True(t, ok, "server should be listening")

	// Channel should be closed.
	listening, ok = <-listeningChann
	require.False(t, listening, "server should not be listening")
	require.False(t, ok, "server should not be listening")

	conn, err := net.Dial("tcp", ":1112")
	require.NoError(t, err)
	defer conn.Close()

	_, ok = <-connChan
	require.True(t, ok)

	// Should not be accepting connections anymore.
	conn, err = net.Dial("tcp", ":1112")
	if err == nil {
		defer conn.Close()
		t.Error("Expected 2nd dial to error")
	}

	_, ok = <-connChan
	require.False(t, ok)
}

func TestWs(t *testing.T) {
	t.Parallel()
	// TODO: fix this test.
	t.Skip("This does not consistently succeed")

	connChan := make(chan net.Conn)
	listeningChann := make(chan bool)
	go func() {
		connection.NewConnectionListener(connection.ConnWs, ":1113", connChan, listeningChann)
	}()

	listening, ok := <-listeningChann
	require.True(t, listening)
	require.True(t, ok)

	// Channel should be closed.
	listening, ok = <-listeningChann
	require.False(t, listening)
	require.False(t, ok)

	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:1113", nil)
	require.NoError(t, err)
	defer conn.Close()

	_, ok = <-connChan
	require.True(t, ok)

	conn, _, err = websocket.DefaultDialer.Dial("ws://127.0.0.1:1113", nil)
	if err == nil {
		defer conn.Close()
		t.Error("Expected 2nd dial to error")
	}

	_, ok = <-connChan
	require.False(t, ok)
}
