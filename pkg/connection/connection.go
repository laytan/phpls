package connection

import (
	"errors"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

func NewConnectionListener(connType string, connChan chan<- net.Conn) error {
	switch connType {
	case "ws":
		ws(connChan)
		return nil
	case "tcp":
		tcp(connChan)
		return nil
	}

	return errors.New("no connChan listener for given connType")
}

func tcp(connChan chan<- net.Conn) {
	lis, err := net.Listen("tcp", "127.0.0.1:2001")
	if err != nil {
		panic(err)
	}

	defer func() {
		if lis == nil {
			return
		}

		if err = lis.Close(); err != nil {
			panic(err)
		}
	}()

	for {
		conn, err := lis.Accept()
		if err != nil {
			panic(err)
		}

		connChan <- conn
	}
}

func ws(connChan chan<- net.Conn) {
	upgrader := websocket.Upgrader{}
	http.ListenAndServe(":2001", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}

		connChan <- NewWsConnAdapter(c)
	}))
}
