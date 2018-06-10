package main

import (
    "test/keyValStore"
    "raft"
    "fmt"
    "os"
    "flag"
    "test/utils"
    "time"
    "strconv"
)

// Arguments:
//   - config: path name of config
//   - addr: IP address
//   - comm: x/100 requests are commutative
//   - n: number of total requests
//   - t: number of threads
//   - parallel: are requests parallelized?
// Prints all latencies to stdout in microseconds, 1 per line
func main() {
    addrPtr := flag.String("addr", "127.0.0.1", "IP address and port number of client")
    configPathPtr := flag.String("config", "config", "Path to config file")
    commPercentPtr := flag.Int("comm", 100, "x/100 requests are commutative")
    nPtr := flag.Int("n", 100, "total number of requests")
    threadsPtr := flag.Int("t", 1, "total number of threads running on client")
    parallelPtr := flag.Bool("parallel", false, "true if requests are parallelized, false if serial")

    flag.Parse()

    config, err := utils.ReadConfig(*configPathPtr)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error reading config at ", *configPathPtr, ": ", err)
        return
    }
    servers := config.Servers

    threads := *threadsPtr
    n := *nPtr
    results := make(chan int64, threads * n)

    start := time.Now()
    for t := 0; t < threads; t++ {
        go runClient(t, n, *commPercentPtr, *addrPtr, servers, *parallelPtr, &results)
    }

    resultList := make([]int64, threads * n)
    for i := 0; i < threads * n; i++ {
        elem := <-results
        resultList[i] = elem
    }
    elapsed := time.Since(start)

    // Print results
    for _,result := range resultList {
        fmt.Println(result)
    }
    seconds := float64(elapsed.Seconds()) + (float64(elapsed.Nanoseconds()) / 1000000000.0)
    throughput := float64(threads * n) / seconds
    fmt.Println("THROUGHPUT: ", throughput)


}

func runClient(t int, n int, commPercent int, addr string, servers []raft.ServerAddress, parallel bool, results *chan int64) {
    port := 5000 + t
    addr = addr + ":" + strconv.Itoa(port)
    trans, err := raft.NewTCPTransport(addr, nil, 2, time.Second, nil)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating TCP transport: %s", err)
        return
    }

    client, err := keyValStore.CreateClient(trans, servers)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating client session: %s", err)
        return
    }

    for i := 0; i < n; i++ {
        if parallel {
            go makeRequest(client, i, commPercent, addr, results)
        } else {
            makeRequest(client, i, commPercent, addr, results)
        }
    }
}

func makeRequest(client *keyValStore.Client, i int, commPercent int, addr string, results *chan int64) {
    nonComm := 100 - commPercent
    // interval is -1 if no non-commutative operations
    interval := -1
    if nonComm != 0 {
        interval = 100 / nonComm
    }
    start := time.Now()
    if interval != -1 && i % interval == 0 {
        // Non-commutative operation.
        client.Set("foo", "bar")
    } else {
        // Commutative operation.
        uniqueKey := addr + string(i)
        client.Set(uniqueKey, "bar")
    }
    elapsed := time.Since(start)
    usElapsed := elapsed.Nanoseconds() / 1000
    *results <- usElapsed
}
