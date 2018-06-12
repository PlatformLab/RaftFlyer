package main

import (
    "raft"
    "fmt"
    "test/keyValStore"
    "time"
    "test/utils"
    "os"
)

// Tests that cached client responses are correctly stored in a snapshot and restored
// when the cluster is restarted.

var c *keyValStore.Client

func main() {
    trans, transErr := raft.NewTCPTransport("127.0.0.1:5000", nil, 2, time.Second, nil)
    if transErr != nil {
        fmt.Fprintf(os.Stderr, "Error with creating TCP transport, could not run tests: ", transErr)
        return
    }
    var err error
    servers := []raft.ServerAddress{"127.0.0.1:8000","127.0.0.1:8001","127.0.0.1:8002"}
    c, err = keyValStore.CreateClient(trans, servers)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating client session, could not run tests: ", err)
        return
    }

    testsFailed := utils.RunTestSuite(testRestartWithClientCaches)
    fmt.Println(testsFailed)
}

func testRestartWithClientCaches() (error) {
    err1 := c.Set("foo", "bar")
    if err1 != nil {
        return fmt.Errorf("Error sending RPC first time: %v", err1)
    }
    time.Sleep(2*time.Second)
    val, err2 := c.Get("foo")
    if err2 != nil {
        return fmt.Errorf("Error retransmitting RPC: %v", err2)
    }
    if val != "bar" {
        return fmt.Errorf("Didn't correctly restore value of \"bar\" after restart, isntead %s", val)
    }
    return nil
}
