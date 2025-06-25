### examples/routing/simple

This app contains a simple routing config to demonstrate how routing happens in the `routeit` server.
Run the application using `go run main.go`.

There are four featured endpoints, all of which are in a linear tree:
- `/`
- `/a`
- `/a/heavily/nested`
- `/a/heavily/nested/route`

Each route returns a simple `text/plain` response with a message indicating which route was invoked.

Despite being linear, there are gaps (e.g. `/a/heavily` is not a listed route).
This example shows how the routing works using the adapted trie approach and how we can be sure that a route is present.
