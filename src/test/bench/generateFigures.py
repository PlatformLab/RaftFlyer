#import matplotlib.pyplot as plt

clients = ["192.168.1.123:8000"]

latencies, throughput = runExper("config", clients, 1, 100, False)

print "LATENCIES: "
print latencies
print "THROUGHPUT: "
