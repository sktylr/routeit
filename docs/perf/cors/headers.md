## CORS Headers Validation

In the CORS spec, a client can send a list of headers in their pre-flight request.
These are headers they intend to use in the actual request that will follow if the pre-flight request succeeds client and server validation.
The server must inform the client whether it is willing to receive requests containing those headers or not.

In my CORS middleware, the user can include a list of headers that they will accept.
So when we get the list from the client, we must cross check it against the user's list.
The comparison must also be case insensitive.

Since this will happen for many incoming requests, it is crucial that this operation is quick, so I decided to benchmark it.

My first solution was to use the speed of map lookups and effectively use a map as a lookup table, ignoring the values stored.
This had the benefit of ignoring duplicates - if the user provided a heavily duplicated list of the headers they are willing to accept, we would ignore the duplicates which reduces the footprint compared to storing the list and just performing an O(N) lookup for each incoming header from the client.

My second solution was to build a trie (a proper one that is split per character, instead of on slashes).
Tries are great for determining whether a key is present in a set, and also avoid duplicates.
Crucially, I could also automatically build in the case insensitivity and character validation to my lookup and insert.
Out of the 128 ASCII characters, headers are permitted to use 76 of these per RFC-7230.
This can be reduced to 51 when we perform a case insensitive comparison.
The map lookup technically didn't validate these, which is also a security vulnerability.

I benchmarked both scenarios and produced the following results

