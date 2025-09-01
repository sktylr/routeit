# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
