package utils

import (
    "fmt"
    "os"
    "runtime"
    "reflect"
)

// Run a series of tests and print if they pass or fail.
// Params:
//   - tests: variable-length list of test functions to run. Expected
//            to have no arguments and return an error (nil if success
//            description of error if failure).
// Returns: number of tests failed
func RunTestSuite(tests ...func()(error)) int {
    testsFailed := 0

    for _,test := range tests {
        err := test()
        testName := runtime.FuncForPC(reflect.ValueOf(test).Pointer()).Name()
        if err != nil {
            fmt.Fprintf(os.Stderr, "%v FAILING: %v\n", testName, err)
            testsFailed += 1
        }
        fmt.Fprintf(os.Stderr, "%v passing\n", testName)
    }

    return testsFailed
}
