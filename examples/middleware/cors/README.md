### examples/middleware/cors

This example shows how `routeit`'s CORS middleware can be used to secure your server for cross-origin requests.
This example contains two servers - one for the client (a simple HTML web page served on port 3000), and one for the "API" server.
The application can be run using `go run main.go`, with the website viewed at [`localhost:3000`](http://localhost:3000).

It is best to open the network tab on your browser to understand how the CORS requests work.
Since the client (`localhost:3000`) and the server (`localhost:8080`) are on different origins, the browser will enforce CORS when required.
There are 4 routes in the server, served by 5 different handlers.

The CORS config for the server is set to allow any origins of the form `http://localhost:*`, i.e. any request from `localhost` with a non-empty port.
The server will accept CORS requests for `PUT` and `DELETE` methods, as well as the default of `GET`, `HEAD` and `POST` which are simple methods.
Outside of the CORS safe headers, the server will also allow cross origin requests that contain the headers `X-Requested-With` and `X-Custom-Header`.
Lastly, the CORS config tells the client to invalidate its cache after 15 seconds.
In a typical server, this would normally be higher, however this allows us to more easily observe the pre-flight requests that the client will make.
If an endpoint requires a pre-flight request, and a pre-flight request has already succeeded (for that endpoint) in the last 15 seconds, the client will not send another pre-flight request and will just send the request as is.
However, if more than 15 seconds have elapsed, the browser will send a pre-flight request.

We also set the exposes headers in the CORS config.
This is a list of the non CORS safe headers that can be exposed from the response to the client side JavaScript making the cross-origin request.
For this example, we introduce additional middleware that adds this response header to all requests that are forwarded through the CORS middleware.
The value of `X-Response-Header` can be seen in the client-side HTML page available on [`localhost:3000`](http://localhost:3000) when running this example.

`/simple`. This route will respond to `GET` (and therefore implicitly `HEAD`) and `POST` requests.
Crucially, the `POST` request will be made with `text/plain` `Content-Type`, which is one of the CORS safe `Content-Type` values. Neither of these requests should involve CORS from the browser's side.
The can both be triggered by the corresponding buttons in the simple HTML page.
Additionally, they can be reached through cURL, though the ability to test CORS using cURL is more limited.

```bash
# GET request
$ curl http://localhost:8080/simple
Hello from GET simple!

# GET request, allowed origin
$ curl http://localhost:8080/simple -H "Origin: http://localhost:1234"
Hello from GET simple!

# GET request, disallowed Origin
$ curl http://localhost:8080/simple -H "Origin: https://example.com"
403: Forbidden

# POST request
$ curl http://localhost:8080/simple -d "Hello from cURL" -H "Content-Type: text/plain"
Hello from POST simple with message: Hello from cURL

# POST request, allowed origin
$ curl http://localhost:8080/simple -d "Hello from cURL" -H "Content-Type: text/plain" -H "Origin: http://localhost:1"
Hello from POST simple with message: Hello from cURL

# POST request, disallowed origin
$ curl http://localhost:8080/simple -d "Hello from cURL" -H "Content-Type: text/plain" -H "Origin: http://localhost:"
403: Forbidden
```

`/update`. In the browser, this endpoint will be rejected pre-flight.
This is visible by opening the network tab and triggering the request.
The server will respond with `405: Method Not Allowed` when the browser sends the pre-flight request, due to the server's CORS config not accepting `PATCH` methods.
However, on the same origin (or allowed origin without pre-flight), we can send a request to the endpoint, which will successfully respond.

```bash
# Regular PATCH request
$ curl http://localhost:8080/update -X PATCH
Hello from PATCH /update!

# PATCH request with allowed origin
$ curl http://localhost:8080/update -X PATCH -H "Origin: http://localhost:65535"
Hello from PATCH /update!

# PATCH request from disallowed origin
$ curl http://localhost:8080/update -X PATCH -H "Origin: https://evil.com"
403: Forbidden
```

`/create`. This endpoint uses a simple method - `POST`.
However, it uses a non simple `Content-Type` header: `application/json`.
As a result, the browser will send a pre-flight request to verify the server is ready to receive a `POST` request with a `Content-Type` header.

```bash
# Regular request
$ curl http://localhost:8080/create -H "Content-Type: application/json" -d '{"name": "Name", "age": 101}'
{"message":"Hello Name (age 101), thanks for your message!"}

# Disallowed origin
$ curl http://localhost:8080/create -H "Content-Type: application/json" -d '{"name": "Name", "age": 101}' -H "Origin: foo"
403: Forbidden
```

`/remove`. Since the CORS config allows `DELETE` requests, this endpoint will be supported.
The browser will send a pre-flight request though, since `DELETE` is not a simple method.

```bash
# Allowed origin
$ curl http://localhost:8080/remove -X DELETE -H "Origin: http://localhost:8192"

# Disallowed origin
$ curl http://localhost:8080/remove -X DELETE -H "Origin: localhost"
403: Forbidden
```
