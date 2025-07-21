### examples/errors/handling

This examples showcases how custom handlers can be used to respond to 4xx and 5xx status codes.
There are 7 registered routes on the server, and 3 of the 4xx and 5xx status codes have custom error handling.
These handlers are all called after the request has completed its full cycle (i.e. passed through middleware, routing, explicit handling).
They are registered for `401`, `404` and `500` status codes.

The example can be run using `go run main.go`.

`/no-auth`. This route serves both `GET` and `POST` requests, but always returns `ErrUnauthorized()`, which will map to a `401: Unauthorized` status.
The server has a handler for `401: Unauthorized` responses, so we can transform the response here.
We perform a simple transformation that uses a JSON error detail response on both endpoints (`GET` and `POST`).
Technically, we could split this up as we have access to the request, so we could have different logic depending on the request method, or path, or anything that is accessible from the request.
However, too much granularity is discouraged.

```bash
# GET request
$ curl http://localhost:8080/no-auth
{"error":{"message":"Provide a valid access token","code":"unauthorised"}}

# POST request
$ curl http://localhost:8080/no-auth -X POST
{"error":{"message":"Provide a valid access token","code":"unauthorised"}}
```

`/crash`. This endpoint returns a response with a 500 status code, which has a registered handler.
We directly return `routeit`'s `ErrInternalServerError()`, but attach a cause, which is used to compose the response.

```bash
$ curl http://localhost:8080/crash
{"error":{"message":"An internal error has occurred. We are aware and are investigating. Please try again later or reach out support if it persists. uh oh we crashed","code":"internal_server_error"}}
```

`/panic`. This endpoint panics, which will be mapped into a `500 Internal Server Error`.
Again, it goes through the custom handler.

```bash
$ curl http://localhost:8080/panic
{"error":{"message":"An internal error has occurred. We are aware and are investigating. Please try again later or reach out support if it persists. oops","code":"internal_server_error"}}
```

`/custom-error`. This endpoint returns a custom error.
When `routeit` receives this error, it attempts to infer the corresponding `HttpError`, falling back to `ErrInternalServerError()` if it cannot, which is what is done here.

```bash
$ curl http://localhost:8080/custom-error
{"error":{"message":"An internal error has occurred. We are aware and are investigating. Please try again later or reach out support if it persists. this custom error will be mapped to a 500: Internal Server Error","code":"internal_server_error"}}
```

`/manual-status`. Here, we manually set the status of the response to `500`.
This also goes through the custom handling for `500: Internal Server Error`.

```bash
$ curl http://localhost:8080/manual-status
{"error":{"message":"An internal error has occurred. We are aware and are investigating. Please try again later or reach out support if it persists.","code":"internal_server_error"}}
```

`/bad-request`. This endpoint always returns `400: Bad Request`, which does not have a corresponding error handler.
As a result, the default `routeit` response is used.

```bash
$ curl http://localhost:8080/bad-request
400: Bad Request
```

`/not-found`. Although this is actually a valid route, it will always return `404: Not Found`, via a panic of `fs.ErrNotExist`, which is mapped to a `404`.
The server has a handler for `404` status codes, so this will be passed through it.

```bash
$ curl http://localhost:8080/not-found
{"error":{"message":"No matching route found for /not-found","code":"not_found"}}
```
