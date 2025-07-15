### examples/middleware/simple

This example shows how middleware can be added to a `routeit` server.
The example shows a very naive authorisation middleware that simply checks the `Authorization` header of the incoming request and compares it to the string `LET ME IN`.

One endpoint is exposed that shows how the authorisation middleware is used, although it is used globally across all incoming requests, regardless of if they request a valid resource.

```bash
# Authorised request
$ curl http://localhost:8080/hello -H "Authorization: LET ME IN"
Hello authorised user!

# Unauthorised request
$ curl http://localhost:8080/hello
401: Unauthorized
```