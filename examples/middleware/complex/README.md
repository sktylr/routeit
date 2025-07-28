### examples/middleware/complex

This example shows how middleware can be chained together and use request context to store additional state on the request as it passes through the chain.
The example middleware first authenticates the user's credentials, and then the next piece of middleware verifies the user's permissions for the endpoint.
There are four endpoints exposed and the server can be run using `go run main.go`.

`/scoped`. This is the most tightly scoped endpoint.
It only allows authenticated "super" users to reach the endpoint.

```bash
# Authenticated super user
$ curl http://localhost:8080/scoped -H "Authorization: Bearer superuser_123"
You are authenticated and have the correct scopes: "fooscope:write", "fooscope:read" and "barscope".

# Authenticated regular user
$ curl http://localhost:8080/scoped -H "Authorization: Bearer user_123"
403: Forbidden

# Authenticated, no scopes
$ curl http://localhost:8080/scoped -H "Authorization: Bearer person_123"
403: Forbidden

# Unauthorised user
$ curl http://localhost:8080/scoped
401: Unauthorized
```

`/scopeless`. This endpoint requires authorisation, but does not limit the scopes of the authenticated user.

```bash
# Authenticated super user
$ curl http://localhost:8080/scopeless -H "Authorization: Bearer superuser_123"
You are authenticated and have the following scopes: [fooscope:write fooscope:read barscope]

# Authenticated regular user
$ curl http://localhost:8080/scopeless -H "Authorization: Bearer user_123"
You are authenticated and have the following scopes: [barscope]

# Authenticated, no scopes
$ curl http://localhost:8080/scopeless -H "Authorization: Bearer person_123"
You are authenticated and have the following scopes: []

# Unauthorised user
$ curl http://localhost:8080/scopeless
401: Unauthorized
```

`/no-auth`. This endpoint does not require any authorisation.

```bash
# Authenticated super user
$ curl http://localhost:8080/no-auth -H "Authorization: Bearer superuser_123"
You do not need to be authenticated to reach this endpoint!

# Authenticated regular user
$ curl http://localhost:8080/no-auth -H "Authorization: Bearer user_123"
You do not need to be authenticated to reach this endpoint!

# Authenticated, no scopes
$ curl http://localhost:8080/no-auth -H "Authorization: Bearer person_123"
You do not need to be authenticated to reach this endpoint!

# Unauthorised user
$ curl http://localhost:8080/no-auth
You do not need to be authenticated to reach this endpoint!
```

`/hello`. This endpoint requires authorisation.
The user must have the `barscopes` scope, but does not need "super" user access.

```bash
# Authenticated super user
$ curl http://localhost:8080/hello -H "Authorization: Bearer superuser_123"
You need to be authenticated and have "barscopes" to reach this endpoint. You have [fooscope:write fooscope:read barscope] scopes

# Authenticated regular user
$ curl http://localhost:8080/hello -H "Authorization: Bearer user_123"
You need to be authenticated and have "barscopes" to reach this endpoint. You have [barscope] scopes

# Authenticated, no scopes
$ curl http://localhost:8080/hello -H "Authorization: Bearer person_123"
403: Forbidden

# Unauthorised user
$ curl http://localhost:8080/hello
401: Unauthorized
```
