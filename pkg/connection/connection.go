package connection

import (
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

type ConnType uint

const (
    ConnStdio ConnType = iota
    ConnWs
    ConnTcp
)


func NewConnectionListener(connType ConnType, URL string, connChan chan<- net.Conn) {
	switch connType {
	case ConnWs:
		ws(URL, connChan)
		return
	case ConnTcp:
		tcp(URL, connChan)
		return
	case ConnStdio:
		connChan <- NewDefaultStdio()
		return 
	}
}

func tcp(URL string, connChan chan<- net.Conn) {
	lis, err := net.Listen("tcp", URL)
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

func ws(URL string, connChan chan<- net.Conn) {
	upgrader := websocket.Upgrader{}
	http.ListenAndServe(URL, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}

		connChan <- NewWsConnAdapter(c)
	}))
}
