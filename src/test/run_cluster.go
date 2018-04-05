package main 

import(
    "raft"
    "test/server"
    "os"
    "os/signal"
)

func main() {
    addrs := []raft.ServerAddress{"127.0.0.1:8000","127.0.0.1:8001","127.0.0.1:8002"}
    server.MakeCluster(3, server.CreateWorkers(3), addrs)
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    <-c
}
