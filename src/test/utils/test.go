package utils

import (
    "fmt"
    "os"
    "runtime"
    "reflect"
)

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
