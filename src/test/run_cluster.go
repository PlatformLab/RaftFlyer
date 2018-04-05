package main 

import(
    "raft"
    "test/keyValStore"
    "os"
    "os/signal"
)

func main() {
    addrs := []raft.ServerAddress{"127.0.0.1:8000","127.0.0.1:8001","127.0.0.1:8002"}
    keyValStore.MakeCluster(3, keyValStore.CreateWorkers(3), addrs)
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    <-c
}
