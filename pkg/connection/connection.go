package connection

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type ConnType string

const (
	ConnStdio ConnType = "stdio"
	ConnWs    ConnType = "ws"
	ConnTCP   ConnType = "tcp"
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
	case ConnTCP:
		tcp(URL, connChan, listeningChann)
		return
	case ConnStdio:
		listeningChann <- true
		close(listeningChann)

		connChan <- NewDefaultStdio()
		close(connChan)
		return
	}
}

func tcp(URL string, connChan chan<- net.Conn, listeningChann chan<- bool) {
	lis, err := net.Listen("tcp", URL)
	if err != nil {
		log.Fatal(err)
	}

	defer lis.Close()

	listeningChann <- true
	close(listeningChann)

	conn, err := lis.Accept()
	if err != nil {
		log.Fatal(err)
	}

	connChan <- conn
	close(connChan)
}

func ws(URL string, connChan chan<- net.Conn, listeningChann chan<- bool) {
	srv := http.Server{Addr: URL, ReadHeaderTimeout: time.Second}

	upgrader := websocket.Upgrader{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatal(err)
		}

		if err = srv.Close(); err != nil {
			log.Error(fmt.Errorf("Error closing WS HTTP server: %w", err))
		}

		connChan <- NewWsConnAdapter(c)
		close(connChan)
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error(err)
		}
	}()

	listeningChann <- true
	close(listeningChann)
}
