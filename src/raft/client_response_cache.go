package raft

import (
	"time"
)

// Manages the cache of client responses for use in RIFL, including
// garbage collecting the cache.

// clientResponseEntry holds state about the response to a client RPC.
// For use in RIFL.
type clientResponseEntry struct {
	response  interface{}
	timestamp time.Time
}

// Continuously check to garbage collect the cache.
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

// Garbage collect entries in the cache that have expired.
func (r *Raft) gcClientResponseCache() {
	r.clientResponseLock.RLock()
	currTime := time.Now()
	for clientID, clientCache := range r.clientResponseCache {
		for seqNo, entry := range clientCache {
			if currTime.Sub(entry.timestamp) >= r.conf.ClientResponseGcRemoveTime {
				r.clientResponseLock.RUnlock()
				r.clientResponseLock.Lock()
				delete(clientCache, seqNo) // does nothing if key does not exist, no race condition
				r.clientResponseLock.Unlock()
				r.clientResponseLock.RLock()
			}
		}
		r.clientResponseCache[clientID] = clientCache
	}
	r.clientResponseLock.RUnlock()
}
