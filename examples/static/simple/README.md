### examples/static/simple

This example shows how static file loading can be done.
Here we store assets on disk and serve them from the server using a static asset configuration.

In the example, all files are accessible under the `/static` URL namespace.

There are a number of example files provided.

As usual the server can be run using `go run main.go`.

#### HTML

There is a simple HTML website provided that features 2 HTML documents, 1 image and a stylesheet.
The website is accessible in the browser at `http://localhost:8080/static/index.html`.
The browser will automatically load the images and stylesheet and allow routing between the pages by clicking the respective buttons.

#### Plain text

There are two plaintext files served on the server: `hello.txt` and `permission-denied.txt`.
They are accessible via `curl http://localhost:8080/static/hello.txt` and `curl http://localhost:8080/static/permission-denied.txt` respectively.
To get the full experience, run `chmod 000 static/permission-denied.txt`.
This will give 1 endpoint (`/static/hello.txt`) that will return a plaintext response, and one endpoint that will resolve to a 403: Permission Denied error.
