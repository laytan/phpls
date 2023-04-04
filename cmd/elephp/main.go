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
			defaultConf := config.Default()
			_, _ = fmt.Println(defaultConf.LogsDir())
			return
		case "stubs":
			defaultConf := config.Default()
			_, _ = fmt.Println(defaultConf.StubsDir())
			return
		case "bin":
			defaultConf := config.Default()
			_, _ = fmt.Println(defaultConf.BinDir())
			return
		}
	}

	conf := config.New()
	config.Current = conf

	disregardErr, err := conf.Initialize()
	if disregardErr {
		_, _ = fmt.Println(err)
		return
	}

	if err != nil {
		_, _ = fmt.Println(err)
		return
	}

	stop := logging.Configure(conf.LogsDir())
	defer stop()

	_, _ = fmt.Printf(
		"Output will be going into the logs file at \"%s\" from now on",
		logging.LogsPath(conf.LogsDir()),
	)

	connType, err := conf.ConnType()
	if err != nil {
		log.Println(err)
		return
	}

	if pid, isset := conf.ClientPid(); isset {
		processwatch.NewExiter(pid)
		log.Printf("Monitoring process ID: %d", pid)
	}

	if conf.UseStatsviz() {
		go func() {
			log.Println("Starting Statsviz at http://localhost:6060/debug/statsviz")
			if err := statsviz.RegisterDefault(); err != nil {
				log.Println(fmt.Errorf("Unable to register statsviz routes: %w", err))
			}

			log.Println(
				//nolint:gosec // This is ok because we use statsviz for locally visualizing, this is not opened to the internet.
				http.ListenAndServe(
					"localhost:6060",
					nil,
				),
			)
		}()
	}

	ctx := context.Background()

	connChan := make(chan net.Conn, 1)
	listeningChann := make(chan bool, 1)
	go func() { connection.NewConnectionListener(connType, conf.ConnURL(), connChan, listeningChann) }()

	<-listeningChann
	log.Printf("Waiting for connection of type: %s\n", connType)

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
