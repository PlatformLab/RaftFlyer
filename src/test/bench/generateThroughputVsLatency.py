from runExperiment import runExper
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import numpy as np
import sys

# Firist command line argument is % commutative requests
commPercent = sys.argv[1]
maxThreads = 10
# Add more clients with ports, iterate over by adding clients each time
clients = ["192.168.1.123", "192.168.1.124", "192.168.1.125"]
throughputList = []
latencyList = []
for i in range(1, maxThreads):
    tempClients = []
    for client in clients:
        for j in range(0, i):
            tempClients.append(client + ":" + str(5000 + j))
    avgThroughput = 0
    avgLatency = 0
    latencies, throughput = runExper("/home/evd/RaftFlyer/src/test/bench/config", tempClients, 100, True, int(commPercent))
    latencies = np.array(latencies, dtype=np.int32)
    throughputList.append(throughput)
    latencyList.append(np.mean(latencies))
plt.plot(throughputList, latencyList, linestyle='-', marker='o')
plt.xlabel("Throughput (ops/sec)")
plt.ylabel("Latency (microsec)")
plt.title("Throughput vs Latency with %s%% commutative requests" % commPercent)
plt.savefig("ThroughputVsLatency-%s.png" % commPercent)
