### examples/https/upgrade-http

This example server showcases what happens when a server is willing to accept both HTTP and HTTPS connections, but instructs the client to upgrade their connection to HTTPS before it will be processed.
This functionality is built into `routeit` is is managed through `ServerConfig.HttpConfig.UpgradeToHttps` and `ServerConfig.HttpConfig.UpgradeInstructionMaxAge`.
The server must already be willing to listen for HTTPS connections (and therefore have a TLS config) for these settings to take effect.
The server can be run using `go run main.go`.

Only 1 endpoint is exposed, which is `/echo`.
This responds to `GET` requests and will echo the `message` query parameter, or respond with `"You didn't send a message!"` if the parameter is not present.
It also takes the first occurrence of `message` in the query parameters when parsing the request.

```bash
# HTTPS request with a message
# The Strict-Transport-Security response header is included here by the server
# to tell the client that it should continue to use HTTPS, and should remember
# that decision for at least 1 second, including on any subdomains of the host
$ curl --cacert ../certs/ca.crt --get https://localhost:8443/echo --data-urlencode "message=Hello over HTTPS" -i
HTTP/1.1 200 OK
Strict-Transport-Security: max-age=1; includeSubdomains
Content-Length: 16
Content-Type: text/plain
Date: Tue, 09 Sep 2025 21:53:45 GMT
Server: routeit

Hello over HTTPS

# HTTPS request without a message
# The Strict-Transport-Security header is also included here
$ curl --cacert ../certs/ca.crt https://localhost:8443/echo -i
HTTP/1.1 200 OK
Server: routeit
Strict-Transport-Security: max-age=1; includeSubdomains
Content-Length: 26
Content-Type: text/plain
Date: Tue, 09 Sep 2025 21:54:09 GMT

You didn't send a message!

# Plain HTTP request
# This is redirected to the correct HTTPS URI
$ curl http://localhost:8123/echo  -i
HTTP/1.1 301 Moved Permanently
Server: routeit
Location: https://localhost:8443/echo
Date: Tue, 09 Sep 2025 21:55:53 GMT

```
