package main

import (
    "raft"
    "fmt"
    "test/keyValStore"
    "time"
    "test/utils"
    "os"
)

// Tests correct garbage collection of client responses. Assumes that server is configured
// to garbage collect responses that have been stored for less than 1 second.

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

    testsFailed := utils.RunTestSuite(testGc)
    fmt.Println(testsFailed)
}

func testGc() (error) {
    val1, err1 := c.IncWithSeqno(1234)
    if err1 != nil {
        return fmt.Errorf("Error sending RPC first time: %v", err1)
    }
    time.Sleep(time.Second)
    val2, err2 := c.IncWithSeqno(1234)
    if err2 != nil {
        return fmt.Errorf("Error retransmitting RPC: %v", err2)
    }
    if val1 == val2 {
        return fmt.Errorf("Cached responses not correctly garbage collected.")
    }
    return nil
}
