package main 

import(
    "raft"
    "test/keyValStore"
    "os"
    "os/signal"
    "time"
    "strconv"
    "fmt"
    "syscall"
)

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
    addrs := []raft.ServerAddress{"127.0.0.1:8000","127.0.0.1:8001","127.0.0.1:8002"}
    keyValStore.MakeNewCluster(3, keyValStore.CreateWorkers(3), addrs, gcInterval, gcRemoveTime)
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    time.Sleep(time.Second)
    syscall.Kill(syscall.Getpid(), syscall.SIGINT)
    <-c
    <-c
}
