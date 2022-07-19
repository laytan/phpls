package connection

import (
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

type ConnType string

const (
	ConnStdio ConnType = "stdio"
	ConnWs    ConnType = "ws"
	ConnTcp   ConnType = "tcp"
)

func NewConnectionListener(
	connType ConnType,
	URL string,
	connChan chan<- net.Conn,
	listeningChann chan<- bool,
) {
	switch connType {
	case ConnWs:
		ws(URL, connChan, listeningChann)
		return
	case ConnTcp:
		tcp(URL, connChan, listeningChann)
		return
	case ConnStdio:
		listeningChann <- true
		close(listeningChann)

		connChan <- NewDefaultStdio()
		return
	}
}

func tcp(URL string, connChan chan<- net.Conn, listeningChann chan<- bool) {
	lis, err := net.Listen("tcp", URL)
	if err != nil {
		panic(err)
	}

	defer lis.Close()

	listeningChann <- true
	close(listeningChann)

	conn, err := lis.Accept()
	if err != nil {
		panic(err)
	}

	connChan <- conn
	close(connChan)
}

func ws(URL string, connChan chan<- net.Conn, listeningChann chan<- bool) {
	srv := http.Server{Addr: URL}

	upgrader := websocket.Upgrader{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}

		srv.Close()
		connChan <- NewWsConnAdapter(c)
		close(connChan)
	})

	go srv.ListenAndServe()
	listeningChann <- true
	close(listeningChann)
}
