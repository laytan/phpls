package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/arl/statsviz"
	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/internal/logging"
	"github.com/laytan/elephp/internal/server"
	"github.com/laytan/elephp/pkg/connection"
	"github.com/laytan/elephp/pkg/processwatch"
	log "github.com/sirupsen/logrus"

	// TODO: Check the difference between v1 and v2 of this.
	"github.com/jdbaldry/go-language-server-protocol/jsonrpc2"
)

func main() {
	if os.Args[1] == "logs" {
		if err := logging.Tail(); err != nil {
			panic(err)
		}

		return
	}

	config := config.New()
	err := config.Initialize()
	if err != nil {
		os.Exit(1)
	}

	logging.Configure(config)

	connType, err := config.ConnType()
	if err != nil {
		log.Fatal(err)
	}

	if pid, isset := config.ClientPid(); isset {
		processwatch.NewExiter(pid)
		log.Infof("Monitoring process ID: %d\n", pid)
	}

	if config.UseStatsviz() {
		go func() {
			log.Infoln("Starting Statsviz at http://localhost:6060/debug/statsviz")
			if err := statsviz.RegisterDefault(); err != nil {
				log.Error(fmt.Errorf("Unable to register statsviz routes: %w", err))
			}

			log.Error(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	ctx := context.Background()

	connChan := make(chan net.Conn, 1)
	listeningChann := make(chan bool, 1)
	go func() { connection.NewConnectionListener(connType, config.ConnURL(), connChan, listeningChann) }()

	<-listeningChann
	log.Infof("Waiting for connection of type: %s\n", connType)

	conn := <-connChan
	log.Infoln("Connected with client")

	stream := jsonrpc2.NewHeaderStream(conn)
	rpcConn := jsonrpc2.NewConn(stream)
	client := protocol.ClientDispatcher(rpcConn)
	server := server.NewServer(client)
	rpcConn.Go(ctx, protocol.Handlers(protocol.ServerHandler(server, jsonrpc2.MethodNotFound)))
	<-rpcConn.Done()
	if err := rpcConn.Err(); err != nil {
		log.Fatal(err)
	}
}
