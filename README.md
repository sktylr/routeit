## routeit [![Go Reference](https://pkg.go.dev/badge/github.com/sktylr/routeit.svg)](https://pkg.go.dev/github.com/sktylr/routeit) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/rs/cors/master/LICENSE) [![routeit](https://github.com/sktylr/routeit/actions/workflows/main.yml/badge.svg)](https://github.com/sktylr/routeit/actions/workflows/main.yml) [![examples](https://github.com/sktylr/routeit/actions/workflows/examples.yml/badge.svg)](https://github.com/sktylr/routeit/actions/workflows/examples.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/sktylr/routeit)](https://goreportcard.com/report/github.com/sktylr/routeit)

`routeit` is a lightweight web framework built in Go that conforms to the [HTTP/1.1](https://www.rfc-editor.org/rfc/rfc9112.html) spec.
This is meant to be exploratory and educational and is not fit for production purposes.
`routeit` can be seen as a simpler version of [`net/http`](https://pkg.go.dev/net/http) that allows for building simple servers and has been built from the ground up with no usage of non-standard libraries.
It includes some features, such as parameterised dynamic routing, built-in error handling and JSON handling that `net/http` does not handle out of the box.

### Getting Started

Before beginning, make sure you have installed Go and setup your [GOPATH](http://golang.org/doc/code.html#GOPATH) correctly.
Create a `.go` file called `server.go` and add the following code:

```go
package main

import "github.com/sktylr/routeit"

type HelloWorld struct {
	Hello string `json:"hello"`
}

func main() {
	srv := routeit.NewServer(routeit.ServerConfig{Debug: true})
	srv.RegisterRoutes(routeit.RouteRegistry{
		"/hello": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			body := HelloWorld{Hello: "World"}
			return rw.Json(body)
		}),
	})
	srv.StartOrPanic()
}
```

Install `routeit` from the command line:

```bash
$ go get github.com/sktylr/routeit
```

Run the server:

```bash
$ go run server.go
```

The server is now running on `localhost:8080` and will respond to `GET` requests on the `/hello` endpoint.

```bash
$ curl -D - http://localhost:8080/hello
HTTP/1.1 200 OK
Content-Type: application/json
Date: Mon, 01 Sep 2025 12:06:32 GMT
Server: routeit
Content-Length: 17

{"hello":"World"}
```

Check out the [`examples/`](/examples/) directory for further examples of using `routeit`'s features.

### Documentation

Documentation for the latest released version can be found on [`pkg.go.dev/github.com/sktylr/routeit`](https://pkg.go.dev/github.com/sktylr/routeit).
The repository also contains a [`docs/`](/docs/) directory that contains notes and further details regarding the library.

Documentation for the current development version can be generated using [`godoc`](https://pkg.go.dev/golang.org/x/tools/cmd/godoc) and requires the repository to be cloned.

```bash
# Install the package if not already installed
$ go install golang.org/x/tools/cmd/godoc@latest

# Run the documentation server on port 3000
$ godoc -http=:3000
```

The documentation can now be viewed at [`localhost:3000/pkg/github.com/sktylr/routeit/`](http://localhost:3000/pkg/github.com/sktylr/routeit/).

### Features

**HTTP Version Support**: Only HTTP/1.1 is supported. My implementation is mostly based off [RFC-9112](https://httpwg.org/specs/rfc9112.html) and [Mozilla](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference) developer specs.

| HTTP Method | Supported? | Notes                                                                                                                           |
| ----------- | ---------- | ------------------------------------------------------------------------------------------------------------------------------- |
| GET         | ✅         |                                                                                                                                 |
| HEAD        | ✅         | Cannot be implemented by the integrator, it is baked into the server implementation.                                            |
| POST        | ✅         |                                                                                                                                 |
| PUT         | ✅         |                                                                                                                                 |
| DELETE      | ✅         |                                                                                                                                 |
| CONNECT     | ❌         | Will never be implemented since I will not support tunnelling                                                                   |
| OPTIONS     | ✅         | Baked into the server implementation.                                                                                           |
| TRACE       | ✅         | Baked into the server implementation but is defaulted OFF. Can be turned on using the `AllowTraceRequests` configuration option |
| PATCH       | ✅         |                                                                                                                                 |

| Content Types      | Request supported? | Response supported? | Notes                                                                                                                                                                                                              |
| ------------------ | ------------------ | ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `application/json` | ✅                 | ✅                  | Parsing and encoding is handled automatically by `routeit`                                                                                                                                                         |
| `text/plain`       | ✅                 | ✅                  |                                                                                                                                                                                                                    |
| ...                | ✅                 | ✅                  | Any request or response type can be supported, but the integrator must handling the parsing and marshalling. The `ResponseWriter.RawWithContentType` and `Request.BodyFromRaw` methods can be used correspondingly |

#### HTTPS

`routeit` supports both HTTP and HTTPS.
Unlike with `net/http` and other libraries, a single `routeit` server can support both HTTP and HTTPS without needing to explicitly manage the separate threads and ports.
The `HttpConfig` allows the HTTP(s) ports to be specified, as well as the TLS config required if a server wants to accept HTTPS communication.

TLS is backed by the [`crypto/tls`](https://pkg.go.dev/crypto/tls) standard library, which is the same used in `net/http`.
This provides sensible out-of-the-box defaults while allowing high levels of cusomisation if required.
Once a TLS config is provided to a `routeit` server, the server will **only** listen for HTTPS communication, unless plain HTTP communication is explicitly enabled.
`routeit` also comes with built-in HTTPS upgrade mechanisms, which will instruct clients to upgrade their connections to HTTPS before they will be accepted, which is controlled through the `HttpConfig.UpgradeToHttps` and `HttpConfig.UpgradeInstructionMaxAge` properties.
Check out [`examples/https`](/examples/https/) for example setups showcasing each of the 3 configuration options that use HTTPS.

#### Handlers

Each resource on the server is served by a `Handler`.
Handlers can respond to multiple HTTP methods, or just one.
`routeit` handles method delegation and will respond with a `405: Method Not Allowed` response with the correct `Allow` header if the resource exists but does not respond to the request method.

`Get`, `Post`, `Put`, `Patch` and `Delete` can all be used to construct a handler that responds to a single method, and accept a function of `func(*routeit.ResponseWriter, *routeit.Request) error` signature.
If an endpoint should respond to multiple HTTP methods, `MultiMethod` can be used, which accepts a struct to allow selection of the methods the handler should respond to.
The `error` returned from the handler function does not need to be a specific `routeit` error in every situation.

#### Middleware

`routeit` gives the developer the ability to write custom middleware to perform actions such as rate-limiting or authorisation handling.
Examples can be found in [`examples/middleware`](/examples/middleware).
Linking is performed through a `Chain` interface which is passed as an argument to the middleware function.
Multiple middleware functions can be attached to a single server.
The order of attachment is important, as that is the order used when processing the middleware for each incoming request.
Middleware can choose to block a request (by not invoking `Chain.Proceed`) but be aware that the server will always attempt to send a response to the client for every incoming request.
It is more common to block a request by returning a specific error that can be conformed to a HTTP response.

#### Routing

`routeit` supports both static and dynamic routing, as well as allowing for enforcing specific prefixes and/or suffixes to be part of a dynamic match.
Routes are registered to the server using `Server.RegisterRoutes` and `Server.RegisterRoutesUnderNamespace` using a map from route to handler.
The server setup allows for namespaces - both global and local.
These reduce the complexity of the routing setup by avoiding redundant duplication.
For example, we might set the server's global namespace to `"api"`, meaning that all routes are registered with `/api` as the prefix, without needing to write it out for each route.

Dynamic matches are denoted using colon (`:`) syntax, with the characters after the colon used as the name for the dynamic component.
A dynamic component can be forced to contain a required prefix and/or suffix using pipe (`|`) notation, as shown in the example below.
Once matched, the handler (or middleware) can use `Request.PathParam` to extract the name path segment.
Critically, `Request.PathParam` will always return a non-empty string so long as the provided name does appear in the matched path.
If prefixes or suffixes are enforced, the path component will be the entire component - including a prefix and/or suffix.

```go
"/:foo|pre/bar/:baz||suf/:qux/:corge|pre|suf": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
	// The first path segment - this must be prefixed by "pre" to match
	foo := req.PathParam("foo")
	// The third path segment - this must end with "suf" to match
	baz := req.PathParam("baz")
	// The fourth path segment
	qux := req.PathParam("qux")
	// The fifth path segment - this must start with "pre" and end with "suf",
	// with at least 1 character between them to match
	corge := req.PathParam("corge")

	// ...
	return nil
})
```

In the above example, the router would invoke this handler if a URI such as `/prefix/bar/my-suf/anything/prefix-then-suf` was received at the edge.
In this case, `foo` would be `"prefix"`, `baz` would be `"my-suf"`, `qux` would be `"anything"` and `corge` would be `"prefix-then-suf"`.

Routes can also be rewritten before being routed or processed.
For example, we might want to rewrite `/` to `/static/index.html`, as `/` is what the browser will request and is a much cleaner URI than the actual URI needed for the resource.
Rewriting is covered in [`docs/rewrites.md`](/docs/rewrites.md) as well as by example in [`examples/static/rewrites`](/examples/static/rewrites/).

Further details about the structure of routing can be found in [`docs/trie.md`](/docs/trie.md), which also covers key information such as prioritisation of routing if multiple routes can match the incoming URI.

#### Testing

Testing is baked into the `routeit` library and can be used to increase confidence in the server.
There are two level of tests supported, both of which are currently experimental.

The `TestClient` allows for E2E-like tests and operates on a full server instance.
To reduce flakiness, TCP connections are not opened, however every other piece of the server is tested - parsing, URI rewriting, routing, middleware, handling, error management and static asset loading.
Making a request with a `TestClient` instance will return a `TestResponse` which allows assertions to be made on the status code, headers, and response body.
Usage of `TestClient` is recommended for high-level validation that all moving parts of the server work as expected, without caring about the specific implementation details.
Each example project in [`examples/`](/examples/) contains E2E tests using `TestClient`.

For finer isolation, middleware and handlers can both be tested independently, using `TestMiddleware` and `TestHandler` respectively.
Both functions accept a handler or middleware instance, and a `TestRequest`, which can be constructed using `NewTestRequest`.
They will also return an `error` and `TestResponse` for making assertions.
In the case of testing middleware, a boolean is also returned to indicate whether the middleware proceeded to the next piece of middleware or not.

A key point to note is that, unlike with `TestClient`, the `error` is not run through any error mapping or handling.
So if the `error` is returned from `TestHandler` or `TestMiddleware`, it is the exact `error` that the corresponding handler or middleware returned.
In these cases, `TestResponse` will **not** be `nil`, but most meaningful assertions will not pass, unless the handler or middleware explicitly wrote a header or response body before returning an `error`.

#### Errors

Application code can return errors of any type to the library in their handlers.
A number of helpful error functions are exposed which allow the application code to conform their errors to HTTP responses.
If non-library errors are returned (or the application code panics with an error), we attempt to infer the reason or cause and map that to a HTTP error.
The integrator can provide a custom mapper using the `ServerConfig.ErrorMapper` which can provide additional custom logic in mapping from an `error` type to an error the server understands.
If the `ErrorMapper` cannot assign a more appropriate error type, it can return `nil` which will pass off the default inference which maps common errors to sensible defaults.
For example, if an `ErrNotExist` error is returned, we map that to a `404: Not Found` HTTP error.
We fallback to mapping to a `500: Internal Server Error` if we cannot establish a mapping.

Additionally, custom handling can be provided for specific HTTP status codes, if `routeit`'s default response is not sufficient.
This allows for additional logging, new headers, or transformation of the response body to something more meaningful than `routeit`'s default response, for example.
Common use cases include for `404: Not Found`, and `500: Internal Server Error`.
These can be registered using `Server.RegisterErrorHandlers`.

[`examples/errors`](/examples/errors/) contains examples for how custom error handling and mapping can be performed using `routeit`.

#### Logging

Each valid incoming request is logged with the corresponding method, path (both edge and rewritten) and response status.
4xx responses are logged using the `WARN` log level, 5xx responses are logged using `ERROR` and all other responses are logged using `INFO`.
Only requests that can be successfully parsed are logged.
