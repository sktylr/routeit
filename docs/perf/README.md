## PERF

Benchmarking is built into Go and is one of its most powerful features.
As a server, it is imperative that my code performs well (both in time and space terms).
Where I wasn't sure whether one data structure or algorithm was better than another, I tried to benchmark the approaches and use data to choose which solution to use.
This was of particular importance for some middleware, which is run on the majority of requests and where performance penalties can hamper the rest of the system.
