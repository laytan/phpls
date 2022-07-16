package main

import (
	"context"
	"errors"
	"log"
	"net"
	"time"

	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/jessevdk/go-flags"
	"github.com/laytan/elephp/internal/server"
	"github.com/laytan/elephp/pkg/connection"

	// TODO: Check the difference between v1 and v2 of this
	"github.com/jdbaldry/go-language-server-protocol/jsonrpc2"
)

var (
    ErrIncorrectConnTypeAmt = errors.New("Elephp requires exactly one connection type to be selected")
)

type Opts struct {
	ClientProcessId uint16 `long:"clientProcessId" description:"Process ID that when terminated, terminates the language server"`
	UseStdio        bool   `long:"stdio" description:"Communicate over stdio"`
	UseWs           bool   `long:"ws" description:"Communicate over websockets"`
	UseTcp          bool   `long:"tcp" description:"Communicate over TCP"`
	URL             string `long:"url" description:"The URL to listen on for tcp or websocket connections" default:"127.0.0.1:2001"`
}

func (o *Opts) ConnType() (connection.ConnType, error) {
    connTypes := map[connection.ConnType]bool{
        connection.ConnStdio: o.UseStdio,
        connection.ConnTcp: o.UseTcp,
        connection.ConnWs: o.UseWs,
    }
    
    var result connection.ConnType
    var found bool
    for connType, selected := range connTypes {
        if !selected {
            continue
        }

        if found {
        return result, ErrIncorrectConnTypeAmt
    }

        result = connType
        found = true
    }

    if !found {
        return result, ErrIncorrectConnTypeAmt
    }

    return result, nil
}

func main() {
	opts := Opts{}
	_, err := flags.Parse(&opts)
	if err != nil {
		return
	}

    connType, err := opts.ConnType()
    if err != nil {
        log.Fatalf(err.Error())
    }

	ctx := context.Background()

	connChan := make(chan net.Conn, 1)
	go func() { connection.NewConnectionListener(connType, opts.URL, connChan); }()
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
