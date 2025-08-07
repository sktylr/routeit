## URI Rewrites

When a URI is rewritten, it is returned as a templated path, which is a list of path components.
Each of these needs to be gathered, and the final component may feature a query string which should be parsed.

The original algorithm, in pseudo code, did the following

```pseudo
rewrittenPath = []
rewrittenQuery = ""
for segment in segments:
	if segment contains '?':
		path, query = segment.Cut('?')
		rewrittenPath = rewrittenPath + path
		rewrittenQuery = query
		break
	else:
		rewrittenPath = rewrittenPath + path
```

However, when looking at this, I thought there might be opportunity to reduce the number of string iterations.
On each iteration, we loop over the string at least once, and when a query string is found, we iterate over the same string twice.
I decided to attempt to perform just 1 string iteration - cutting on the `'?'` character on each segment, and then determining whether to treat the segment as a query segment based on the presence of the query split.
The algorithm looks like below

```pseudo
rewrittenPath = []
rewrittenQuery = ""
for segment in segments:
	path, query = segment.Cut('?')
	rewrittenPath = rewrittenPath + path
	if query != "":
		rewrittenQuery = query
		break
```

However, the benchmarking results were slightly in favour of the original approach.
This is likely due to the increased cost of `strings.Cut` versus `strings.ContainsRune`, whereby (n * `strings.ContainsRune` + 1 * `strings.Cut`) is cheaper than n * `strings.Cut`.
This makes sense, since `strings.Cut` is building structures and gathering elements, compared to `strings.ContainsRune`, which is just iterating over the characters and checking for presence.
Also, the fact that `strings.ContainsRune` is just looking for a single character, versus a substring, means it is likely to perform quicker than if we were looking for a substring of length > 2.

```
goos: darwin
goarch: arm64
pkg: github.com/sktylr/routeit
cpu: Apple M1 Pro
                                     │ original.txt │              after.txt               │
                                     │    sec/op    │    sec/op     vs base                │
RewriteUri/static_no_query-8           96.23n ± ∞ ¹   98.17n ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/static_with_query-8         230.7n ± ∞ ¹   224.7n ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/dynamic_var_no_query-8      287.1n ± ∞ ¹   290.3n ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/dynamic_var_with_query-8    351.0n ± ∞ ¹   393.5n ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_1-8    49.39n ± ∞ ¹   50.47n ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_2-8    75.45n ± ∞ ¹   82.78n ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_3-8    106.1n ± ∞ ¹   110.4n ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_4-8    112.5n ± ∞ ¹   117.7n ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_5-8    147.9n ± ∞ ¹   154.3n ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_6-8    152.7n ± ∞ ¹   166.7n ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_7-8    158.6n ± ∞ ¹   168.3n ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_8-8    162.7n ± ∞ ¹   173.6n ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_9-8    213.6n ± ∞ ¹   226.2n ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_10-8   220.3n ± ∞ ¹   234.0n ± ∞ ¹       ~ (p=1.000 n=1) ²
geomean                                149.3n         156.8n        +5.05%
¹ need >= 6 samples for confidence interval at level 0.95
² need >= 4 samples to detect a difference at alpha level 0.05

                                     │ original.txt │              after.txt              │
                                     │     B/op     │    B/op      vs base                │
RewriteUri/static_no_query-8            64.00 ± ∞ ¹   64.00 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/static_with_query-8          219.0 ± ∞ ¹   217.0 ± ∞ ¹       ~ (p=1.000 n=1) ³
RewriteUri/dynamic_var_no_query-8       128.0 ± ∞ ¹   128.0 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/dynamic_var_with_query-8     240.0 ± ∞ ¹   252.0 ± ∞ ¹       ~ (p=1.000 n=1) ³
RewriteUri/long_static_segments_1-8     24.00 ± ∞ ¹   24.00 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_2-8     56.00 ± ∞ ¹   56.00 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_3-8     120.0 ± ∞ ¹   120.0 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_4-8     120.0 ± ∞ ¹   120.0 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_5-8     248.0 ± ∞ ¹   248.0 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_6-8     248.0 ± ∞ ¹   248.0 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_7-8     248.0 ± ∞ ¹   248.0 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_8-8     248.0 ± ∞ ¹   248.0 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_9-8     504.0 ± ∞ ¹   504.0 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_10-8    504.0 ± ∞ ¹   504.0 ± ∞ ¹       ~ (p=1.000 n=1) ²
geomean                                 161.2         161.7        +0.28%
¹ need >= 6 samples for confidence interval at level 0.95
² all samples are equal
³ need >= 4 samples to detect a difference at alpha level 0.05

                                     │ original.txt │              after.txt              │
                                     │  allocs/op   │  allocs/op   vs base                │
RewriteUri/static_no_query-8            4.000 ± ∞ ¹   4.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/static_with_query-8          8.000 ± ∞ ¹   8.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/dynamic_var_no_query-8       7.000 ± ∞ ¹   7.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/dynamic_var_with_query-8     10.00 ± ∞ ¹   10.00 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_1-8     2.000 ± ∞ ¹   2.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_2-8     3.000 ± ∞ ¹   3.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_3-8     4.000 ± ∞ ¹   4.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_4-8     4.000 ± ∞ ¹   4.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_5-8     5.000 ± ∞ ¹   5.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_6-8     5.000 ± ∞ ¹   5.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_7-8     5.000 ± ∞ ¹   5.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_8-8     5.000 ± ∞ ¹   5.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_9-8     6.000 ± ∞ ¹   6.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
RewriteUri/long_static_segments_10-8    6.000 ± ∞ ¹   6.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
geomean                                 4.918         4.918        +0.00%
¹ need >= 6 samples for confidence interval at level 0.95
² all samples are equal
```
