package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/arl/statsviz"
	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/internal/server"
	"github.com/laytan/elephp/pkg/connection"
	"github.com/laytan/elephp/pkg/processwatch"

	// TODO: Check the difference between v1 and v2 of this
	"github.com/jdbaldry/go-language-server-protocol/jsonrpc2"
)

func main() {
	config := config.New()
	err := config.Initialize()
	if err != nil {
		log.Println(err.Error())
		return
	}

	connType, err := config.ConnType()
	if err != nil {
		log.Fatalf(err.Error())
	}

	if pid, isset := config.ClientPid(); isset {
		processwatch.New(pid, time.Second*10, func() {
			log.Fatal("The client process has exited, exiting elephp to")
		})
	}

	if config.UseStatsviz() {
		go func() {
			statsviz.RegisterDefault()
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	ctx := context.Background()

	connChan := make(chan net.Conn, 1)
	go func() { connection.NewConnectionListener(connType, config.ConnURL(), connChan, nil) }()
	conn := <-connChan

	stream := jsonrpc2.NewHeaderStream(conn)
	rpcConn := jsonrpc2.NewConn(stream)
	client := protocol.ClientDispatcher(rpcConn)
	server := server.NewServer(client)
	rpcConn.Go(ctx, protocol.Handlers(protocol.ServerHandler(server, jsonrpc2.MethodNotFound)))
	<-rpcConn.Done()
	if err := rpcConn.Err(); err != nil {
		panic(err)
	}
}
