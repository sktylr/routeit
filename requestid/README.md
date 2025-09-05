## routeit/requestid

`requestid` is a package under the [`routeit`](https://github.com/sktylr/routeit) module.
`routeit` allows users to define `RequestIdProvider`s that generate an ID for incoming requests.

This package provides a convenient implementation using V7 UUIDs.

### Example usage

```go
package main

import (
	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/requestid"
)

type HelloWorld struct {
	Hello string `json:"hello"`
}

func main() {
	srv := routeit.NewServer(routeit.ServerConfig{
		Debug: true,
		RequestIdProvider: requestid.NewUuidV7Provider(),
	})
	srv.RegisterRoutes(routeit.RouteRegistry{
		"/hello": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			body := HelloWorld{Hello: "World"}
			return rw.Json(body)
		}),
	})
	srv.StartOrPanic()
}
```
