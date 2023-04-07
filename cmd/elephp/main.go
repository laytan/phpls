package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/arl/statsviz"
	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/internal/logging"
	"github.com/laytan/elephp/internal/server"
	"github.com/laytan/elephp/pkg/connection"
	"github.com/laytan/elephp/pkg/processwatch"

	// TODO: what is the difference between jsonrpc2 and jsonrpc2_v2?
	"github.com/laytan/go-lsp-protocol/pkg/jsonrpc2"
	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "logs":
			config.Parse(os.Args[2:])
			_, _ = fmt.Println(config.Current.LogsPath)
			return
		case "stubs":
			config.Parse(os.Args[2:])
			_, _ = fmt.Println(config.Current.StubsPath)
			return
		}
	}

	config.Parse(os.Args[1:])
	stop := logging.Configure(config.Current.LogsPath)
	defer stop()

	_, _ = fmt.Printf(
		"Output will be going into the logs file at %q from now on\n",
		logging.LogsPath(config.Current.LogsPath),
	)

	if pid := config.Current.Server.ClientPid; pid != 0 {
		processwatch.NewExiter(pid)
		log.Printf("Monitoring process ID: %d", pid)
	}

	if config.Current.Statsviz.Enabled {
		go func() {
			log.Printf("Starting Statsviz at %s", config.Current.Statsviz.URL)
			if err := statsviz.RegisterDefault(); err != nil {
				log.Printf("Unable to register statsviz routes: %v", err)
			}

			log.Println(
				//nolint:gosec // This is ok because we use statsviz for locally visualizing, this is not opened to the internet.
				http.ListenAndServe(
					config.Current.Statsviz.URL,
					nil,
				),
			)
		}()
	}

	ctx := context.Background()

	connChan := make(chan net.Conn, 1)
	listeningChann := make(chan bool, 1)
	go func() {
		connection.NewConnectionListener(
			config.Current.Server.Communication,
			config.Current.Server.URL,
			connChan,
			listeningChann,
		)
	}()

	<-listeningChann
	log.Printf("Waiting for connection of type: %s\n", config.Current.Server.Communication)

	conn := <-connChan
	log.Println("Connected with client")

	stream := jsonrpc2.NewHeaderStream(conn)
	rpcConn := jsonrpc2.NewConn(stream)
	client := protocol.ClientDispatcher(rpcConn)
	server := server.NewServer(client)
	rpcConn.Go(ctx, protocol.Handlers(protocol.ServerHandler(server, jsonrpc2.MethodNotFound)))
	<-rpcConn.Done()
	if err := rpcConn.Err(); err != nil {
		log.Panicln(err)
	}
}
