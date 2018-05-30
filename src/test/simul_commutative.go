package main

import (
    "raft"
    "fmt"
    "test/keyValStore"
    "time"
    "test/utils"
    "os"
)

// Test that simultaneous commutative and non-commutative operations execute without
// error.

var c1 *keyValStore.Client
var c2 *keyValStore.Client

func main() {
    trans, transErr := raft.NewTCPTransport("127.0.0.1:5000", nil, 2, time.Second, nil)
    if transErr != nil {
        fmt.Fprintf(os.Stderr, "Error with creating TCP transport, could not run tests: ", transErr)
        return
    }
    var err error
    servers := []raft.ServerAddress{"127.0.0.1:8000","127.0.0.1:8001","127.0.0.1:8002"}
    c1, err = keyValStore.CreateClient(trans, servers)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating client session, could not run tests: ", err)
        return
    }
    c2, err = keyValStore.CreateClient(trans, servers)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating second client session, could not run tests: ", err)
        return
    }

    testsFailed := utils.RunTestSuite(testSimultaneousCommutative, testSimultaneousNotCommutative)
    fmt.Println(testsFailed)
}

func testSimultaneousCommutative() (error) {
    resultCh := make(chan error, 2)
    go func() {
        resultCh <- c1.Set("foo","1")
    }()
    go func() {
        resultCh <- c2.Set("bar","1")
    }()
    err1 := <-resultCh
    err2 := <-resultCh
    if err1 != nil {
        return fmt.Errorf("Error sending simultaneous commutative requests: %v", err1)
    }
    if err2 != nil {
        return fmt.Errorf("Error sending simultaneous commutative requests: %v", err2)
    }
    result1, err3 := c1.Get("foo")
    result2, err4 := c2.Get("bar")
    if err3 != nil {
        return fmt.Errorf("Error checking result of simultaneous commutative request: %s", err3)
    }
    if err4 != nil {
        return fmt.Errorf("Error checking result of simultaneous commutative request: %s", err4)
    }
    if result1 != "1" {
        return fmt.Errorf("Error with simultaneous commutative request for foo: expected 1 but received %d", result1)
    }
    if result2 != "1" {
        return fmt.Errorf("Error with simultaneous commutative request for bar: expected 1 but received %d", result2)
    }
    return nil
}

func testSimultaneousNotCommutative() (error) {
    resultCh := make(chan error, 2)
    go func() {
        resultCh <- c1.Set("foo","1")
    }()
    go func() {
        resultCh <- c2.Set("foo","1")
    }()
    err1 := <-resultCh
    err2 := <-resultCh
    if err1 != nil {
        return fmt.Errorf("Error sending simultaneous non-commutative requests: %v", err1)
    }
    if err2 != nil {
        return fmt.Errorf("Error sending simultaneous non-commutative requests: %v", err2)
    }
    result, err3 := c1.Get("foo")
    if err3 != nil {
        return fmt.Errorf("Error checking result of simultaneous non-commutative request: %s", err3)
    }
    if result != "1" {
        return fmt.Errorf("Error in non-commutative operation: expected 1 and got %s", result)
    }
    return nil
}
