package keyValStore 

import(
	"fmt"
	"raft"
	"io/ioutil"
	"time"
	"log"
	"os"
)


func MakeNewCluster(n int, fsms []raft.FSM, addrs []raft.ServerAddress, gcInterval time.Duration, gcRemoveTime time.Duration) *cluster {
    return MakeCluster(n, fsms, addrs, gcInterval, gcRemoveTime, nil)
}

func RestartCluster(c *cluster) {
    for i := range c.fsms {
        err := raft.RecoverCluster(c.conf, c.fsms[i], c.stores[i], c.stores[i], c.snaps[i], c.trans[i], c.configuration)
        if err != nil {
            fmt.Println("[ERR] err: %v", err)
        }
    }
}

// Starts up a new cluster.
func MakeCluster(n int, fsms []raft.FSM, addrs []raft.ServerAddress, gcInterval time.Duration, gcRemoveTime time.Duration, startingCluster *cluster) *cluster {
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

		store := raft.NewInmemStore()
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
		c.rafts = append(c.rafts, raft)
	}

    return c
}

type cluster struct {
	dirs             []string
	stores           []*raft.InmemStore
	fsms             []raft.FSM
	snaps            []*raft.FileSnapshotStore
	trans            []raft.Transport
	rafts            []*raft.Raft
	conf             *raft.Config
	propagateTimeout time.Duration
	longstopTimeout  time.Duration
	startTime        time.Time
    configuration    raft.Configuration
}
