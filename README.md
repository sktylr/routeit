### RouteIt

RouteIt is a lightweight web framework built in go.
It is designed as an introduction to Go to help me learn the language.
The goal of this is to build a framework similar to the already excellent [`net/http`](https://pkg.go.dev/net/http) package myself, and to avoid using any non-standard libraries.
The only usages of the `net/http` library in my own framework are to establish the socket connection and consume requests and write responses.
All parsing and routing logic is handled by my framework.

This library is not meant to be production ready :).

The source and test code can be found in `/src`.
Where possible I have written detailed comments explaining usage of framework's API.

"Real-world" examples are also included in the `/examples` directory.
I add to these as new features are built or improved upon, to showcase how the interfaces are intended to be used.
This also helps me understand how I should design my interfaces, as I get hands on experience using them.

### Documentation

Documentation for this package can be generated using [`godoc`](https://pkg.go.dev/golang.org/x/tools/cmd/godoc). Steps for viewing the docs are below

```bash
# Install the package if not already installed
$ go install golang.org/x/tools/cmd/godoc@latest

# Change to the source directory
$ cd src

# Run the documentation server on port 3000
$ godoc -http=:3000
```

The documentation can now be viewed at http://localhost:3000/pkg/github.com/sktylr/routeit/.

If your `$GOPATH` is not set, this may fail to run. The `$GOPATH` defaults to `$HOME/go` but go can sometimes have difficulty due to `src` containing a go module. Wherever your go binaries are installed needs to be in your `$PATH`, or you can reference the absolute path when running `godoc -http=:3000`.

```bash
# Setting GOPATH explicitly
$ export GOPATH=~/go
$ export PATH="$PATH:$GOPATH/bin"
$ godoc -http=:3000

# Using absolute path
$ /abs/path/to/godoc -http=:3000
```

### Features

**HTTP Version Support**: Only HTTP/1.1 is supported. My implementation is mostly based off https://httpwg.org/specs/rfc9112.html and [Mozilla](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference) developer specs.

| HTTP Method | Supported? | Notes                                                                                |
| ----------- | ---------- | ------------------------------------------------------------------------------------ |
| GET         | ✅         |                                                                                      |
| HEAD        | ✅         | Cannot be implemented by the integrator, it is baked into the server implementation. |
| POST        | ✅         |                                                                                      |
| PUT         | ✅         |                                                                                      |
| DELETE      | ❌         |                                                                                      |
| CONNECT     | ❌         | Will never be implemented since I will not support HTTPS                             |
| OPTIONS     | ✅         | Baked into the server implementation.                                                |
| TRACE       | ❌         |                                                                                      |
| PATCH       | ❌         |                                                                                      |

If the server has a valid route for the request, but the route does not respond to the requested method, the server will return a `405: Method Not Allowed` response with the `Allow` header populated to indicate which methods are supported.

| Content Types      | Request supported? | Response supported? | Notes                                                                                                                                                 |
| ------------------ | ------------------ | ------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| `application/json` | ✅                 | ✅                  | Parsing and encoding is handled automatically by `routeit`                                                                                            |
| `text/plain`       | ❌                 | ✅                  |                                                                                                                                                       |
| ...                | ❌                 | ✅                  | Any response type can be supported, but the integrator must convert the response body to bytes and use the `ResponseWriter.RawWithContentType` method |

#### Status codes

All status codes in the official HTTP/1.1 spec are supported.
They are currently all exposed to the integrator, meaning that the application developer can choose to return any of the status types.

#### Errors

Application code can return errors of any type to the library in their handlers.
A number of helpful error functions are exposed which allow the application code to conform their errors to HTTP responses.
If non-library errors are returned (or the application code panics), we attempt to infer the reason or cause and map that to a HTTP error.
For example, if an `ErrNotExist` error is returned, we map that to a 404: Not Found HTTP error.
We fallback to mapping to a 500: Internal Server Error if we cannot establish a mapping.

#### Routing

Routing is currently handled using a trie-like structure.
Typically tries are separated at the character level, but in my case I separate at the path separators (`/`), so each node contains a path segment.
Currently only static paths are supported.

#### Testing

The framework supports a simple testing paradigm that allows user to perform E2E-like tests on their server.
A `TestClient` is provided which takes a `Server` structure as an argument and performs nearly all of the server work that a running server would perform.
To reduce flakiness, the test server does not actually start the server and open TCP connections for each request.
Instead, the raw request is handed off to the server directly.
This means that parsing, routing and handling are all still performed under the test, giving a near E2E feel for the tests.
The response that is returned by the test client features intuitive methods to perform assertions on the response itself, including the status code and body.
The test client will handle all panics or errors reported by the application, so there is no need to use a defer-recover block to handle expected panics within the code.
Examples of how to use the testing API can be found in the [`examples`](/examples) directory.
Each example project in this directory features tests, which give me a place to explore how I would like testing to work, while also providing an indicator if any bugs or regressions are introduced.

#### Middleware

`routeit` gives the developer the ability to write custom middleware to perform actions such as rate-limiting or authorisation handling.
Examples can be found in [`examples/middleware`](/examples/middleware).
Linking is performed through a `Chain` struct which is passed as an argument to the middleware function.
Multiple middleware functions can be attached to a single server.
The order of attachment is important, as that is the order used when processing the middleware for each incoming request.
