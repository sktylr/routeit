### examples/static/rewrites/static

This example shows how URL rewrites can be used when loading static assets to make the URLs appear more user friendly.
Run `go run main.go` and go to [`localhost:3000`](http://localhost:3000) in your browser.

The server's URL rewrites are in found in [`conf/rewrites.conf`](./conf/rewrites.conf).
This file shows the syntax of static URL rewrites.
The comment delimiter is `#`, and it can appear on its own or at the end of a line.
URLs must be prefixed with a leading slash and should not have a trailing slash.
The syntax is `/incoming /rewrite/to`, where the key of the config is the public facing URL, and the value is the internal URL that it should be associated with.

There are two HTML pages exposed by the server, though it responds to 4 endpoints with HTML content.
Within HTML content, the rewritten URLs can be used as targets (e.g. within JavaScript, as links, for favicons or stylesheets).

`/` and `/statics/index.html` returns the home page.

`/about` and `/statics/about.html` returns the about page, which explains how URL rewriting happens.

There are also three PNG endpoints: `/favicon.ico`, `/target.png` and `/statics/images/target.png`.

Stylesheets are returned by two endpoints: `/styles.css` and `/statics/css/styles.css`.

Lastly, there is one handler implemented within the server, which responds to `/hello/there` and `/rewrite/me/please`.
This endpoint is called within the `/about` page using a button, which demonstrates how the URL rewrites can be used on the client.
If you open the network tab when making the request, you will see it is a `GET /rewrite/me/please` request, which is internally transformed to a `GET /hello/there` request and responded to by the server.
