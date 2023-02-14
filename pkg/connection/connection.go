package connection

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type ConnType string

const (
	ConnStdio ConnType = "stdio"
	ConnWs    ConnType = "ws"
	ConnTCP   ConnType = "tcp"
)

func NewConnectionListener(
	connType ConnType,
	url string,
	connChan chan<- net.Conn,
	listeningChann chan<- bool,
) {
	switch connType {
	case ConnWs:
		ws(url, connChan, listeningChann)
		return
	case ConnTCP:
		tcp(url, connChan, listeningChann)
		return
	case ConnStdio:
		listeningChann <- true
		close(listeningChann)

		connChan <- NewDefaultStdio()
		close(connChan)
		return
	}
}

func tcp(url string, connChan chan<- net.Conn, listeningChann chan<- bool) {
	lis, err := net.Listen("tcp", url)
	if err != nil {
		log.Panic(err)
	}

	defer lis.Close()

	listeningChann <- true
	close(listeningChann)

	conn, err := lis.Accept()
	if err != nil {
		log.Panicln(err)
	}

	connChan <- conn
	close(connChan)
}

func ws(url string, connChan chan<- net.Conn, listeningChann chan<- bool) {
	srv := http.Server{Addr: url, ReadHeaderTimeout: time.Second}

	upgrader := websocket.Upgrader{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Panic(err)
		}

		if err = srv.Close(); err != nil {
			log.Println(fmt.Errorf("Error closing WS HTTP server: %w", err))
		}

		connChan <- NewWsConnAdapter(c)
		close(connChan)
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	listeningChann <- true
	close(listeningChann)
}
