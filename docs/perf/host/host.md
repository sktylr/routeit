## Host Header Validation

Per HTTP/1.1, all requests **must** contain the `Host` header.
The server is expected to validate this header as well, and confirm that it is one of the host values it will accept requests for.

`routeit` installs middleware that handles every single well-formed request and validates the host header per an allowlist defined by the user.
This allowlist can contain exact matches, or dynamic matches that allow for a (single-layered) subdomain to also be included.
For example, if the server will accept requests whose host conform to `.example.com`, this means that `api.example.com` and `example.com` are valid, but `site.web.example.com` is not.

My initial solution pre-compiled a regex and used it to match in incoming requests.
I wanted to see how that compared against another approach, so I benchmarked the performance.

The next solution was to use the same dynamic matching approach used for CORS origin validation, which is a similar problem.

The benchmark results are below.
Unsurprisingly (given the CORS results), the strings comparison approach is more performant, even when dealing with a large number of hosts to validate against.
In terms of duration per operation, the string comparison was nearly 40% faster in the samples, and used ~64% less bytes while making the same number of allocations per operation.

I may revisit these approaches to benchmark against a trie-like approach as well, as this should reduce the number of comparisons needed.

```
goos: darwin
goarch: arm64
pkg: github.com/sktylr/routeit
cpu: Apple M1 Pro
                                                                                          │  host-re.txt  │            host-string.txt             │
                                                                                          │    sec/op     │    sec/op      vs base                 │
HostValidationMiddleware/1_allowed_hosts/exact_-_first-8                                     404.2n ± ∞ ¹    220.5n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/exact_-_middle-8                                    387.2n ± ∞ ¹    216.2n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/exact_-_last-8                                      382.1n ± ∞ ¹    206.0n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/exact_miss-8                                        378.6n ± ∞ ¹    362.6n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/subdomain_-_first-8                                 553.5n ± ∞ ¹    231.7n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/subdomain_-_middle-8                                546.1n ± ∞ ¹    231.8n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/subdomain_-_last-8                                  559.7n ± ∞ ¹    227.7n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/subdomain_-_miss-8                                  699.1n ± ∞ ¹    381.8n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/subdomain_-_miss,_too_many_subdomain_levels-8       601.6n ± ∞ ¹    390.0n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/duplicate_-_exact-8                                 389.7n ± ∞ ¹    205.7n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/duplicate_-_subdomain-8                             556.9n ± ∞ ¹    233.5n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/exact_-_first-8                                    510.7n ± ∞ ¹    207.4n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/exact_-_middle-8                                   454.8n ± ∞ ¹    221.8n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/exact_-_last-8                                     414.4n ± ∞ ¹    237.3n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/exact_miss-8                                       364.7n ± ∞ ¹    396.8n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/subdomain_-_first-8                                579.2n ± ∞ ¹    236.7n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/subdomain_-_middle-8                               606.5n ± ∞ ¹    272.7n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/subdomain_-_last-8                                 642.0n ± ∞ ¹    310.5n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/subdomain_-_miss-8                                 730.7n ± ∞ ¹    471.9n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/subdomain_-_miss,_too_many_subdomain_levels-8      613.6n ± ∞ ¹    497.5n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/duplicate_-_exact-8                                516.1n ± ∞ ¹    207.7n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/duplicate_-_subdomain-8                            586.6n ± ∞ ¹    230.2n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/exact_-_first-8                                   931.0n ± ∞ ¹    207.6n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/exact_-_middle-8                                 1087.0n ± ∞ ¹    374.4n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/exact_-_last-8                                   1082.0n ± ∞ ¹    442.3n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/exact_miss-8                                      363.6n ± ∞ ¹    463.4n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/subdomain_-_first-8                              1079.0n ± ∞ ¹    228.9n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/subdomain_-_middle-8                             1237.0n ± ∞ ¹    570.6n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/subdomain_-_last-8                               1229.0n ± ∞ ¹    880.1n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/subdomain_-_miss-8                                878.0n ± ∞ ¹   1072.0n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/subdomain_-_miss,_too_many_subdomain_levels-8     702.1n ± ∞ ¹   1048.0n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/duplicate_-_exact-8                               935.1n ± ∞ ¹    204.7n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/duplicate_-_subdomain-8                          1092.0n ± ∞ ¹    228.2n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/exact_-_first-8                                 1184.0n ± ∞ ¹    219.3n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/exact_-_middle-8                                 1.705µ ± ∞ ¹    1.304µ ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/exact_-_last-8                                   1.695µ ± ∞ ¹    2.560µ ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/exact_miss-8                                     385.6n ± ∞ ¹   1166.0n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/subdomain_-_first-8                             1378.0n ± ∞ ¹    248.4n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/subdomain_-_middle-8                             1.925µ ± ∞ ¹    3.508µ ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/subdomain_-_last-8                               1.933µ ± ∞ ¹    6.775µ ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/subdomain_-_miss-8                               1.013µ ± ∞ ¹    6.968µ ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/subdomain_-_miss,_too_many_subdomain_levels-8    852.7n ± ∞ ¹   7059.0n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/duplicate_-_exact-8                             1244.0n ± ∞ ¹    208.7n ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/duplicate_-_subdomain-8                         1340.0n ± ∞ ¹    239.7n ± ∞ ¹        ~ (p=1.000 n=1) ²
geomean                                                                                      737.3n          445.7n        -39.55%
¹ need >= 6 samples for confidence interval at level 0.95
² need >= 4 samples to detect a difference at alpha level 0.05

                                                                                          │ host-re.txt  │           host-string.txt            │
                                                                                          │     B/op     │    B/op      vs base                 │
HostValidationMiddleware/1_allowed_hosts/exact_-_first-8                                     16.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/exact_-_middle-8                                    16.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/exact_-_last-8                                      16.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/exact_miss-8                                        128.0 ± ∞ ¹   128.0 ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/subdomain_-_first-8                                 28.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/1_allowed_hosts/subdomain_-_middle-8                                21.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/1_allowed_hosts/subdomain_-_last-8                                  23.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/1_allowed_hosts/subdomain_-_miss-8                                  133.0 ± ∞ ¹   128.0 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/1_allowed_hosts/subdomain_-_miss,_too_many_subdomain_levels-8       136.0 ± ∞ ¹   128.0 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/1_allowed_hosts/duplicate_-_exact-8                                 16.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/duplicate_-_subdomain-8                             27.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/10_allowed_hosts/exact_-_first-8                                    16.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/exact_-_middle-8                                   16.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/exact_-_last-8                                     16.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/exact_miss-8                                       128.0 ± ∞ ¹   128.0 ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/subdomain_-_first-8                                20.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/10_allowed_hosts/subdomain_-_middle-8                               25.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/10_allowed_hosts/subdomain_-_last-8                                 28.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/10_allowed_hosts/subdomain_-_miss-8                                 137.0 ± ∞ ¹   128.0 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/10_allowed_hosts/subdomain_-_miss,_too_many_subdomain_levels-8      134.0 ± ∞ ¹   128.0 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/10_allowed_hosts/duplicate_-_exact-8                                16.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/duplicate_-_subdomain-8                            27.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/100_allowed_hosts/exact_-_first-8                                   56.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/100_allowed_hosts/exact_-_middle-8                                  50.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/100_allowed_hosts/exact_-_last-8                                    50.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/100_allowed_hosts/exact_miss-8                                      128.0 ± ∞ ¹   128.0 ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/subdomain_-_first-8                               43.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/100_allowed_hosts/subdomain_-_middle-8                              35.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/100_allowed_hosts/subdomain_-_last-8                                53.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/100_allowed_hosts/subdomain_-_miss-8                                159.0 ± ∞ ¹   128.0 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/100_allowed_hosts/subdomain_-_miss,_too_many_subdomain_levels-8     156.0 ± ∞ ¹   128.0 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/100_allowed_hosts/duplicate_-_exact-8                               49.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/100_allowed_hosts/duplicate_-_subdomain-8                           33.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/1000_allowed_hosts/exact_-_first-8                                 759.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/1000_allowed_hosts/exact_-_middle-8                                645.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/1000_allowed_hosts/exact_-_last-8                                  508.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/1000_allowed_hosts/exact_miss-8                                     128.0 ± ∞ ¹   128.0 ± ∞ ¹        ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/subdomain_-_first-8                             540.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/1000_allowed_hosts/subdomain_-_middle-8                            853.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/1000_allowed_hosts/subdomain_-_last-8                              649.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/1000_allowed_hosts/subdomain_-_miss-8                               838.0 ± ∞ ¹   128.0 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/1000_allowed_hosts/subdomain_-_miss,_too_many_subdomain_levels-8    681.0 ± ∞ ¹   128.0 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/1000_allowed_hosts/duplicate_-_exact-8                             434.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
HostValidationMiddleware/1000_allowed_hosts/duplicate_-_subdomain-8                         408.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ³
geomean                                                                                      77.99         28.21        -63.83%
¹ need >= 6 samples for confidence interval at level 0.95
² all samples are equal
³ need >= 4 samples to detect a difference at alpha level 0.05

                                                                                          │ host-re.txt │           host-string.txt           │
                                                                                          │  allocs/op  │  allocs/op   vs base                │
HostValidationMiddleware/1_allowed_hosts/exact_-_first-8                                    1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/exact_-_middle-8                                   1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/exact_-_last-8                                     1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/exact_miss-8                                       3.000 ± ∞ ¹   3.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/subdomain_-_first-8                                1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/subdomain_-_middle-8                               1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/subdomain_-_last-8                                 1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/subdomain_-_miss-8                                 3.000 ± ∞ ¹   3.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/subdomain_-_miss,_too_many_subdomain_levels-8      3.000 ± ∞ ¹   3.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/duplicate_-_exact-8                                1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1_allowed_hosts/duplicate_-_subdomain-8                            1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/exact_-_first-8                                   1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/exact_-_middle-8                                  1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/exact_-_last-8                                    1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/exact_miss-8                                      3.000 ± ∞ ¹   3.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/subdomain_-_first-8                               1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/subdomain_-_middle-8                              1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/subdomain_-_last-8                                1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/subdomain_-_miss-8                                3.000 ± ∞ ¹   3.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/subdomain_-_miss,_too_many_subdomain_levels-8     3.000 ± ∞ ¹   3.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/duplicate_-_exact-8                               1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/10_allowed_hosts/duplicate_-_subdomain-8                           1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/exact_-_first-8                                  1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/exact_-_middle-8                                 1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/exact_-_last-8                                   1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/exact_miss-8                                     3.000 ± ∞ ¹   3.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/subdomain_-_first-8                              1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/subdomain_-_middle-8                             1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/subdomain_-_last-8                               1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/subdomain_-_miss-8                               3.000 ± ∞ ¹   3.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/subdomain_-_miss,_too_many_subdomain_levels-8    3.000 ± ∞ ¹   3.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/duplicate_-_exact-8                              1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/100_allowed_hosts/duplicate_-_subdomain-8                          1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/exact_-_first-8                                 1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/exact_-_middle-8                                1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/exact_-_last-8                                  1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/exact_miss-8                                    3.000 ± ∞ ¹   3.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/subdomain_-_first-8                             1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/subdomain_-_middle-8                            1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/subdomain_-_last-8                              1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/subdomain_-_miss-8                              3.000 ± ∞ ¹   3.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/subdomain_-_miss,_too_many_subdomain_levels-8   3.000 ± ∞ ¹   3.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/duplicate_-_exact-8                             1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
HostValidationMiddleware/1000_allowed_hosts/duplicate_-_subdomain-8                         1.000 ± ∞ ¹   1.000 ± ∞ ¹       ~ (p=1.000 n=1) ²
geomean                                                                                     1.349         1.349        +0.00%
¹ need >= 6 samples for confidence interval at level 0.95
² all samples are equal
```
