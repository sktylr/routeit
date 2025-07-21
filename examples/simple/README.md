### examples/simple

This example is a simple example that exposes 9 endpoints.
Due to the different HTTP methods supported by `routeit`, the server will only respond to the correct HTTP method in the request.
If a route exists but it does not support the method, the server will respond with a `405: Method Not Allowed` response and include the allowed methods in the `Allow` header.
The application can be run using `go run main.go`.

`/hello`: This endpoint returns a simple JSON payload. It is hardcoded and not dependent on the request input.

```bash
$ curl http://localhost:8080/hello
{"name":"John Doe","nested":{"age":25,"height":1.82}}
```

`/echo`: This endpoint echoes the `message` query parameter.
If not present, it responds indicating as such.

```bash
$ curl http://localhost:8080/echo
Looks like you didn't want me to echo anything!
$ curl http://localhost:8080/echo?message=Hello
Received message to echo: Hello
```

`/error`: This endpoint returns an error that is not part of the `routeit` package. Internally the `routeit` package will map this to a 500 Internal Server Error.

```bash
$ curl http://localhost:8080/error
500: Internal Server Error
```

`/crash`: This endpoint also returns an error, but uses the `ErrInternalServerError()` function defined in the `routeit` package.

```bash
$ curl http://localhost:8080/crash
500: Internal Server Error
```

`/panic`: `routeit` will also recover any panics thrown by the application code and map them to internal server errors.

```bash
$ curl http://localhost:8080/panic
500: Internal Server Error
```

`/bad-status`. This endpoint shows what happens when the integrator does not provide an appropriate `HttpStatus` object when constructing the response.
The user tries to directly instantiate the `HttpStatus` struct, but cannot access the private properties that are needed for this struct to be considered valid.
As a result, the `ResponseWriter` panics, which is recovered by the server and mapped to a `500: Internal Server Error`.
Doing it this way strips the integrator of control, preventing the error from being ignored.

```bash
$ curl http://localhost:8080/bad-route
500: Internal Server Error
```

`/`: This endpoint is a `POST` endpoint that reads the input in Json and responds using the input.

```bash
$ curl http://localhost:8080/ -H "Content-Type: application/json" -d '{"name": "Foo Bar", "nested": {"age": 19, "height": 1.45}}'
{"from":{"name":"Foo Bar","nested":{"age":19,"height":1.45}},"to":{"name":"Jane Doe","nested":{"age":29,"height":1.62}}}
```

`/multi`. This endpoint is an example of a URI that responds to multiple HTTP methods.
In this case, the endpoint supports `GET` and `POST` methods (and `HEAD` implicitly due to supporting `GET`).

```bash
# GET request
$ curl http://localhost:8080/multi
{"name":"From GET","nested":{"age":100,"height":2}}

# POST request
$ curl http://localhost:8080/multi -H "Content-Type: application/json" -d '{"age": 23, "height": 1.75}'
{"name":"From POST","nested":{"age":23,"height":1.75}}
```

`/modify`. This is an example `PUT` endpoint.
For simplicity it just echoes the request body back to the client.

```bash
$ curl http://localhost:8080/modify -X PUT -H "Content-Type: text/plain" -d 'Hello /modify endpoint!'
Hello /modify endpoint!
```

`/delete`. This is an example `DELETE` endpoint.
It just uses `routeit`'s default response for a `DELETE` request, which is a status of `204: No Content`, and an empty body.

```bash
$ curl http://localhost:8080/delete -X DELETE
```
