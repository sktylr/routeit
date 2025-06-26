### examples/static

This example shows how static file loading can be done.
Here we store assets on disk and serve them from the server using a static asset configuration.

In the example, all files are accessible under the `/static` URL namespace.

To get the full experience, run `chmod 000 static/permission-denied.txt`.
This will give 1 endpoint (`/static/hello.txt`) that will return a plaintext response, and one endpoint that will resolve to a 403: Permission Denied error.

As usual the server can be run using `go run main.go`.
