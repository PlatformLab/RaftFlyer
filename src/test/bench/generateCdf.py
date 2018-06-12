from runExperiment import runExper
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import numpy as np
import sys

# First command line argument is % commutative requests
commPercent = sys.argv[1]

clients = ["192.168.1.123:5000"]

latencies, throughput = runExper("/home/evd/RaftFlyer/src/test/bench/config",
        clients, 1, 100, False, int(commPercent))

print "LATENCIES: "
print latencies
print "THROUGHPUT: "
print throughput

latencies = np.array(latencies, dtype=np.int32)

sorted_data = np.sort(latencies)
yvals=np.arange(len(sorted_data))/float(len(sorted_data)-1)
plt.plot(sorted_data, yvals)

plt.xlabel("Latency (microseconds)")
plt.title("CDF of Latency with %s%% commutative requests" % commPercent)
plt.savefig("cdf-%s.png" % commPercent)
