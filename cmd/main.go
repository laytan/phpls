package main

import (
	"context"
	"net"

	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/internal/server"
	"github.com/laytan/elephp/pkg/connection"

	// TODO: Check the difference between v1 and v2 of this
	"github.com/jdbaldry/go-language-server-protocol/jsonrpc2"
)

func main() {
	ctx := context.Background()

	// TODO: Based on cmdline args and flags,
	// TODO: also make listener port etc configurable.
	// See: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#implementationConsiderations
	connType := "ws"
	connChan := make(chan net.Conn, 1)
	go func() {
		if err := connection.NewConnectionListener(connType, connChan); err != nil {
			panic(err)
		}
	}()
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
