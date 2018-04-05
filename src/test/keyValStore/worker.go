package keyValStore

import(
    "raft"
    "encoding/json"
    "fmt"
    "io"
)

// *WorkerFSM implements raft.FSM by implementing Apply,
// Snapshot, Restore.
type WorkerFSM struct {
    // Map representing key-value store.
    KeyValMap       map[string]string
}

// Create array of worker FSMs for starting a cluster.
func CreateWorkers(n int) ([]raft.FSM) {
    workers := make([]*WorkerFSM, n)
    for i := range workers {
        workers[i] = &WorkerFSM{
            KeyValMap:  make(map[string]string),
        }
    }
    fsms := make([]raft.FSM, n)
    for i, w := range workers {
        fsms[i] = w
    }
    return fsms
}

// TODO: remove callbacks from raft
// Apply command to FSM and return response and callback
func (w *WorkerFSM) Apply(log *raft.Log)(interface{}, []func() [][]byte) {
    args := make(map[string]string)
    err := json.Unmarshal(log.Data, &args)
    if err != nil {
        fmt.Println("Poorly formatted request: ", err)
        return nil, nil
    }
    function := args[FunctionArg]
    switch function {
        case GetCommand:
            return GetResponse{Value: w.KeyValMap[args[KeyArg]]}, nil
        case SetCommand:
            w.KeyValMap[args[KeyArg]] = args[ValueArg]
            return nil, nil

    }
    return nil, []func()[][]byte{}
}

// TODO: implement for key-value store
func (w *WorkerFSM) Snapshot() (raft.FSMSnapshot, error) {
    return nil, nil
}

// TODO: implement for key-value store
func (w *WorkerFSM) Restore(i io.ReadCloser) error {
    return nil
}
