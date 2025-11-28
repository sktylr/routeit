# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [v1.2.2] — 2025-11-28

### Changed

- Upgraded Go toolchain from `1.24.4` to `1.25.4`.

### Internal

- Migrated `stripDuplicates` to the internal helper library.
- Migrated header utilities to the internal helper library.
- Updated examples to reference `v1.2.1`.
- Bumped `golang.org/x/crypto` from `0.40.0` to `0.45.0` in `examples/todo`.

---

## [v1.2.1] — 2025-09-15

### Fixed

- Regression in `v1.2.0` where HTTP-only servers without an explicitly configured port would bind to a random ephemeral port instead of the default `8080`. HTTPS servers and explicitly configured HTTP ports were unaffected.

---

## [v1.2.0] - 2025-09-15

### Added

- **HTTPS support**

  - Servers can now be configured to accept HTTPS connections via `ServerConfig.HttpConfig`.
  - TLS is supported out-of-the-box using Go’s `crypto/tls` configuration.
  - Multiple modes are supported:
    - HTTP only (default, port `8080`).
    - HTTPS only (default port `443`).
    - Both HTTP and HTTPS simultaneously.
  - Built-in validation ensures misconfigurations (e.g. enabling HTTPS without TLS config) are caught early.

- **Automatic HTTP -> HTTPS upgrade**

  - New middleware and configuration options (`UpgradeToHttps`, `UpgradeInstructionMaxAge`) allow automatic redirection of HTTP traffic to HTTPS.
  - Responses can include `Strict-Transport-Security` headers to instruct clients to prefer HTTPS for a configurable period.

- **TLS-aware testing** _(Experimental)_

  - The `TestClient` and related utilities now support injecting TLS state for requests.
  - Enables unit and E2E testing of TLS-aware handlers and middleware.

### Changed

- **Server configuration**

  - `ServerConfig.Port` has been replaced by the more flexible `HttpConfig` struct.
  - `HttpConfig` contains ports for HTTP/HTTPS and optional TLS settings, making server setup more explicit.
  - Default behavior:
    - If no config provided -> serve HTTP on port `8080`.
    - If only HTTPS configured -> serve HTTPS on port `443`.
    - If both provided -> serve both.

### Fixed

- **Documentation**: Corrected several inaccurate docstrings.
- **Tests**: Fail faster on invalid conditions to avoid hidden panics.

### Notes

- Testing utilities remain experimental and may change in future releases.
- The replacement of `ServerConfig.Port` with `HttpConfig` is technically a breaking change. However, since there are no external consumers, this is treated as a safe, non-major release.
- `CONNECT` requests remain unsupported due to a current lack of support for tunnelling.

---

## [v1.1.0] - 2025-09-05

### Added

- **Request ID support**
  - Each incoming request can now be assigned a unique ID through a middleware.
  - Enabled by default if `ServerConfig.RequestIdProvider` is non-nil.
  - Request IDs are automatically included in logging.
  - Exposed on the request object via `Request.Id()`.
  - Response header `"X-Request-Id"` is added by default, configurable through the server.

---

## [v1.0.3] - 2025-09-01

### Fixed

- **Namespace root route registration**
  Routes registered on the root of a namespace (e.g. `/api`) are now correctly recognized and dispatched.
  - Previously, handlers registered for namespace roots (like `/api`) were ignored when serving requests.
  - This fix ensures that both namespace roots and their subroutes (e.g. `/api/items`) can be handled as expected.

---

## [v1.0.2] - 2025-09-01

### Changed

- Moved internal implementation packages (`trie`, `cmp`) into the `internal/` directory.
  - This enforces proper encapsulation and makes it clear which APIs are public and supported.
  - No changes were made to the public API surface of `routeit`.

### Notes

- Since no external consumers exist yet, this change is non-breaking.
- Future consumers should use only the public `routeit` APIs; internals are now properly hidden.

---

## [v1.0.1] - 2025-08-31

### Fixed

- **Repository structure**: Source code and `go.mod` moved to the repository root.
  This resolves issues where `go get github.com/sktylr/routeit` would not fetch the actual code.

### Note

- This is the **first usable release** of the library.
- `v1.0.0` was published with an invalid layout and should be considered deprecated.

---

## [v1.0.0] - 2025-08-31

### Added

- **HTTP/1.1 support**
  Full request parsing and response handling for all standard HTTP/1.1 methods, except `CONNECT`.

  - `TRACE` is supported but disabled by default (enable with `AllowTraceRequests`).
  - Proper `405 Method Not Allowed` responses with `Allow` headers.

- **Content type handling**

  - Automatic JSON request decoding and response encoding.
  - Native support for `text/plain`.
  - Support for arbitrary request/response content types via:
    - `Request.BodyFromRaw`
    - `ResponseWriter.RawWithContentType`

- **Routing**

  - Trie-based router with static and dynamic route matching.
  - Dynamic path parameters with prefix/suffix matching (`:id|prefix|suffix`).
  - Built-in URL rewriting support for cleaner static asset serving.

- **Error handling**

  - Automatic conversion of all returned errors and panics into HTTP responses.
  - Customizable error mapping via `ServerConfig.ErrorMapper`.
  - Per-status error handler registration (`RegisterErrorHandlers`) for observability and custom responses.

- **Middleware**

  - Middleware chaining supported.
  - Ordering is respected and deterministic.
  - **Experimental:** Middleware can be tested in isolation (`TestMiddleware`).

- **Testing utilities** _(Experimental)_

  - `TestClient`: runs E2E-like requests without opening TCP sockets.
  - `TestMiddleware` and `TestHandler`: enable unit testing of individual components.
  - All testing APIs are considered experimental and may change.

- **Logging**
  - Built-in request logging with levels:
    - `INFO` for successful responses.
    - `WARN` for 4xx responses.
    - `ERROR` for 5xx responses.
