### examples/https/https-only

This example project only accepts HTTPS requests.
Requests over HTTP will not succeed, due there not being a port that listens for plain TCP requests.
The app can be run using `go run main.go`.

There is 1 endpoint exposed, which tells the client what TLS Version and Cipher Suites it negotiated in the TLS handshake, as well as the name of the Server.

```bash
# HTTPS request
$ curl --cacert ../certs/ca.crt https://localhost:8443/info
{"version":"TLS 1.3","cipher_suite":"TLS_CHACHA20_POLY1305_SHA256","server_name":"localhost"}

# HTTP request
$ curl http://localhost:8443/info -v -i
* Host localhost:8443 was resolved.
* IPv6: ::1
* IPv4: 127.0.0.1
*   Trying [::1]:8443...
* Connected to localhost (::1) port 8443
> GET /info HTTP/1.1
> Host: localhost:8443
> User-Agent: curl/8.7.1
> Accept: */*
>
* Request completely sent off
* Empty reply from server
* Closing connection
curl: (52) Empty reply from server

# When targeting an IP directly, the server's name is not included in the TLS handshake

# HTTPS request to IPv4 host name
$ curl --cacert ../certs/ca.crt "https://127.0.0.1:8443/info"
{"version":"TLS 1.3","cipher_suite":"TLS_CHACHA20_POLY1305_SHA256","server_name":""}

# HTTPS request to IPv6 host name
$ curl --cacert ../certs/ca.crt "https://[::1]:8443/info"
{"version":"TLS 1.3","cipher_suite":"TLS_CHACHA20_POLY1305_SHA256","server_name":""}
```
