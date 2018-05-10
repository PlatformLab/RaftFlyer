package main

import (
    "test/keyValStore"
    "raft"
    "fmt"
    "strings"
    "time"
    "test/utils"
)

// Sanity check to verify that client can send request and receive response.

var c *keyValStore.Client

func main() {
    trans, err := raft.NewTCPTransport("127.0.0.1:5000", nil, 2, time.Second, nil)
    if err != nil {
        fmt.Println("Error with creating TCP transport: ", err)
        return
    }
    servers := []raft.ServerAddress{"127.0.0.1:8000","127.0.0.1:8001","127.0.0.1:8002"}
    c, err = keyValStore.CreateClient(trans, servers)
    if err != nil {
        fmt.Println("Can't create client session", err)
        return
    }

    testsFailed := utils.RunTestSuite(testSanityCheck)
    fmt.Println(testsFailed)
}

func testSanityCheck() (error) {
    c.Set("foo","bar")
    str, getErr := c.Get("foo")
    if getErr != nil {
        return fmt.Errorf("Error sending Get RPC: %v", getErr)
    }
    if strings.Compare(str,"bar") != 0 {
        return fmt.Errorf("Should have received 'bar' but instead received '%v'", str)
    }
    return nil
}
