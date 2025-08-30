## examples/todo

This example ties together many of the components of `routeit` to build a fully-fledged TODO list app.
The app can be run using `go run .`.
Head to [`localhost:3000`](http://localhost:3000) to get started!
The logs for the server and client can be found in the terminal that ran `go run .`.

Here is a non-exhaustive list of the features of `routeit` that are used in this app
- URL rewrites
- Static file loading
- Host validation
- CORS middleware

There are two servers in this application - 1 for the frontend HTML, CSS and JavaScript, and 1 for the backend.
Although these could be combined this gives a more realistic representation of how many full-stack web applications are built.
This also forces us to be aware of CORS and setup the configuration correctly.

### GenAI

This app was largely vibe-coded using various GPT models from OpenAI.
I fed the LLM various pieces of documentation from the `routeit` library to assist with building individual components.
Due to the fact that `routeit` was a private library when developing this app, there were difficulties with hallucinations or cases where the LLM attempted to conform the code to mainstream libraries (such as `net/http`).
In general though it was quite successful and helped me to improve documentation and patterns in the library to make it easier to understand and build with.

### Database

The backend server uses a MySQL database to persist data.
Much of the setup is handled internally by the server on startup but creation of users is required to be done in advance.
This setup is tailored for OSX and expects that `homebrew` is installed on the machine.
This setup is only required to be run once.
`brew services start mysql` may need to be run if the server cannot establish a connection as the engine may have stopped running.

```bash
# Install MySQL if not already
# For clarity, the version used while developing is: Ver 9.4.0 for macos15.4 on arm64 (Homebrew)
brew install mysql

# Start up the engine
brew services start mysql

# (Optional) configure security
mysql_secure_installation

# Login as the root user
# Enter the configured password from above
mysql -u root -p

# Create the database within the DBMS
> create database todo_sample_app;

# Switch the context to the database we created
> use todo_sample_app;

# Create a user that the server can use to connect to the database
# For local development it is common to use a very insecure username and password combination
> CREATE USER 'todo_sample_user'@'localhost' IDENTIFIED BY 'password';

# Grant all privileges to the sample user
# This is optional, but recommended for testing purposes and safe since this is only used in a local development context
> GRANT ALL PRIVILEGES ON todo_sample_app.* TO 'todo_sample_user'@'localhost';

## Propagate the privileges to the user
> FLUSH PRIVILEGES;

# At this point it is safe to exit the MySQL console and start the server
# All going well, the server will establish a connection with the database and create the relevant tables
```

### Endpoints

There are too many endpoints to neatly describe them here.
However, the backend server has been designed in a modular manner meaning the registration of routes to the server is concise.
This can be found in [`backend.go`](./backend.go) and shows the core routes that are registered and their corresponding handlers.
The actual handlers are found in the [`handlers`](./handlers/) package.

The frontend uses static file loading and does not have any explicit handlers.
The easiest way to understand which routes are available on the client is to check [`frontend_test.go`](./frontend_test.go) which features a lean table-driven test that exercises all expected routes, including different rewritten forms.
