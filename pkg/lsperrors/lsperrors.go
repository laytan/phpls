package lsperrors

import "github.com/jdbaldry/go-language-server-protocol/jsonrpc2"

var (
	// Error code indicating that a server received a notification or
	// request before the server has received the `initialize` request.
	ErrServerNotInitialized = jsonrpc2.NewError(-32002, "LSP Server not initialized")

	// A request failed but it was syntactically correct, e.g the
	// method name was known and the parameters were valid. The error
	// message should contain human readable information about why
	// the request failed.
	ErrRequestFailed = func(message string) error {
		return jsonrpc2.NewError(-32803, message)
	}

	// The server cancelled the request. This error code should
	// only be used for requests that explicitly support being
	// server cancellable.
	ErrServerCancelled = jsonrpc2.NewError(-32802, "LSP Server has cancelled the request")

	// The server detected that the content of a document got
	// modified outside normal conditions. A server should
	// NOT send this error code if it detects a content change
	// in it unprocessed messages. The result even computed
	// on an older state might still be useful for the client.
	ErrContentModified = jsonrpc2.NewError(
		-32801,
		"LSP Server detected modified content outside normal conditions",
	)

	// The client has canceled a request and a server as detected
	// the cancel.
	ErrRequestCancelled = jsonrpc2.NewError(
		-32800,
		"LSP Server has cancelled the request after client's cancel request",
	)
)
