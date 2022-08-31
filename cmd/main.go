package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"

	"github.com/arl/statsviz"
	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/internal/logging"
	"github.com/laytan/elephp/internal/server"
	"github.com/laytan/elephp/pkg/connection"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/processwatch"

	// TODO: Check the difference between v1 and v2 of this.
	"github.com/jdbaldry/go-language-server-protocol/jsonrpc2"
)

func main() {
	config := config.New()
	err := config.Initialize()
	if err != nil {
		os.Exit(1)
	}

	stop := logging.Configure(path.Join(pathutils.Root(), "logs"), config.Name())
	defer stop()

	connType, err := config.ConnType()
	if err != nil {
		log.Fatal(err)
	}

	if pid, isset := config.ClientPid(); isset {
		processwatch.NewExiter(pid)
		log.Printf("Monitoring process ID: %d\n", pid)
	}

	if config.UseStatsviz() {
		go func() {
			log.Println("Starting Statsviz at http://localhost:6060/debug/statsviz")
			if err := statsviz.RegisterDefault(); err != nil {
				log.Println(fmt.Errorf("Unable to register statsviz routes: %w", err))
			}

			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	ctx := context.Background()

	connChan := make(chan net.Conn, 1)
	listeningChann := make(chan bool, 1)
	go func() { connection.NewConnectionListener(connType, config.ConnURL(), connChan, listeningChann) }()

	<-listeningChann
	log.Printf("Waiting for connection of type: %s\n", connType)

	conn := <-connChan
	log.Println("Connected with client")

	stream := jsonrpc2.NewHeaderStream(conn)
	rpcConn := jsonrpc2.NewConn(stream)
	client := protocol.ClientDispatcher(rpcConn)
	server := server.NewServer(client, config)
	rpcConn.Go(ctx, protocol.Handlers(protocol.ServerHandler(server, jsonrpc2.MethodNotFound)))
	<-rpcConn.Done()
	if err := rpcConn.Err(); err != nil {
		log.Fatal(err)
	}
}
