### examples/https

`routeit` supports both HTTP and HTTPS servers.
This examples directory contains a number of examples for the different combinations of HTTP + HTTPS options a user of `routeit` has.

- [`both`](/examples/https/both/) - this server responds to both HTTP and HTTPS connections, and infers which type of connection was used.

### Certificates

For simplicity, some certificates and keys are committed to the repository.
These are only intended to be used as examples and are not used outside of the repository.
They are found in [`certs`](//examples/https/certs/), which contains a README outlining their creation process and purposes.
They are referred to multiple times within the README's of the respective example projects within this directory.
