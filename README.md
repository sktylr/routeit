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
If non-library errors are returned (or the application code panics with an error), we attempt to infer the reason or cause and map that to a HTTP error.
The integrator can provide a custom mapper using the `ServerConfig.ErrorMapper` which can provide additional custom logic in mapping from an `error` type to an error the server understands.
If the `ErrorMapper` cannot assign a more appropriate error type, it can return `nil` which will pass off the default inference which maps common errors to sensible defaults.
For example, if an `ErrNotExist` error is returned, we map that to a 404: Not Found HTTP error.
We fallback to mapping to a 500: Internal Server Error if we cannot establish a mapping.

[`examples/errors`](/examples/errors/) contains examples for how custom error handling can be performed using `routeit`.

#### Routing

Routing is currently handled using a trie-like structure.
Typically tries are separated at the character level, but in my case I separate at the path separators (`/`), so each node contains a path segment.
Static and dynamic handling is supported.

Dynamic handling is managed by extending the values stored in the value nodes of the trie.
Dynamic components are registered to the trie with an empty key and a mark that they are dynamic.
These dynamic components match against all inputs.

Dynamic components are denoted with a leading `:`, followed by the name they should be looked up by.
The naming is case sensitive.
Currently dynamic routing only supports full string matching and does not support any regex.

```golang
"/:foo/bar/:baz": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
	foo, _ := req.PathParam("foo")
	baz, _ := req.PathParam("baz")

	// ...
	return nil
})
```

A more comprehensive example can be found in [`examples/routing/dynamic`](/examples/routing/dynamic).

To provide more control over dynamic matches, they can also optionally require a prefix and/or suffix on the match.
The syntax uses pipes (`|`) to separate the path parameter name from the prefix and suffix, and both components are optional.
`/:foo|prefix` will match against anything that has exactly 1 path component that starts with `prefix`.
It requires that at least 1 alphanumeric character (or - or \_) follows `prefix`.
So `/prefix-` and `/prefixed` will match, but `/prefix` on its own will not.

`/:foo||suffix` will match against any incoming path that has exactly 1 path component that ends with `suffix`.
`/_suffix` and `/mysuffix` will match, but `/suffix` will not.

These can be combined to a pattern like `/:foo|prefix|suffix`, which requires both a prefix and a suffix.
Again, paths like `/prefix-suffix` or `/prefixandsuffix` will match, but `/prefixsuffix` will not.
All path components in the above examples can be extracted from the request using `routeit.Request.PathParam("foo")`.
It is worth calling out that although the syntax above enforces certain prefixes or suffixes are used when matching, the entire path component is returned from `PathParam`, regardless of the prefixes or suffixes required for matching.

Server setup will panic if any of the dynamic syntax is invalid.

When traversing the trie, eligible candidates are gathered into a slice and iterated over to perform a BFS.
Eligible nodes are rejected if their children do not feature a valid node.
Once all eligible value nodes are found, they are iterated to find the one of highest priority.

Static matches have the highest priority.
Dynamic matches are judged on their specificity.

A dynamic route is more specific than another if it has strictly less dynamic components (where a dynamic component is a path segment that dynamically matches) than the other.
If the number of dynamic paths is equal, we compare for required prefixes and suffixes, by counting the total number of required prefixes and suffixes over a given dynamic route.
This is always capped by 2 \* the total number of dynamic components for a route.
If route A has strictly more prefixes and suffixes than route B, A is strictly more specific than B if they have the same number of dynamic components.
If A and B have the same number of dynamic components and prefixes and suffixes, we compare the first occurrence of a dynamic component in their path.
A is strictly more specific than B (given the same number of dynamic components and prefixes and suffixes) if A's first dynamic component appears _after_ B's first dynamic component.
This is because A has more leading static components, so therefore is more specific.
If the routes still cannot be separated, the route which was inserted first is chosen.
An example is shown in the table below.

| A                       | B                | Comparing       | More specific | Reason                                                            |
| ----------------------- | ---------------- | --------------- | ------------- | ----------------------------------------------------------------- |
| `/foo/bar`              | `/foo/:baz`      | `/foo/bar`      | A             | Static path                                                       |
| `/:foo/bar`             | `/foo/:bar`      | `/foo/bar`      | B             | Same number of dynamic components, more leading static components |
| `/foo/:bar/baz`         | `/:foo/bar/:baz` | `/foo/bar/baz`  | A             | Less dynamic components                                           |
| `/foo/:bar`             | `/foo/:bar\|baz` | `/foo/baza`     | B             | More prefixes                                                     |
| `/foo/:bar/:baz\|\|qux` | `/foo/:bar/:baz` | `/foo/bar/aqux` | A             | More suffixes                                                     |

> [!NOTE]
> Due to how prefixes and suffixes are chosen, there can be ambiguity with separate routes that match against the same path space that use prefixes and suffixes.
> For example, the routes `/foo/:bar|baz` and `/foo/:bar||qux` could both match against the same input (e.g. `/foo/bazqux`).
> Currently, `routeit` does not provide higher precedence to individual prefixes or suffixes, so these cannot be separated.
> It is best to avoid these types of matches and only use dynamic routes with prefixes and suffixes sparingly.

