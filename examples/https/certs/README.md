### examples/https/certs

This directory contains example certificates and keys that are used by the examples servers that communicate over HTTPS.
In order to make the examples as reproducible as possible and require minimal effort to run, I have included a local Certificate Authority (CA) which is used to sign the server certs.
This way, the client can be configured to use the local CA and will trust the server certs.

The local CA can be used in clients.
For example, in cURL we can do `curl --cacert ca.crt ...`, and in the browser we can load the CA using browser settings.

### Cert creation

```bash
# Generate the local CA cert and key
$ openssl req -x509 -new -nodes -keyout ca.key -out ca.crt -subj "/CN=Dev CA" -days 365

# Generate a key for localhost
openssl genrsa -out localhost.key 2048

# Generate a Certificate Signing Request (CSR) for localhost.
# This will be signed by the local CA so that it can be trusted.
$ openssl req -new -key localhost.key -out localhost.csr -subj "/CN=Dev CA"

# We use an extensions file "localhost.ext" which allows (amongst other things)
# for multiple IP addresses to be associated with the cert. This is useful since
# "localhost" can often be resolved to "127.0.0.1" or "::1" by some clients.

# Sign the localhost CRS with the local CA
$ openssl x509 -req -in localhost.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out localhost.crt -days 365 -sha256 -extfile localhost.ext
```
