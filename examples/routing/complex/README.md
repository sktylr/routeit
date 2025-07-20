### examples/routing/complex

This example demonstrates how prefixes and suffixes can be used within the routing strategy of the server to provide finer grained control over the routes that are matched.

The server can be run using `go run main.go` and registers 4 handlers that respond to `GET`, `HEAD` and `OPTIONS` request methods.
However, technically the server only responds to 1 endpoint, since they all match against the same endpoint.
Each handler returns information in the response to understand which route the request was matched against.

`/:path`. This is the most general endpoint and matches against all incoming requests with a path like `/<anything>`.
However, it is the least specific, so is only used to process the request if no other route matches it.

```bash
$ curl http://localhost:8080/hello
{"incoming_url":"/hello","handler_route":"/:path","path_param":"hello"}
```

`/:path|prefix`. This is one of two routes with equal specificity on the server.
It will match against incoming requests that have a path of the shape `/prefix<anything>`, but may not necessarily take precedence over other routes (i.e. the two below).
This is one of the drawbacks of using prefixes and suffixes in otherwise identical routes.

```bash
$ curl http://localhost:8080/prefix-route
{"incoming_url":"/prefix-route","handler_route":"/:path|prefix","path_param":"prefix-route"}
```

`/:path||suffix`. This is similar to `/:path|prefix`, except it requires a suffix of `suffix` on the path (so accepts paths like `/<anything>suffix`).
It takes direct precedence over `/:path`, but cannot be unambiguously separated from `/:path|prefix` if a pattern than matches both is provided, such as `/prefixsuffix`.
In fact, `/prefixsuffix` is the only incoming path that would non-deterministically match (i.e. it would depend on insertion order, which is made non-deterministic by go due to map iteration).
Any path with any alphanumeric character (or / or \_) between `prefix` and `suffix` would match against `/:path|prefix|suffix` below unambiguously.
This shows how being too clever with route setup can cause issues that are harder to catch.

```bash
$ curl http://localhost:8080/route-suffix
{"incoming_url":"/route-suffix","handler_route":"/:path||suffix","path_param":"route-suffix"}
```

`/:path|prefix|suffix`. This route is the most specific and will unambiguously take precedence over the other routes if it matches the path.
It requires a path of the shape `/prefix<anything>suffix`.

```bash
$ curl http://localhost:8080/prefix-suffix
{"incoming_url":"/prefix-suffix","handler_route":"/:path|prefix|suffix","path_param":"prefix-suffix"}
```

**Non-determinism**

As mentioned above, the `/:path|prefix` and `/:path||suffix` routes will introduce non-determinism in handler selection.
The non-determinism is only at server start up, i.e. once the server has started, its "choice" of handler will remain over its lifetime until it is shut down and restarted, at which point it may change.
This is due to a feature of Golang that means map iteration does not guarantee order.
Here's a [blog](https://go.dev/blog/maps#iteration-order) and the [Go 1 release notes](https://go.dev/doc/go1#iteration) explaining.
The Go language developers chose this to reduce brittle tests by forcing developers to change the paradigm used when comparing map values.

Here's an example of how the server would handle a request to `/prefixsuffix` depending on which route was registered to the router first (and will therefore be chosen) when the server was stared.
Also, a reminder that in the current server setup, `/prefixsuffix` is the **only** route that introduces such non-determinism.
Every other incoming request will be deterministically routed.

```bash
# /:path|prefix inserted first
curl http://localhost:8080/prefixsuffix
{"incoming_url":"/prefixsuffix","handler_route":"/:path|prefix","path_param":"prefixsuffix"}

# /:path||suffix inserted first
curl http://localhost:8080/prefixsuffix
{"incoming_url":"/prefixsuffix","handler_route":"/:path||suffix","path_param":"prefixsuffix"}
```