A specific example using dynamic components with prefixes and suffixes can be found in [`examples/routing/complex`](/examples/routing/complex/).

#### URL Rewrites

`routeit` allows the integrator to define rewrite rules for incoming URLs.
This is commonly preferred in servers that serve static content, as the raw URLs for static content is `<static directory>/<file name>.<extension>`, which is quite ugly.
For example, the convention for HTML servers is to define the home page in `static/index.html`, which means the landing page of the website would be `http://url.com/static.index.html`.

Rewrites allow us to change that to `http://url.com/`, which is much easier.
Additionally, the links used within the static content (e.g. JavaScript, HTML, stylesheets, images) can use the rewritten URL, meaning it is easier to swap components out.
If you want to change the `/about` page from `/static/about1.html` to `/static/about2.html`, you can use `/about` as the link referenced in all static content, then just change the URL rewrite rule and the change is made across the entire server.

The rewrite rules are defined using a `.conf` file and passed in using the `routeit.ServerConfig.URLRewritePath`.
Unlike the static directory definition, this can be any file on the system, though it must be accessible by the server and exist at start up.
The server will panic if the file is corrupted or badly formed and must be restarted to propagate any changes from the file.

**Static rewrites**

The syntax is a straightforward key value syntax with optional comments using `#`.
Routes should be specified on their own line using `/incoming /rewrite/to`, where the key is the URL the server receives (public facing, e.g. `/about`), and the value is the URL it should rewrite to.
The keys and values should both be prefixed with leading slashes and not have trailing slashes, and must generally by valid path components of a URL.
For example, `/foo/ //` is illegal for 3 reasons - `/foo/` ends with a slash, and `//` features an empty path component and ends with a trailing slash.

There must be at least 1 whitespace character between the key and value, though the type and amount of whitespace is up to the integrator, so long as it is not a new line.
Useless assignments (e.g. `/foo /foo`) are valid, though in practice the server will ignore them.
Conflicting rules - such as `/foo /bar` combined with `/foo /baz` - are illegal.

**Dynamic rewrites**

Dynamic rewrites are also supported by `routeit`.
They leverage the existing Trie data structure used by URL routing to also allow for dynamic rewriting.
The syntax is quite similar to that used for regex substitution.
The syntax is extended to allow the inclusion of `$`, `{` and `}` characters in both the key and value.

Within the key, the dynamic syntax can only be used in a single path component in its entirety.
So `/foo/${bar}` will allow for dynamic matching of any two-part path starting with `/foo` (e.g. `/foo/bar`, `/foo/qux`), but `/foo/${bar}baz` will only match exactly against `/foo/${bar}baz`.
The same is true for incomplete variable escaping, such as `/foo/${bar`, which will only match the exact string `/foo/${bar`.

Optional prefixes or suffixes can be specified in the key of the URL rewrite, to increase specificity.
These prefixes and suffixes are judged on their specificity in the same way that dynamic routes are managed in the router, which is covered above in [Routing](#routing).
The syntax is nearly the same - `${name|prefix|suffix}`, where `|prefix` and `|suffix` are both optional, but if you wish to match against a suffix but not a prefix, you must do `${name||suffix}`.
As with dynamic routing, the prefixes and suffixes used in dynamic URL rewrites are not stripped from the match, so if we had `/${name||.css} /css/${name}` and passed in `/styles.css`, this would be rewritten to `/css/styles.css`.
See [`examples/static/rewrites/complex`](/examples/static/rewrites/complex/) for an example of how to use these.
The same non-determinism issue mentioned above for dynamic routing can _not_ be introduced for URL rewrites, since the configuration file is read in in-order, so the order of children is always the same (unless the file is changed in some way).

Static and dynamic rules that collide (such as `/hello -> ...` and `/${name} -> ...`) are allowed, since the Trie structure will unambiguously be able to choose the static selection if it receives the exact string.
However, conflicting dynamic rules are not allowed, in the same way conflicting static rules are not.
This means that the following config would cause the server to panic.
Notice how even though the variable names are different, the set of strings they both match against are exactly the same.

```conf
/foo/${bar} 	/baz/${bar}/qux
/foo/${baz}		/bar/${baz}
```

It is not a requirement that all variables from the key are used in the value and the value may use more complex composition.
A common example (shown in [`examples/static/rewrites/dynamic`](/examples/static/rewrites/dynamic/)) is to append file extensions to incoming requests, which avoids the need to use the extension in the request.

```conf
/${page}	/assets/${page}.html
```

This will match all incoming single-part requests (excluding the root `/`), append `.html` and prepend `/assets`, resulting in the file being loaded correctly if it exists.

**Additional**

Chaining is not supported, meaning at most 1 rewrite takes place per request.
For example, assume the following rewrite rules.

```conf
/foo /bar
/bar /baz
```

If the server receives a request to `/foo`, it will redirect to `/bar`, not `/baz`, even though `/bar` redirects to `/baz`.

The actual targets (e.g. `/static/index.html`) are still accessible using their actual routes.
The server does not hide the values of the config from the public, so introducing a rewrite rule means that a resource will be available at two routes.

Examples of using URL rewrites can be found in [`examples/static/rewrites`](/examples/static/rewrites/).

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
