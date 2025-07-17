### examples/routing/dynamic

This example server shows how dynamic routing can be performed using `routeit`.
`routeit` supports both static and dynamic routing.
Precedence is always given to static routes (i.e. matched routes where there are no dynamic components).
`routeit` also supports ambiguous dynamic routes (where an incoming URI can be resolved to multiple dynamic routes) and routes the request to the route with the higher priority.
Priority is determined using a measure of specificity.

A dynamic route is strictly more specific than another dynamic route if it features strictly less dynamic components, where a dynamic component is a path component that is matched against dynamically.
If two dynamic routes feature the same number of dynamic components, then the location of their first dynamic component is examined.
A dynamic route is strictly more specific than another dynamic route if it features more static path components before its first dynamic component than the other route.
If the routes can still not be separated, then the more recent route found in the router is chosen.

`/hello/:name`. This is the first dynamic route on the server.
The client can dynamically set the second component of the path.
The simple implementation will then extract the name and use it in the response as the subject of the greeting message.
```bash
$ curl http://localhost:8080/hello/routeit -H "Content-Type: application/json" -d '{"message": "This will hit the /hello/:name handler"}'
{"to":"routeit","from":"routeit dynamic route","message":"This will hit the /hello/:name handler"}
```

`/hello/bob`. This is the only static route on the server.
As mentioned above, it will take priority over the over routes when a `POST /hello/bob` request is received.
It simply responds with a hard-coded greeting object.
```bash
$ curl http://localhost:8080/hello/bob -H "Content-Type: application/json" -d '{"message": "This will hit the /hello/bob handler"}'
{"to":"bob","from":"routeit static route","message":"This will hit the /hello/bob handler"}
```

`/:greeting/bob`. This route is another dynamic route.
Compared to `/hello/:name`, it is less specific due to having a dynamic component earlier in the route.
It uses the `greeting` path parameter to compose the output object.
```bash
$ curl http://localhost:8080/welcome/bob -X POST
{"to":"bob","from":"routeit custom greeting route","message":"welcome"}
```
