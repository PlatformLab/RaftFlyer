package main

import (
    "test/keyValStore"
    "raft"
    "fmt"
    "strings"
    "time"
)

func main() {
    trans, err := raft.NewTCPTransport("127.0.0.1:5000", nil, 2, time.Second, nil)
    if err != nil {
        fmt.Println("Error with creating TCP transport: ", err)
        return
    }
    servers := []raft.ServerAddress{"127.0.0.1:8000","127.0.0.1:8001","127.0.0.1:8002"}
    c, createErr := keyValStore.CreateClient(trans, servers)
    if createErr != nil {
        fmt.Println("Can't create client session", createErr)
    }
    c.Set("foo","bar")
    str, getErr := c.Get("foo")
    if getErr != nil || strings.Compare(str,"bar") != 0 {
        fmt.Println("Failure: error was ", getErr, " and returned string was ", str)
    } else {
        fmt.Println("Success!")
    }
}
