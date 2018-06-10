package keyValStore 

import(
	"fmt"
	"raft"
	"io/ioutil"
	"time"
	"log"
	"os"
    "raft-boltdb"
)

// Manage keyValStore cluster locally, including making a new cluster, 
// restarting a cluster, and shutting down a cluster.

// Start a new Raft cluster locally.
// Params:
//   - n: number of servers in cluster.
//   - fsms: fsms to run on servers.
//   - addrs: addresses of servers in cluster.
//   - gcInterval: interval at which to garbage collect client responses
//   - gcRemoveTime: length of time to cache client responses before garbage collection.
// Returns: running cluster.
func MakeNewCluster(n int, fsms []raft.FSM, addrs []raft.ServerAddress, gcInterval time.Duration, gcRemoveTime time.Duration) *cluster {
    return MakeCluster(n, fsms, addrs, gcInterval, gcRemoveTime, nil)
}

// Given a cluster that has been stopped, restart it. 
// Params:
//   - c : cluster to restart.
func RestartCluster(c *cluster) {
    for i := range c.fsms {
        trans, err := raft.NewTCPTransport(string(c.trans[i].LocalAddr()), nil, 2, time.Second, nil)
        if err != nil {
            fmt.Println("[ERR] err creating transport: ", err)
        }
        c.trans[i] = trans
    }

    for i := range c.fsms {
		peerConf := c.conf
		peerConf.LocalID = c.configuration.Servers[i].ID
        peerConf.Logger = log.New(os.Stdout, string(peerConf.LocalID) + " : ", log.Lmicroseconds)

       err := raft.RecoverCluster(peerConf, c.fsms[i], c.stores[i], c.stores[i], c.snaps[i], c.trans[i], c.configuration)
        if err != nil {
            fmt.Println("[ERR] err: %v", err)
        }
        raft, err := raft.NewRaft(peerConf, c.fsms[i], c.stores[i], c.stores[i], c.snaps[i], c.trans[i])
		if err != nil {
		    fmt.Println("[ERR] NewRaft failed: %v", err)
		}

		raft.AddVoter(peerConf.LocalID, c.trans[i].LocalAddr(), 0, 0)
    }
}

// Shutdown a set of running Raft servers.
// Params:
//   - nodes: array of running Raft nodes.
func ShutdownCluster(nodes []*raft.Raft) {
    for _,node := range nodes {
        f := node.Shutdown()
        if f.Error() != nil {
            fmt.Println("Error shutting down cluster: ", f.Error())
        }
    }
}

