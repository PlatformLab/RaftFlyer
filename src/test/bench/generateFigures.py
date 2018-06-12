from runExperiment import runExper
#import matplotlib.pyplot as plt

clients = ["192.168.1.123:5000"]

latencies, throughput = runExper("/home/evd/RaftFlyer/src/test/bench/config",
        clients, 1, 100, False, 100)

print "LATENCIES: "
print latencies
print "THROUGHPUT: "
print throughput
