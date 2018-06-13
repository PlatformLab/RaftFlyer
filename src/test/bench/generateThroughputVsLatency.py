from runExperiment import runExper
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import numpy as np
import sys

# Firist command line argument is % commutative requests
commPercent = sys.argv[1]
averageRuns=1#5
maxThreads=2
# Add more clients with ports, iterate over by adding clients each time
clients = ["192.168.1.123:5000", "192.168.1.124:5000", "192.168.1.125:5000"]
throughputList = []
latencyList = []
for i in range(1, maxThreads):
    avgThroughput = 0
    avgLatency = 0
    for j in range(0, averageRuns):
        latencies, throughput = runExper("/home/evd/RaftFlyer/src/test/bench/config", clients, i, 100, True, int(commPercent))
        latencies = np.array(latencies, dtype=np.int32)
        avgLatency += np.mean(latencies) 
        avgThroughput += throughput
    avgThroughput /= float(averageRuns)
    avgLatency /= float(averageRuns)
    throughputList.append(avgThroughput)
    latencyList.append(avgLatency)
plt.plot(throughputList, latencyList, linestyle='-', marker='o')
plt.xlabel("Throughput (ops/sec)")
plt.ylabel("Latency (microsec)")
plt.savefig("ThroughputVsLatency.png")
