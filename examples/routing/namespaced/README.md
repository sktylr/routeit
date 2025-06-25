### examples/routing/namespaced

`routeit` allows routes to be namespaced at both a global and local level, which can reduce the amount of code and complexity required in setting up routes.
The namespaces in both cases can have multiple slashes (`/`).

#### Global namespace

Global namespaces are applied to all routes. Common examples would be `/api`. Since they are applied on all routes, the boiler plate is reduced as `routeit` allows you to only define them once. Use the `Namespace` property on the `ServerConfig` struct.

#### Local namespaces

Local namespaces are also helpful, for example for a group of endpoints. An example would be all of the authentication related endpoints being routed under `/auth`. The `RegisterRoutesUnderNamespace` function allows registering a specific set of routes under an additional namespace. Note: the global namespace is always respected when registering the routes. So remember that if a global namespace exists, the local namespace will be **after** the global namespace in the registered route.

### Endpoints

This example server can be run using `go run main.go`.
It exposes 2 routes: `/api/hello` and `/api/namespace/hello`.
Both endpoints echo the URL they were invoked on, and interestingly use the exact same handler.
