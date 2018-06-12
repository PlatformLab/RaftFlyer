package keyValStore

import(
    "raft"
    "encoding/json"
    "fmt"
    "io"
)

// FSM running on Raft servers to implement key-val store.
// *WorkerFSM implements raft.FSM by implementing Apply,
// Snapshot, Restore.
type WorkerFSM struct {
    // Map representing key-value store.
    KeyValMap       map[string]string
    counter         uint64
}

type WorkerSnapshot struct{}

// Create array of worker FSMs for starting a cluster.
// Params:
//   - n: number of workers to create.
// Returns: array of raft FSMs of length n.
func CreateWorkers(n int) ([]raft.FSM) {
    workers := make([]*WorkerFSM, n)
    for i := range workers {
        workers[i] = &WorkerFSM{
            KeyValMap:  make(map[string]string),
            counter:    0,
        }
    }
    fsms := make([]raft.FSM, n)
    for i, w := range workers {
        fsms[i] = w
    }
    return fsms
}

// Apply command to FSM and return response.
// Params:
//   - log: log entry to apply to FSM.
// Returns: response JSON object.
func (w *WorkerFSM) Apply(log *raft.Log)(interface{}) {
    args := make(map[string]string)
    err := json.Unmarshal(log.Data, &args)
    if err != nil {
        fmt.Println("Poorly formatted request: ", err)
        return nil
    }
    function := args[FunctionArg]
    switch function {
        case GetCommand:
            return GetResponse{Value: w.KeyValMap[args[KeyArg]]}
        case SetCommand:
            w.KeyValMap[args[KeyArg]] = args[ValueArg]
            return nil
        case IncCommand:
            w.counter += 1
            return IncResponse{Value: w.counter}
    }
    return nil
}

// Don't need full implementation for testing.
func (w *WorkerFSM) Snapshot() (raft.FSMSnapshot, error) {
    return WorkerSnapshot{}, nil
}

// Don't need full implementation for testing.
func (w *WorkerFSM) Restore(i io.ReadCloser) error {
    return nil
}

// Don't need full implementation for testing.
func (s WorkerSnapshot) Persist(sink raft.SnapshotSink) error {return nil}

// Don't need full implementation for testing.
func (s WorkerSnapshot) Release() {}
