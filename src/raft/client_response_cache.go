package raft

import(
    "time"
)

// clientResponse holds state about the response to a client RPC.
// For use in RIFL.
type clientResponseEntry struct {
    responseData        []byte
    timestamp   time.Time
}

func (r *Raft) runGcClientResponseCache() {
    for {
        select {
            case <-randomTimeout(r.conf.ClientResponseGcInterval):
                r.gcClientResponseCache()

            case <-r.shutdownCh:
                return
        }
    }
}

func (r *Raft) gcClientResponseCache() {
    r.clientResponseLock.RLock()
    currTime := time.Now()
    for clientID, clientCache := range r.clientResponseCache {
        for seqNo, entry := range clientCache {
            if currTime.Sub(entry.timestamp) >= r.conf.ClientResponseGcRemoveTime {
                r.clientResponseLock.RUnlock()
                r.clientResponseLock.Lock()
                delete(clientCache, seqNo)  // does nothing if key does not exist, no race condition
                r.clientResponseLock.Unlock()
                r.clientResponseLock.RLock()
            }
        }
        r.clientResponseCache[clientID] = clientCache
    }
    r.clientResponseLock.RUnlock()
}
