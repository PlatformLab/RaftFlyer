go run run_cluster.go > /dev/null &> /dev/null &
sleep .1
FAILED=$(go run rifl_client.go)
kill $(ps aux | grep "run_cluster" | grep -v grep | awk '{print $2}') &> /dev/null
go run run_cluster.go > /dev/null &> /dev/null &
sleep .1
FAILED=$(expr $(go run gc_client.go 100 10) + $FAILED)
kill $(ps aux | grep "run_cluster" | grep -v grep | awk '{print $2}') &> /dev/null
echo "***** TESTS FAILED: "$FAILED" *****"

