package server

import(
    "raft"
    "encoding/json"
    "fmt"
    "io"
)

// *WorkerFSM implements raft.FSM by implementing Apply,
// Snapshot, Restore.
type WorkerFSM struct {

}

// Create array of worker FSMs for starting a cluster.
func CreateWorkers(n int) ([]raft.FSM) {
    workers := make([]*WorkerFSM, n)
    for i := range workers {
        workers[i] = &WorkerFSM{}
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
    }
    // TODO: implement key-value store here
    fmt.Println("applying command: ", args)
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
