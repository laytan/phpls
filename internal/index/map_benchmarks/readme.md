A collection of benchmarks with different strategies of putting file content into a map.

The use was for thousands of files,
and using an sync.errgroup with 8 goroutines getting file content and putting it in the map.

Using a normal map, set to an initial high capacity
(1000 in this case, but it is a tradeoff between performance and memory usage)
with a mutex was the fastest.

Using the channel approach is the same speed as the mutexes but is more complex.

The concurrent map package follows, with the sync.Map being significantly slower
and finally the counting, initializing, parsing in sync which was very slow (waiting for the counting).

**Tried with the following:**

1. Mutex controlled map initialized with a capacity of 1000
2. Mutex controlled map
3. A channel for results, being processed by one goroutine
4. The https://github.com/orcaman/concurrent-map package
5. The sync.Map
6. Counting the files, initializing the map with that, then starting the parse/walk

Detailed results are in the txt files.