// Starts up a new cluster.
// Params:
//   - n: number of servers in cluster.
//   - fsms: array of FSMs to run at Raft servers.
//   - addrs: addresses of Raft servers.
//   - gcInterval: interval at which to check for expired client responses.
//   - gcRemoveTime: interval at which client responses expire.
func MakeCluster(n int, fsms []raft.FSM, addrs []raft.ServerAddress, gcInterval time.Duration, gcRemoveTime time.Duration, startingCluster *cluster) (*cluster) {
    conf := raft.DefaultConfig()
    if gcInterval != 0 {
        conf.ClientResponseGcInterval = gcInterval
    }
    if gcRemoveTime != 0 {
        conf.ClientResponseGcRemoveTime = gcRemoveTime
    }
    bootstrap := true

    c := &cluster{
		conf:          conf,
		// Propagation takes a maximum of 2 heartbeat timeouts (time to
		// get a new heartbeat that would cause a commit) plus a bit.
		propagateTimeout: conf.HeartbeatTimeout*2 + conf.CommitTimeout,
		longstopTimeout:  5 * time.Second,
	}

	// Setup the stores and transports
	for i := 0; i < n; i++ {
		dir, err := ioutil.TempDir("", "raft")
		if err != nil {
			fmt.Println("[ERR] err: %v ", err)
		}

        file, err := ioutil.TempFile(dir, "log")
        if err != nil {
            fmt.Println("[ERR] err creating log: %v ", err)
        }

		store, err := raftboltdb.NewBoltStore(file.Name())
        if err != nil {
            fmt.Println("[ERR] err creating store: ", err)
        }
		c.dirs = append(c.dirs, dir)
		c.stores = append(c.stores, store)
        c.fsms = append(c.fsms, fsms[i])


	    snap, err := raft.NewFileSnapshotStore(dir, 3, nil)
		c.snaps = append(c.snaps, snap)

        trans, err := raft.NewTCPTransport(string(addrs[i]), nil, 2, time.Second, nil)
        if err != nil {
            fmt.Println("[ERR] err creating transport: ", err)
        }
        c.trans = append(c.trans, trans)
        c.configuration.Servers = append(c.configuration.Servers, raft.Server{
            Suffrage:   raft.Voter,
            ID:         raft.ServerID(fmt.Sprintf("server-%s", trans.LocalAddr())),
            Address:    addrs[i],
        })
	}

	// Create all the rafts
	c.startTime = time.Now()
	for i := 0; i < n; i++ {
		logs := c.stores[i]
		store := c.stores[i]
		snap := c.snaps[i]
		trans := c.trans[i]

		peerConf := conf
		peerConf.LocalID = c.configuration.Servers[i].ID
        peerConf.Logger = log.New(os.Stdout, string(peerConf.LocalID) + " : ", log.Lmicroseconds)

		if bootstrap {
			err := raft.BootstrapCluster(peerConf, logs, store, snap, trans, c.configuration)
			if err != nil {
				fmt.Println("[ERR] BootstrapCluster failed: %v", err)
			}
		}

		raft, err := raft.NewRaft(peerConf, c.fsms[i], logs, store, snap, trans)
		if err != nil {
		    fmt.Println("[ERR] NewRaft failed: %v", err)
		}

		raft.AddVoter(peerConf.LocalID, trans.LocalAddr(), 0, 0)
		c.Rafts = append(c.Rafts, raft)
	}

    return c
}

// Start single node
func StartNode(fsm raft.FSM, addrs []raft.ServerAddress, i int) {
    conf := raft.DefaultConfig()
    bootstrap := true

	// Setup the stores and transports
    dir, err := ioutil.TempDir("", "raft")
	if err != nil {
		fmt.Println("[ERR] err: %v ", err)
	}

    file, err := ioutil.TempFile(dir, "log")
    if err != nil {
        fmt.Println("[ERR] err creating log: %v ", err)
    }

	store, err := raftboltdb.NewBoltStore(file.Name())
    if err != nil {
        fmt.Println("[ERR] err creating store: ", err)
    }

    snap, err := raft.NewFileSnapshotStore(dir, 3, nil)

    trans, err := raft.NewTCPTransport(string(addrs[i]), nil, 2, time.Second, nil)
    if err != nil {
        fmt.Println("[ERR] err creating transport: ", err)
    }
    configuration := raft.Configuration{}

    for _,addr := range addrs {
        configuration.Servers = append(configuration.Servers, raft.Server{
            Suffrage:   raft.Voter,
            ID:         raft.ServerID(fmt.Sprintf("server-%s", addr)),
            Address:    addr,
        })
    }

	// Create all the rafts
	logs := store

	conf.LocalID = configuration.Servers[i].ID
    conf.Logger = log.New(os.Stdout, string(conf.LocalID) + " : ", log.Lmicroseconds)

	if bootstrap {
		err := raft.BootstrapCluster(conf, logs, store, snap, trans, configuration)
		if err != nil {
			fmt.Println("[ERR] BootstrapCluster failed: %v", err)
		}
	}

	raft, err := raft.NewRaft(conf, fsm, logs, store, snap, trans)
	if err != nil {
		fmt.Println("[ERR] NewRaft failed: %v", err)
	}

	raft.AddVoter(conf.LocalID, trans.LocalAddr(), 0, 0)
}


// Representation of cluster.
type cluster struct {
	dirs             []string
	stores           []*raftboltdb.BoltStore
	fsms             []raft.FSM
	snaps            []*raft.FileSnapshotStore
	trans            []raft.Transport
	Rafts            []*raft.Raft
	conf             *raft.Config
	propagateTimeout time.Duration
	longstopTimeout  time.Duration
	startTime        time.Time
    configuration    raft.Configuration
}
