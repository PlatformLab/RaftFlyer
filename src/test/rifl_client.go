package main

import (
    "raft"
    "fmt"
    "test/keyValStore"
    "time"
)

var totalTests uint32
var testsFailed uint32

var c1 *keyValStore.Client
var c2 *keyValStore.Client

func main() {
    trans, transErr := raft.NewTCPTransport("127.0.0.1:5000", nil, 2, time.Second, nil)
    if transErr != nil {
        fmt.Println("Error with creating TCP transport, could not run tests: ", transErr)
        return
    }
    var err error
    servers := []raft.ServerAddress{"127.0.0.1:8000","127.0.0.1:8001","127.0.0.1:8002"}
    c1, err = keyValStore.CreateClient(trans, servers)
    if err != nil {
        fmt.Println("Error creating client session, could not run tests: ", err)
        return
    }
    c2, err = keyValStore.CreateClient(trans, servers)
    if err != nil {
        fmt.Println("Error creating second client session, could not run tests: ", err)
        return
    }
    totalTests = 0
    testsFailed = 0

    runTest(testSameClientSameSeqno)
    runTest(testSameClientDiffSeqno)
    runTest(testDiffClientSameSeqno)

    fmt.Println("***** BASIC RIFL TESTS FAILING: ", testsFailed, "/", totalTests, " *****")
}

func runTest(test func()(error)) {
    err := test()
    if err != nil {
        fmt.Println("Error running test: ", err)
        testsFailed += 1
    }
    totalTests += 1
}

func testSameClientSameSeqno() (error) {
    val1, val2, err := sendIncRpcs(c1, c1, 1234, 1234)
    if err != nil {
        return fmt.Errorf("Error sending same request from same client with same sequence number: %v", err)
    }
    if val1 != val2 {
        return fmt.Errorf("Requests from same client with same sequence number produced different results: %v, %v", val1, val2)
    }
    return nil
}

func testSameClientDiffSeqno() (error) {
    // Test same client, different sequence number
    val1, val2, err := sendIncRpcs(c1, c1, 12, 34)
    if err != nil {
        return fmt.Errorf("Error sending same request from same client with different sequence number: %v", err)
    }
    if val1 == val2 {
        return fmt.Errorf("Requests from same client with different sequence numbers produced same results: %v, %v", val1, val2)
    }
    return nil
}

func testDiffClientSameSeqno() (error) {
    val1, val2, err := sendIncRpcs(c1, c2, 123, 123)
    if err != nil {
        return fmt.Errorf("Error sending same request from different clients with same sequence number: %v", err)
    }
    if val1 == val2 {
        return fmt.Errorf("Requests from different clients with same sequence numbers produced same results: %v, %v", val1, val2)
    }
    return nil
}

func sendIncRpcs(c1 *keyValStore.Client, c2 *keyValStore.Client, seqno1 uint64, seqno2 uint64) (uint64, uint64, error) {
    val1, err1 := c1.IncWithSeqno(seqno1)
    if err1 != nil {
        return 0, 0, fmt.Errorf("Error sending RPC first time: %v", err1)
    }
    val2, err2 := c2.IncWithSeqno(seqno2)
    if err2 != nil {
        return 0, 0, fmt.Errorf("Error retransmitting RPC: %v", err2)
    }
    return val1, val2, nil
}
