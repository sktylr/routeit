package socket

import "net"

// A [Socket] controls connections over the network.
type Socket interface {
	// Binds the socket to the network configuration. This will return an error
	// if the configuration conflicts with any existing configurations, for
	// example if a port is in use.
	Bind() error

	// The socket will start accepting traffic, and will pipe the connections
	// to the provided connection consumer. Each new connection is handled
	// inside its own go-routine.
	Serve(onConnection, onError)

	// [Close] should be called to close the underlying networking socket(s)
	// opened by the [Socket]. This may return an error, for example if the
	// closing failed, or if the socket has not been bound correctly.
	Close() error
}

// [onConnection] consumes a [net.Conn] and is invoked in a new go-routine
// whenever a socket accepts a new connection. The expectation is that the
// consumer of the connection will close it properly after they have finished
// with it, and should manage read and write deadlines themselves.
type onConnection func(net.Conn)

// [onError] is called whenever the attempted acceptance of a new connection
// results in an error. The function will be called inside its own go-routine.
type onError func(error)
