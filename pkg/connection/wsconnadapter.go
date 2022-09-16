// Credits https://github.com/function61/holepunch-server/blob/master/pkg/wsconnadapter/wsconnadapter.go
// Gorilla websocket has 'conn.UnderlyingConn()' but that seems to work differently
// than this implementation and does not work.

// It does currently sent back 2 messages, one with the headers and one with the body.
// which should be fixed when that becomes a problem.

// Also, this is probably not the way an editor would call the lSP,
// but using websockets is very good for testing requests/connections by hand with Postman.

// an adapter for representing WebSocket connection as a net.Conn
// some caveats apply: https://github.com/gorilla/websocket/issues/441
package connection

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Adapter struct {
	conn       *websocket.Conn
	readMutex  sync.Mutex
	writeMutex sync.Mutex
	reader     io.Reader
}

func NewWsConnAdapter(conn *websocket.Conn) *Adapter {
	return &Adapter{
		conn: conn,
	}
}

func (a *Adapter) Read(b []byte) (int, error) {
	// Read() can be called concurrently, and we mutate some internal state here
	a.readMutex.Lock()
	defer a.readMutex.Unlock()

	if a.reader == nil {
		messageType, reader, err := a.conn.NextReader()
		if err != nil {
			return 0, fmt.Errorf("Error reading websocket message: %w", err)
		}

		if messageType != websocket.TextMessage {
			return 0, errors.New("unexpected websocket message type")
		}

		a.reader = reader
	}

	bytesRead, err := a.reader.Read(b)
	if err != nil {
		a.reader = nil

		// EOF for the current Websocket frame, more will probably come so..
		if err == io.EOF {
			// .. we must hide this from the caller since our semantics are a
			// stream of bytes across many frames
			err = nil
		}
	}

	return bytesRead, err
}

func (a *Adapter) Write(b []byte) (int, error) {
	a.writeMutex.Lock()
	defer a.writeMutex.Unlock()

	nextWriter, err := a.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return 0, fmt.Errorf("Error writing websocket message: %w", err)
	}

	defer nextWriter.Close()

	bytesWritten, err := nextWriter.Write(b)

	return bytesWritten, fmt.Errorf("Error writing websocket message: %w", err)
}

func (a *Adapter) Close() error {
	if err := a.conn.Close(); err != nil {
		return fmt.Errorf("Error closing websocket connection: %w", err)
	}

	return nil
}

func (a *Adapter) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *Adapter) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *Adapter) SetDeadline(t time.Time) error {
	if err := a.SetReadDeadline(t); err != nil {
		return err
	}

	return a.SetWriteDeadline(t)
}

func (a *Adapter) SetReadDeadline(t time.Time) error {
	if err := a.conn.SetReadDeadline(t); err != nil {
		return fmt.Errorf("Error setting read deadline: %w", err)
	}

	return nil
}

func (a *Adapter) SetWriteDeadline(t time.Time) error {
	if err := a.conn.SetWriteDeadline(t); err != nil {
		return fmt.Errorf("Error setting write deadline: %w", err)
	}

	return nil
}
