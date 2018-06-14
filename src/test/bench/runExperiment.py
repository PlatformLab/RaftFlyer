import time
import sys, string
import subprocess
import os

# IP addresses of the form 192.168.1.120:8000 where machine number is 20
def ipAddrToMachineNum(ipAddr):
    ip = ipAddr.split(":")[0]
    fourthElem = ip.split(".")[3]
    return int(fourthElem) - 100

def getPortNum(ipAddr):
    port = ipAddr.split(":")[1].strip()
    print "port: %s" % port
    return port

def runExper(config, clients, numThreads, numReqs, parallel, percentCommutative):
    # Read config
    f = open(config, 'r')
    servers = f.readlines()
    devNull = open(os.devnull, 'w')
    # Start Raft servers
    serverProcesses = []
    for i in range(len(servers)):
        serverCmd = "ssh rc%s \"fuser -k %s/tcp; ./RaftFlyer/src/test/bench/server -config=%s -i=%s\"" % (ipAddrToMachineNum(servers[i]), getPortNum(servers[i]), config, i)
        print "SERVER CMD"
        print serverCmd
        process = subprocess.Popen(serverCmd, shell=True)
        serverProcesses.append(process)
    time.sleep(0.5)     # Allow time to reach stability

    # Start Raft clients
    clientProcesses = []
    for client in clients:
        clientCmd = "ssh rc%s \"fuser -k %s/tcp; ./RaftFlyer/src/test/bench/client -config=%s -addr=%s -comm=%d -n=%d -parallel=%s\"" % (ipAddrToMachineNum(client), getPortNum(client), config, client, percentCommutative, numReqs, str(parallel))
        print "CLIENT CMD"
        print clientCmd
        process = subprocess.Popen(clientCmd, shell=True, stdout=subprocess.PIPE)
        clientProcesses.append(process)

    # Collect client measurements
    latencies = []
    totThroughput = 0.0
    for client in clientProcesses:
        output = client.stdout.read()
        outputLines = output.splitlines()
        latencies = latencies + outputLines[0:len(outputLines)-2]   # All lines except last line
        throughputArr = outputLines[len(outputLines)-1].split(":")
        if len(throughputArr) < 2:
            print "ERROR: cannot parse throughput %s" % outputLines[len(outputLines) - 1]
            return
        throughput = float(throughputArr[1])
        totThroughput += throughput
    avgThroughput = totThroughput / float(len(clients))

    # Kill all raft servers
    for process in serverProcesses:
        process.terminate()

    return latencies, avgThroughput
