from runExperiment import runExper
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import numpy as np

clients = ["192.168.1.123:5000"]

latencies, throughput = runExper("/home/evd/RaftFlyer/src/test/bench/config",
        clients, 1, 100, False, 100)

print "LATENCIES: "
print latencies
print "THROUGHPUT: "
print throughput

latencies = np.array(latencies, dtype=np.int32)
#latencies = latencies.astype(np.int)
num_bins = 20
counts, bin_edges = np.histogram(latencies, bins=num_bins, normed=True)
cdf = np.cumsum(counts)
plt.plot(bin_edges[1:], cdf)
#plt.hist(latencies, normed=True, cumulative=True, label='CDF', histtype='step', alpha=0.8, color='k')
plt.title("CURP CDF")
plt.savefig("cdf.png")
