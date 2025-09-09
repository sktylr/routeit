## Examples

This directory contains multiple examples of how different features of `routeit` can be used to create servers.
Each example is its own self-contained Go project that has its own main function and tests.
As well as allowing me to test my implementations, they also allowed me to experiment with the interfaces I was exposing to ensure they were straightforward to use.

The examples are categorised, with each category potentially having more subcategories.
- [`errors`](/examples/errors/) - examples of how errors can be handled and mapped within `routeit`.
- [`https`](/examples/https) - examples of different HTTP and HTTPS configurations
- [`middleware`](/examples/middleware/) - examples of installing middleware to the server and using it to intercept requests
- [`routing`](/examples/routing/) - explorations of `routeit`'s routing system, including static and dynamic URI lookup
- [`simple`](/examples/simple/) - a simple server that exposes at least 1 endpoint for all the HTTP methods `routeit` supports. This was my base drawing board for implementation and testing
- [`static`](/examples/static/) - `routeit` supports static file serving and this examples directory showcases how this can be done, and how it can be made easier using other features such as URL rewriting
- [`todo`](/examples/todo/) - this is the most complex and fully-fledged application. It contains a sample TODO list app and features authentication and database connections as well as full CRUD operations on multiple resources.
