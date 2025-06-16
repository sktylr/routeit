### examples/simple

This example is a simple example that exposes 4 endpoints.
The application can be run using `go run main.go`.

`/hello`: This endpoint returns a simple JSON payload.
```bash
$ curl http://localhost:8080/hello
{"name":"John Doe","nested":{"age":25,"height":1.82}}
```

`/error`: This endpoint returns an error that is not part of the `routeit` package. Internally the `routeit` package will map this to a 500 Internal Server Error.
```bash
$ curl http://localhost:8080/error
500: Internal Server Error
```

`/crash`: This endpoint also returns an error, but uses the `InternalServerError()` function defined in the `routeit` package.
```bash
$ curl http://localhost:8080/crash
500: Internal Server Error
```

`/panic`: `routeit` will also recover any panics thrown by the application code and map them to internal server errors.
```bash
$ curl http://localhost:8080/panic
500: Internal Server Error
```
