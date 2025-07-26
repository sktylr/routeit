### CORS Origin Validation

CORS Origin validation requires determining whether the requested Origin is accepted by the server.
Many simple implementations accept all origins, which greatly speeds up computation.
My middleware allows CORS to be configured in three ways: allow all, custom origin acceptance via an injected function, or a list of allowed origins, which may contain a single wildcard section per element in the list.

I was interested in the performance of my solution, particularly since many servers use CORS and the Origin validation happens for every incoming request that contains an `Origin` header.
The simple solution is to construct a regex, but instead I used a combined prefix, suffix and length comparison to ensure that the origins match.
To evaluate performance, I benchmarked my solution versus a regex solution.
I tried to make both comparisons as fair as possible, including pre-computing where possible and testing four cases for each benchmark iteration:

1. The origin is not allowed - no regex match and the element doesn't appear in the list
2. The origin is allowed, and matches against the first element listed when creating the config
3. The origin is allowed, and matches against the middle element listed when creating the config
4. The origin is allowed, and matches against the last element listed when creating the config.

The combination of these cases allowed me to understand the strengths and weaknesses of both solutions.
In particular, I expected the regex to perform the same in cases 2-4, whereas my own solution would not perform equally, since it required more loop iterations.
Lastly, I ran these for 1, 10, 100, 1000 and 10000 configured allowed origins.
Although those numbers seem excessive, particularly 10000, this gave me an understanding of how both solutions scaled.

Below are the benchmarking results comparison.
I've only included the seconds/operation metric, since the memory related operations were both 0 allocations for both scenarios due to the precomputing I did.

```
goos: darwin
goarch: arm64
pkg: github.com/sktylr/routeit
cpu: Apple M1 Pro
                                                             │  string.txt  │                     re.txt                      │
                                                             │    sec/op    │       sec/op         vs base                    │
CorsOriginValidation/exact_-_1_origins/first_match-8           3.174n ± ∞ ¹         34.470n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/exact_-_1_origins/middle_match-8          3.115n ± ∞ ¹         34.010n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/exact_-_1_origins/last_match-8            3.126n ± ∞ ¹         34.020n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/exact_-_1_origins/non-match-8             2.495n ± ∞ ¹         23.380n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_1_origins/first_match-8        8.105n ± ∞ ¹        113.000n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_1_origins/middle_match-8       8.106n ± ∞ ¹        113.000n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_1_origins/last_match-8         8.103n ± ∞ ¹        113.600n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_1_origins/non-match-8          9.041n ± ∞ ¹        175.600n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/exact_-_10_origins/first_match-8          3.115n ± ∞ ¹        187.500n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/exact_-_10_origins/middle_match-8         15.89n ± ∞ ¹         566.10n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/exact_-_10_origins/last_match-8           24.70n ± ∞ ¹         866.10n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/exact_-_10_origins/non-match-8            8.728n ± ∞ ¹       2640.000n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_10_origins/middle_match-8      46.15n ± ∞ ¹         821.60n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_10_origins/last_match-8        83.99n ± ∞ ¹        1320.00n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_10_origins/first_match-8       8.110n ± ∞ ¹        196.300n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_10_origins/non-match-8         78.80n ± ∞ ¹        4788.00n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/exact_-_100_origins/first_match-8         3.117n ± ∞ ¹      34928.000n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/exact_-_100_origins/middle_match-8        109.4n ± ∞ ¹        36275.0n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/exact_-_100_origins/last_match-8          217.0n ± ∞ ¹        39522.0n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/exact_-_100_origins/non-match-8           73.52n ± ∞ ¹       33169.00n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_100_origins/last_match-8       737.1n ± ∞ ¹        48880.0n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_100_origins/first_match-8      8.107n ± ∞ ¹      47869.000n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_100_origins/middle_match-8     386.1n ± ∞ ¹        48907.0n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_100_origins/non-match-8        723.7n ± ∞ ¹        60414.0n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/exact_-_1000_origins/first_match-8        3.121n ± ∞ ¹     479460.000n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/exact_-_1000_origins/middle_match-8       960.1n ± ∞ ¹       528216.0n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/exact_-_1000_origins/last_match-8         2.058µ ± ∞ ¹        524.991µ ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/exact_-_1000_origins/non-match-8          2.049µ ± ∞ ¹        461.388µ ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_1000_origins/first_match-8     8.106n ± ∞ ¹     667071.000n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_1000_origins/middle_match-8    3.614µ ± ∞ ¹        674.623µ ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_1000_origins/last_match-8      7.190µ ± ∞ ¹        681.543µ ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_1000_origins/non-match-8       7.174µ ± ∞ ¹        761.632µ ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/exact_-_10000_origins/first_match-8       3.119n ± ∞ ¹    8219436.000n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/exact_-_10000_origins/middle_match-8      9.457µ ± ∞ ¹       8869.708µ ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/exact_-_10000_origins/last_match-8        20.51µ ± ∞ ¹        9046.64µ ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/exact_-_10000_origins/non-match-8         7.718µ ± ∞ ¹       6132.172µ ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_10000_origins/middle_match-8   35.85µ ± ∞ ¹       13315.72µ ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_10000_origins/last_match-8     71.78µ ± ∞ ¹       12203.55µ ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_10000_origins/first_match-8    8.741n ± ∞ ¹   11431817.000n ± ∞ ¹           ~ (p=1.000 n=1) ²
CorsOriginValidation/wildcard_-_10000_origins/non-match-8      72.05µ ± ∞ ¹       18227.33µ ± ∞ ¹           ~ (p=1.000 n=1) ²
geomean                                                        112.9n                26.79µ        +23630.45%
¹ need >= 6 samples for confidence interval at level 0.95
² need >= 4 samples to detect a difference at alpha level 0.05
```

Interestingly, the regex solution was **always** slower, regardless of the number of inputs or position of the match.
On average, the regex solution was over 236 _times_ more expensive than the string solution.
As a result, I chose to continue to use the string comparison solution.

The benchmarking source code can be found in [`src/cors_test.go`](/src/cors_test.go), and the commits that performed the benchmarking are:

- `39ea713c62912dae644af62f0f81235848024519` - introduce benchmarking for strings solution
- `2edac590edc6359775dcaec3dc1920d79a371554` - benchmark regex solution
