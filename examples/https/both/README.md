### examples/https/both

This example contains a server that responds to both HTTP and HTTPS messages.
It can be run using `go run main.go`.

Only 1 endpoint is exposed, which is `/hello`.
Over HTTP connections, this will return a simple `"Hello world!"` message.
When the client uses HTTPS (which has to be on a different port), we thank them for being secure in their communication.

```bash
# Plain HTTP request
$ curl http://localhost:8080/hello
Hello world!

# Secure HTTPS request
$ curl --cacert ../certs/ca.crt https://localhost:8443/hello
Hello world! Thanks for being secure!

# HTTP request to HTTPS port
# Request fails due to not having a TLS handshake
$ curl http://localhost:8443/hello -v
* Host localhost:8443 was resolved.
* IPv6: ::1
* IPv4: 127.0.0.1
*   Trying [::1]:8443...
* Connected to localhost (::1) port 8443
> GET /hello HTTP/1.1
> Host: localhost:8443
> User-Agent: curl/8.7.1
> Accept: */*
>
* Request completely sent off
* Empty reply from server
* Closing connection
curl: (52) Empty reply from server

# HTTPS request to HTTP port
# Request fails due to client rejecting the TLS handshake, since the server
# does not reply with a TLS handshake
$ curl --cacert ../certs/ca.crt https://localhost:8080/hello -v
* Host localhost:8080 was resolved.
* IPv6: ::1
* IPv4: 127.0.0.1
*   Trying [::1]:8080...
* Connected to localhost (::1) port 8080
* ALPN: curl offers h2,http/1.1
* (304) (OUT), TLS handshake, Client hello (1):
*  CAfile: ../certs/ca.crt
*  CApath: none
* LibreSSL/3.3.6: error:1404B42E:SSL routines:ST_CONNECT:tlsv1 alert protocol version
* Closing connection
curl: (35) LibreSSL/3.3.6: error:1404B42E:SSL routines:ST_CONNECT:tlsv1 alert protocol version
```
