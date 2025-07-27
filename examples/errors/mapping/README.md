### examples/errors/mapping

This server exposes three endpoints and demonstrates how custom error mapping can be performed and how application errors can be converted into HTTP responses.

`/invalid`. This endpoint returns a custom application error type, which is also mapped using the custom error mapper for the server.
However, the mapped value is invalid - it was not constructed using one of `routeit`'s exposes error constructors, so it falls back to a `500: Internal Server Error`.

```bash
$ curl http://localhost:8080/invalid
500: Internal Server Error
```

`/forbidden`. This endpoint returns an `fs.ErrPermission` error (normally used to indicate the application cannot load a file from disk due to permissions errors).
Since this is a known type, `routeit` automatically infers that this should map to a `403: Forbidden` error.
If the integrator wished to override this behaviour, they could do so by explicitly matching in a custom `ErrorMapper`.
However, if the default mapping is acceptable, then the integrator does not need to explicitly match against `fs.ErrPermission` at all in their custom error mapper.

```bash
$ curl http://localhost:8080/forbidden
403: Forbidden
```

`/login`. This endpoint simulates a login endpoint that can return a number of different error types.
If the provided body is missing the username or password, the server treats this as unprocessable content and returns a `422: Unprocessable Content`.
If the password does not match, we return `400: Bad Request`.
When the username or password is not provided, we return a different custom error depending on which field is missing.
However, both are rooted in the `ErrMissingInformation` type, so we need only match against those in our error mapper.

```bash
# Missing the username
$ curl http://localhost:8080/login -H "Content-Type: application/json" -d '{"password": "Password123!"}'
422: Unprocessable Content

# Missing the password
$ curl http://localhost:8080/login -H "Content-Type: application/json" -d '{"username": "user@email.com"}'
422: Unprocessable Content

# Incorrect password
$ curl http://localhost:8080/login -H "Content-Type: application/json" -d '{"username": "user@email.com", "password": "Password123"}'
400: Bad Request

# Successful login
$ curl http://localhost:8080/login -H "Content-Type: application/json" -d '{"username": "user@email.com", "password": "Password123!"}'
{"access_token":"access_123","refresh_token":"refresh_123"}
```
