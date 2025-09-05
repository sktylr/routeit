// Package routeit is a lightweight web framework for Go.
//
// # Copyright (c) 2025 Sam Taylor
//
// Licensed under the MIT License. You may obtain a copy of the License at
// https://opensource.org/licenses/MIT
//
// This package provides HTTP/1.1 request parsing, routing, middleware chaining,
// error handling, and testing utilities. It is designed as a learning framework
// for Go, demonstrating core web server patterns without external dependencies.
//
// Example usage:
//
//	server := routeit.NewServer(routeit.ServerConfig{Debug: true})
//	server.RegisterRoutes(routeit.RouteRegistry{
//		"/hello": func(rw *routeit.ResponseWriter, req *routeit.Request) error {
//			rw.Text("Hello, World!")
//			return nil
//		},
//	})
//	server.StartOrPanic()
package routeit
