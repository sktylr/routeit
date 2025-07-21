## URL Rewriting

URL rewriting uses the same trie structure as routing to perform the underlying rewrite rule lookup.
The trie used is documented in [`trie.md`](./trie.md) and covers how matches are made using a measure of specificity.

### Syntax

The rewrite rules must be defined in a `.conf` file.
Comments can be added to the file using the `#` character and can either appear at the end of the line, or on their own line.
The syntax of rewrite rules is similar to the syntax or routing discussed in [`trie.md`](./trie.md).
The server will panic when starting up if the file provided for the rewrites does not exist, is corrupted, cannot be opened for another reason, or is malformed (i.e. does not follow the syntax rules laid out).
It is read once at start up time, so the server must be restarted to make any changes take effect.

#### Static Matches

Static matches are simply the path the server should receive followed by whitespace (except new line) followed by the path the server should rewrite it to.
All paths (both keys and values) must start with a leading slash and may not end with a trailing slash.
For example:

```conf
/foo/bar /baz # Rewrites "/foo/bar" to "/baz"

# The following would be illegal if uncommented
# foo/ /baz/
```

#### Dynamic Matches

Rewriting supports dynamic matches as well, including the prefixed and suffixed flavours.
The specificity is measured in the same way as for routing.
The syntax is slightly different - instead of using colons to indicate variable names, variable names are wrapped in `${<name>}`.
The matched variable can also be used as many times as needed in the rewrite value.
Below are some example of dynamic rewrite rules and what they would match

```conf
# Rewrites /foo.png to /images/foo.png
# Must be suffixed by .png
/${img||.png}		/images/${img}

# Rewrites /foo/hello to /baz/hellohello/qux
/foo/${bar}			/baz/${bar}${bar}/qux

# Rewrites /bar/prefix to /qux/prefix/bar/prefix
/bar/${foo|pre}		/qux/${foo}/bar/${foo}

# Rewrites /baz/prefix-suf to /content/prefix-suf.html
/baz/${foo|pre|suf}	/content/${foo}.html
```

Variable capture is the same as in the underlying trie with respect to prefixes and suffixes.
The match must contain the required prefixes and suffixes, but the variable capture does not strip them.

#### Conflicts

If conflicting rules are registered, the server will panic when starting up.
It will also panic in tests, to ensure the issue is as discoverable as possible.
The following are examples of illegal conflicts that would cause a panic.

```conf
# The same static match points to a different rewrite
/foo/bar /baz
/foo/bar /qux

# The same dynamic match points to separate rewrites
/foo/${bar} /baz/${bar}
/foo/${qux} /bar/${qux}/waldo

# The same dynamic match (including prefixes and suffixes) points to conflicting rewrites
/foo/${bar|pref}/${baz||suf}	/bar/${baz}/${bar}/qux
/foo/${qux|pref}/${waldo||suf}	/foo

# Reusing the same variable name in a dynamic match
/foo/${bar}/${bar}	/foo
```

### Chaining

Chaining is not supported, meaning at most 1 rewrite takes place per request.
For example, assume the following rewrite rules.

```conf
/foo /bar
/bar /baz
```

If the server receives a request to `/foo`, it will redirect to `/bar`, not `/baz`, even though `/bar` redirects to `/baz`.

The actual targets (e.g. `/static/index.html`) are still accessible using their actual routes.
The server does not hide the values of the config from the public, so introducing a rewrite rule means that a resource will be available at two routes.

### A Note on Determinism

Routing and URL rewriting theoretically suffer from the same determinism problem if overlapping entries that cannot be distinguished are entered (e.g. the following matches: `/${foo|pre}`, `/${bar||suf}` would both match against `"/presuf"` and cannot be distinguished).
Such cases are discouraged, but the rewriting algorithm will actually deterministically choose the same rewrite rule each time, regardless of how many times the server is restarted.
This is because routing registration uses map iteration, which is non-deterministic, while URL rewriting setup iterates over each line in the configuration file in order, meaning the nodes are always inserted in the same order and therefore selected in the same order.
