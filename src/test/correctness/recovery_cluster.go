package main 

import(
    "raft"
    "test/keyValStore"
    "os"
    "os/signal"
    "time"
    "strconv"
    "fmt"
)

// Run cluster for 2 seconds and then kill leader.
// Optional first argument is interval at which to garbage collect entries from client response cache
// in milliseconds. Optional second argument is length of time that entries should be left in the 
// client response cache before being garbage collected (in milliseconds).
func main() {
    args := os.Args[1:]
    var gcInterval, gcRemoveTime time.Duration
    gcInterval = 0
    gcRemoveTime = 0 
    if len(args) > 0 {
        interval, err := strconv.Atoi(args[0])
        if err != nil {
            fmt.Println("GC Interval must be an integer.")
            return
        }
        gcInterval = time.Duration(interval) * time.Millisecond
    }
    if len(args) > 1 {
        removeTime, err := strconv.Atoi(args[1])
        if err != nil {
            fmt.Println("GC remove time must be an integer.")
            return
        }
        gcRemoveTime = time.Duration(removeTime) * time.Millisecond
    }
    addrs := []raft.ServerAddress{"127.0.0.1:8000","127.0.0.1:8001","127.0.0.1:8002","127.0.0.1:8003","127.0.0.1:8004"}
    cluster := keyValStore.MakeNewCluster(5, keyValStore.CreateWorkers(5), addrs, gcInterval, gcRemoveTime)
    time.Sleep(5*time.Second)
    for _,node := range cluster.Rafts {
        if node.IsLeader() {
            node.Shutdown()
            break
        }
    }
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    <-c
}
