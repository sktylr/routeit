### examples/static/rewrites/dynamic

This simple application shows how dynamic URL rewrites can be used.
As with [other URL rewrite examples](/examples/static/rewrites/), this is most useful when statically serving assets.

The application can be run using `go run main.go` and accessed at [`localhost:8080`](http://localhost:8080/).

The rewrite rules can be found in [`conf/rewrites.conf`](./conf/rewrites.conf).

There are four rules, one of which is dynamic.
To understand more about static routing, check out [the corresponding example](/examples/static/rewrites/static/).

Dynamic matching uses a regex-like syntax to capture path components as variables and use them to transform the URL.
The syntax is `/static/${dynamic} /different/route/${dynamic}`.
Repeated variable names cannot appear in keys, though they may appear in values.
Additionally, the variable must comprise of the entire path component in the key, but this is not the case in the value, as shown in this example.

The dynamic rule for this server is `/${page} /assets/${page}.html`.
This rewrites all incoming requests that have a single path component to the corresponding HTML file in the [`assets`](./assets) directory, appending the HTML file extension to avoid needing to specify it in the request.

Notice how this rule technically collides with two of the static rules - `/favicon.ico` and `/styles`.
However, since these rules are both static rules, they unambiguously take precedence over the rewrite, so we allow this.
Were we to introduce another rule `/${page} /assets/${page}`, the server would not start, as this is ambiguous within the existing rule set.

There is also one custom handled endpoint implemented.
`/api/contact`. This endpoint acts like an API endpoint that the client can invoke to contact the website owners.
In reality, it does nothing of note, but it is integrated into the [`/contact`](http://localhost:8080/contact) page.
```bash
$ curl http://localhost:8080/api/contact -H "Content-Type: application/json" -d '{"name": "John Doe", "email": "hello@example.com", "message": "I am very upset with your service :("}'
Thanks for your message John Doe!
```
