## routeit

`routeit` is a lightweight web framework built in go.
It is designed as an introduction to Go to help me learn the language.
The goal of this is to build a framework similar to the already excellent [`net/http`](https://pkg.go.dev/net/http) package myself, and to avoid using any non-standard libraries.
The only usages of the `net/http` library in my own framework are to establish the socket connection and consume requests and write responses.
All parsing and routing logic is handled by my framework.

This library is not meant to be production ready :).

The source and test code can be found in [`/src`](/src).
Where possible I have written detailed comments explaining usage of framework's API.

"Real-world" examples are also included in the [`/examples`](/examples/) directory.
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

The documentation can now be viewed at [`localhost:3000/pkg/github.com/sktylr/routeit/`](http://localhost:3000/pkg/github.com/sktylr/routeit/).

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

**HTTP Version Support**: Only HTTP/1.1 is supported. My implementation is mostly based off [RFC-9112](https://httpwg.org/specs/rfc9112.html) and [Mozilla](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference) developer specs.

| HTTP Method | Supported? | Notes                                                                                                                           |
| ----------- | ---------- | ------------------------------------------------------------------------------------------------------------------------------- |
| GET         | ✅         |                                                                                                                                 |
| HEAD        | ✅         | Cannot be implemented by the integrator, it is baked into the server implementation.                                            |
| POST        | ✅         |                                                                                                                                 |
| PUT         | ✅         |                                                                                                                                 |
| DELETE      | ✅         |                                                                                                                                 |
| CONNECT     | ❌         | Will never be implemented since I will not support HTTPS                                                                        |
| OPTIONS     | ✅         | Baked into the server implementation.                                                                                           |
| TRACE       | ✅         | Baked into the server implementation but is defaulted OFF. Can be turned on using the `AllowTraceRequests` configuration option |
| PATCH       | ✅         |                                                                                                                                 |

If the server has a valid route for the request, but the route does not respond to the requested method, the server will return a `405: Method Not Allowed` response with the `Allow` header populated to indicate which methods are supported.

| Content Types      | Request supported? | Response supported? | Notes                                                                                                                                                                                                              |
| ------------------ | ------------------ | ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `application/json` | ✅                 | ✅                  | Parsing and encoding is handled automatically by `routeit`                                                                                                                                                         |
| `text/plain`       | ✅                 | ✅                  |                                                                                                                                                                                                                    |
| ...                | ✅                 | ✅                  | Any request or response type can be supported, but the integrator must handling the parsing and marshalling. The `ResponseWriter.RawWithContentType` and `Request.BodyFromRaw` methods can be used correspondingly |

#### Status codes

All status codes in the official HTTP/1.1 spec are supported.
They are currently all exposed to the integrator, meaning that the application developer can choose to return any of the status types.

#### Errors

Application code can return errors of any type to the library in their handlers.
A number of helpful error functions are exposed which allow the application code to conform their errors to HTTP responses.
If non-library errors are returned (or the application code panics with an error), we attempt to infer the reason or cause and map that to a HTTP error.
The integrator can provide a custom mapper using the `ServerConfig.ErrorMapper` which can provide additional custom logic in mapping from an `error` type to an error the server understands.
If the `ErrorMapper` cannot assign a more appropriate error type, it can return `nil` which will pass off the default inference which maps common errors to sensible defaults.
For example, if an `ErrNotExist` error is returned, we map that to a 404: Not Found HTTP error.
We fallback to mapping to a 500: Internal Server Error if we cannot establish a mapping.

Additionally, custom handling can be provided for specific HTTP status codes, if `routeit`'s default response is not sufficient.
Common use cases include for `404: Not Found`, and `500: Internal Server Error`.
These can be registered using `routeit.Server.RegisterErrorHandlers`.

[`examples/errors`](/examples/errors/) contains examples for how custom error handling can be performed using `routeit`.

#### Routing

Routing is handled using a trie-like structure.
More information on the underlying trie can be found in [`docs/trie.md`](/docs/trie.md).

Routing supports static and dynamic matching, with additional control over required prefixes and suffixes in the dynamic path components.
Dynamic components are denoted with a leading `:`, followed by the name they should be looked up by and optionally followed by the required prefix and suffix they should match against, separated by `|`.
Below is an example of a setup of a route that matches against `/pre<anything>/bar/<anything>suffix`.
Given an input `"/prefix/bar/my-suffix"`, `foo` would be `"prefix"`, and `bar` would be `"my-suffix"`.

```golang
"/:foo|pre/bar/:baz||suffix": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
	foo, _ := req.PathParam("foo")
	baz, _ := req.PathParam("baz")

	// ...
	return nil
})
```

Server setup will panic if the routing is misconfigured in some way.
Examples for routing can be found in [`examples/routing`](/examples/routing/).
Pay attention to the gotchas mentioned in [`docs/trie.md`](/docs/trie.md) when configuring routes 😉.

#### URL Rewrites

`routeit` allows the integrator to define rewrite rules for incoming URLs.
This is commonly preferred in servers that serve static content, as the raw URLs for static content is `<static directory>/<file name>.<extension>`, which is quite ugly.
For example, the convention for HTML servers is to define the home page in `static/index.html`, which means the landing page of the website would be `http://url.com/static.index.html`.

Rewrites allow us to change that to `http://url.com/`, which is much easier.
Additionally, the links used within the static content (e.g. JavaScript, HTML, stylesheets, images) can use the rewritten URL, meaning it is easier to swap components out.
If you want to change the `/about` page from `/static/about1.html` to `/static/about2.html`, you can use `/about` as the link referenced in all static content, then just change the URL rewrite rule and the change is made across the entire server.

Rewriting is covered in more depth in [`docs/rewrites.md`](/docs/rewrites.md) and examples of rewriting can be found in [`examples/static/rewrites`](/examples/static/rewrites/).

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

Middleware can also be tested in isolation, using a combination of `routeit.NewTestRequest` and `routeit.TestMiddleware`.
Under isolated test conditions, the middleware does not actually proceed to the next in the chain, but allows you to assert on the response, chain and error returned.
See [`examples/middleware/simple/main_test.go`](/examples/middleware/simple/main_test.go) and [`examples/middleware/complex/main_test.go`](/examples/middleware/complex/main_test.go) for examples of how to use this interface.

#### Logging

Each valid incoming request is logged with the corresponding method, path (both edge and rewritten) and response status.
4xx responses are logged using the `WARN` log level, 5xx responses are logged using `ERROR` and all other responses are logged using `INFO`.
Only requests that can be successfully parsed are logged.
