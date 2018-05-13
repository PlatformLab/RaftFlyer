# Run all CURP tests.

if ! go build run_cluster.go
then
    echo "Cluster build failing. Cannot run tests."
    return
fi
echo "Starting tests..."
echo ""
./run_cluster > /dev/null &> /dev/null &
sleep .1
FAILED=$(go run sanity_check.go)
FAILED=$(expr $(go run rifl_client.go) + $FAILED)
CLUSTER_JOB=$(ps aux | grep "run_cluster" | grep -v grep | awk '{print $2}') &> /dev/null
kill $CLUSTER_JOB &> /dev/null
wait $CLUSTER_JOB &> /dev/null
./run_cluster 10 100 > /dev/null &> /dev/null &
sleep .1
FAILED=$(expr $(go run gc_client.go) + $FAILED)
CLUSTER_JOB=$(ps aux | grep "run_cluster" | grep -v grep | awk '{print $2}') &> /dev/null
kill $CLUSTER_JOB &> /dev/null
wait $CLUSTER_JOB &> /dev/null
go run restart_cluster.go > /dev/null &> /dev/null &
sleep .1
FAILED=$(go run restart_client.go)
CLUSTER_JOB=$(ps aux | grep "run_cluster" | grep -v grep | awk '{print $2}') &> /dev/null
kill $CLUSTER_JOB &> /dev/null
wait $CLUSTER_JOB &> /dev/null
./run_cluster > /dev/null &> /dev/null &
sleep .1
FAILED=$(expr $(go run simul_commutative.go) + $FAILED)
CLUSTER_JOB=$(ps aux | grep "run_cluster" | grep -v grep | awk '{print $2}') &> /dev/null
kill $CLUSTER_JOB &> /dev/null
wait $CLUSTER_JOB &> /dev/null
echo ""
echo "***** TESTS FAILED: "$FAILED" *****"

