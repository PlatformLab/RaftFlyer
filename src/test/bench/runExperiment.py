import sys, string
import subprocess
import os

# IP addresses of the form 192.168.1.120:8000 where machine number is 20
def ipAddrToMachineNum(ipAddr):
    ip = ipAddr.split(":")[0]
    fourthElem = ip.split(".")[3]
    return fourthElem - 100

def runExper(config, clients, numThreads, numReqs, parallel):
    f = open(config)
    servers = f.readlines()
    # Start Raft servers
    serverProcesses = []
    for i in range(len(servers)):
        serverCmd = "ssh rc%s \"go run RaftFlyer/src/test/bench/server.go -config=config -i=%s\"" % (ipAddrToName(servers[i]), i)
        process = subprocess.Popen(serverCmd, shell=True, stdout=devNull)
        serverProcesses.append(process)
    time.sleep(0.5)     # Allow time to reach stability

    # Start Raft clients
    clientProcesses = []
    for client in clients:
        clientCmd = "ssh rc%s \"go run RaftFlyer/src/test/bench/client.go -config=config -addr=%s -comm=%d -n=%d -parallel=%s -t=%d\"" % (ipAddrToName(client), client, percentCommutative, numReqs, str(parallel), numThreads)
        process = subprocess.Popen(clientCmd, shell=True, stdout=subprocess.PIPE)
        clientProcesses.append(process)

    # Collect client measurements
    latencies = []
    totThroughput = 0
    for client in clientProcesses:
        output = client.stdout.read()
        outputLines = output.splitlines()
        latencies = latencies + outputLines[0:len(outputLines)-2]   # All lines except last line
        throughputArr = outputLines[len(outputLines)-1].split(":")
        if len(throughputArr) < 2:
            print "ERROR: cannot parse throughput %s" % outputLines[len(outputLines) - 1]
            return
        totThroughput += throughput
    avgThroughput = float(totThroughput) / len(clients)

    # Kill all raft servers
    for process in serverProcesses:
        process.terminate()

    return latencies, avgThroughput
