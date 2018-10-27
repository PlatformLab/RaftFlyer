package main

import (
    "test/keyValStore"
    "raft"
    "fmt"
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

    testsFailed := utils.RunTestSuite(testLeaderRecovery)
    fmt.Println(testsFailed)
}

func testLeaderRecovery() (error) {
    timeout := time.After(2 * time.Second)
    tick := time.Tick(1*time.Millisecond)
    count := uint64(0)
    for {
        select {
        case <-timeout:
            //expected := uint64((2 * 1000) + 1)  // 2 seconds * # microseconds in second + 1 for last increment
            expected := count + 1
            received, err := c.Inc()
            if err != nil {
                return fmt.Errorf("Error sending increment RPC to test for leader recovery: %s", err)
            }
            if received != expected {
                return fmt.Errorf("Expected %d and received %d, error in leader recovery.", expected, received)
            }
            return nil
        case <-tick:
            _,err := c.Inc()
            if err == nil {
                count += 1
            }
        }
    }
    return nil
}