```
goos: darwin
goarch: arm64
pkg: github.com/sktylr/routeit
cpu: Apple M1 Pro
                                                                    │    map.txt    │               trie.txt                │
                                                                    │    sec/op     │    sec/op     vs base                 │
CorsHeaderValidation/1_allowed_headers/match_-_first-8                 86.98n ± ∞ ¹   65.82n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1_allowed_headers/match_-_middle-8                87.16n ± ∞ ¹   63.88n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1_allowed_headers/match_-_last-8                  91.78n ± ∞ ¹   69.47n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1_allowed_headers/non_match-8                    105.30n ± ∞ ¹   52.12n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/match_-_first-8               101.00n ± ∞ ¹   63.87n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/match_-_middle-8               93.69n ± ∞ ¹   64.01n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/match_-_last-8                 85.94n ± ∞ ¹   63.85n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/non_match-8                    97.64n ± ∞ ¹   52.09n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/multi-match-8                  248.6n ± ∞ ¹   180.0n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/multi-uppercase-trimmed-8      307.6n ± ∞ ¹   205.5n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/multi-uppercase-mixed-8        201.7n ± ∞ ¹   120.8n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/multi-mixed-8                  242.9n ± ∞ ¹   111.9n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/match_-_first-8              102.90n ± ∞ ¹   64.07n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/match_-_middle-8             123.80n ± ∞ ¹   67.52n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/match_-_last-8               162.70n ± ∞ ¹   67.08n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/non_match-8                  158.30n ± ∞ ¹   52.06n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/multi-match-8                 373.4n ± ∞ ¹   181.8n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/multi-uppercase-trimmed-8     469.6n ± ∞ ¹   212.4n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/multi-uppercase-mixed-8       315.5n ± ∞ ¹   120.8n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/multi-mixed-8                 283.3n ± ∞ ¹   112.3n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/match_-_last-8              129.70n ± ∞ ¹   72.19n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/match_-_first-8             137.60n ± ∞ ¹   64.26n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/match_-_middle-8            171.50n ± ∞ ¹   72.28n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/non_match-8                 172.80n ± ∞ ¹   52.50n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/multi-match-8                377.3n ± ∞ ¹   192.1n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/multi-uppercase-trimmed-8    447.3n ± ∞ ¹   223.0n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/multi-uppercase-mixed-8      345.0n ± ∞ ¹   121.6n ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/multi-mixed-8                287.5n ± ∞ ¹   112.8n ± ∞ ¹        ~ (p=1.000 n=1) ²
geomean                                                                177.8n         91.56n        -48.50%
¹ need >= 6 samples for confidence interval at level 0.95
² need >= 4 samples to detect a difference at alpha level 0.05

                                                                    │   map.txt   │               trie.txt               │
                                                                    │    B/op     │    B/op      vs base                 │
CorsHeaderValidation/1_allowed_headers/match_-_first-8                32.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1_allowed_headers/match_-_middle-8               32.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1_allowed_headers/match_-_last-8                 32.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1_allowed_headers/non_match-8                    32.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/match_-_first-8               32.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/match_-_middle-8              32.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/match_-_last-8                32.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/non_match-8                   32.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/multi-match-8                 96.00 ± ∞ ¹   48.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/multi-uppercase-trimmed-8     96.00 ± ∞ ¹   48.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/multi-uppercase-mixed-8       64.00 ± ∞ ¹   32.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/multi-mixed-8                 64.00 ± ∞ ¹   32.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/match_-_first-8              32.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/match_-_middle-8             32.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/match_-_last-8               32.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/non_match-8                  32.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/multi-match-8                96.00 ± ∞ ¹   48.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/multi-uppercase-trimmed-8    96.00 ± ∞ ¹   48.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/multi-uppercase-mixed-8      64.00 ± ∞ ¹   32.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/multi-mixed-8                64.00 ± ∞ ¹   32.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/match_-_last-8              32.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/match_-_first-8             32.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/match_-_middle-8            32.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/non_match-8                 32.00 ± ∞ ¹   16.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/multi-match-8               96.00 ± ∞ ¹   48.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/multi-uppercase-trimmed-8   96.00 ± ∞ ¹   48.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/multi-uppercase-mixed-8     64.00 ± ∞ ¹   32.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/multi-mixed-8               64.00 ± ∞ ¹   32.00 ± ∞ ¹        ~ (p=1.000 n=1) ²
geomean                                                               46.98         23.49        -50.00%
¹ need >= 6 samples for confidence interval at level 0.95
² need >= 4 samples to detect a difference at alpha level 0.05

                                                                    │   map.txt   │               trie.txt               │
                                                                    │  allocs/op  │  allocs/op   vs base                 │
CorsHeaderValidation/1_allowed_headers/match_-_first-8                2.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1_allowed_headers/match_-_middle-8               2.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1_allowed_headers/match_-_last-8                 2.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1_allowed_headers/non_match-8                    2.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/match_-_first-8               2.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/match_-_middle-8              2.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/match_-_last-8                2.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/non_match-8                   2.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/multi-match-8                 4.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/multi-uppercase-trimmed-8     4.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/multi-uppercase-mixed-8       3.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/10_allowed_headers/multi-mixed-8                 3.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/match_-_first-8              2.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/match_-_middle-8             2.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/match_-_last-8               2.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/non_match-8                  2.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/multi-match-8                4.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/multi-uppercase-trimmed-8    4.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/multi-uppercase-mixed-8      3.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/100_allowed_headers/multi-mixed-8                3.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/match_-_last-8              2.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/match_-_first-8             2.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/match_-_middle-8            2.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/non_match-8                 2.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/multi-match-8               4.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/multi-uppercase-trimmed-8   4.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/multi-uppercase-mixed-8     3.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
CorsHeaderValidation/1000_allowed_headers/multi-mixed-8               3.000 ± ∞ ¹   1.000 ± ∞ ¹        ~ (p=1.000 n=1) ²
geomean                                                               2.531         1.000        -60.49%
¹ need >= 6 samples for confidence interval at level 0.95
² need >= 4 samples to detect a difference at alpha level 0.05
```

From this sample, the trie performed better than the map lookup - producing a result an average of 48.5% quicker, using on average half the number of bytes per operation, and allocating 60.49% less memory per operation.
As a result, I decided to continue with the trie approach.
Thankfully both solutions are quite quick - we are talking about nano seconds for the most part.
