package raft

import (
	"errors"
	"sync"
	"time"
)

// Client library for Raft. Provides session abstraction that handles starting
// a session, making requests, and closing a session.

// Connection and associated lock for synchronization.
type syncedConn struct {
	// Connection to Raft server.
	conn *netConn
	// Lock protecting conn.
	lock sync.Mutex
}

// Session abstraction used to make requests to Raft cluster.
type Session struct {
	// Client network layer.
	trans *NetworkTransport
	// Connections to all Raft nodes.
	conns []syncedConn
	// Leader is index into conns or addrs arrays.
	leader     int
	leaderLock sync.RWMutex
	// Addresses of all Raft servers.
	addrs []ServerAddress
	// Client ID assigned by cluster for use in RIFL.
	clientID uint64
	// Sequence number of next RPC for use in RIFL.
	rpcSeqNo uint64
}

// Open client session to cluster.
// Params:
//   - trans: Client transport layer for networking opertaions
//   - addrs: Addresses of all Raft servers
// Return: created session
func CreateClientSession(trans *NetworkTransport, addrs []ServerAddress) (*Session, error) {
	session := &Session{
		trans:    trans,
		conns:    make([]syncedConn, len(addrs)),
		leader:   -1,
		addrs:    addrs,
		rpcSeqNo: 0,
	}

	// Initialize syncedConn array.
	for i := range session.conns {
		session.conns[i] = syncedConn{}
	}

	// Open connections to all raft servers.
	var err error
	for i, addr := range addrs {
		session.conns[i].conn, err = trans.getConn(addr)
		if err == nil {
			session.leader = i
		}
	}

	// Report error if can't connect to any server.
	if session.leader == -1 {
		return nil, ErrNoActiveServers
	}

	// Get a client ID from the leader.
	req := ClientIdRequest{
		RPCHeader: RPCHeader{
			ProtocolVersion: ProtocolVersionMax,
		},
	}
	resp := ClientIdResponse{}
	err = session.sendToActiveLeader(&req, &resp, rpcClientIdRequest)
	if err != nil {
		return nil, err
	}
	session.clientID = resp.ClientID
	return session, nil
}

// Make request to Raft cluster using open session.
// Params:
//   - data: client request to send to cluster
//   - keys: array of keys that request updates, used in commutativity checks
//   - resp: pointer to response that will be populated
func (s *Session) SendRequest(data []byte, keys []Key, resp *ClientResponse) error {
	seqNo := s.rpcSeqNo
	s.rpcSeqNo++
	return s.SendRequestWithSeqNo(data, keys, resp, seqNo)
}

// Make request to Raft cluster using open session and specifying a sequence
// number. Only use for testing! (Use SendRequest in production).
// Params:
//   - data: client request to send to cluster
//   - keys: array of keys that request updates, used in commutativity checks
//   - resp: pointer to response that will be populated
//   - seqno: sequence number to use for request (for testing purposes)
func (s *Session) SendRequestWithSeqNo(data []byte, keys []Key, resp *ClientResponse, seqno uint64) error {
	if resp == nil {
		return errors.New("Response is nil")
	}
	req := ClientRequest{
		RPCHeader: RPCHeader{
			ProtocolVersion: ProtocolVersionMax,
		},
		Entry: &Log{
			Type:     LogCommand,
			Data:     data,
			Keys:     keys,
			ClientID: s.clientID,
			SeqNo:    seqno,
		},
	}
	return s.sendToActiveLeader(&req, resp, rpcClientRequest)
}

// Close client session.
// TODO: GC client request tables.
func (s *Session) CloseClientSession() error {
	return nil
}

// Make request to Raft cluster following CURP protocol. Send to witnesses and
// master simultaneously to complete in 1 RTT.
// Params:
//   - data: client request to send to cluster
//   - keys: array of keys that request updates, used in commutativity checks
//   - resp: pointer to response that will be populated
//   - seqno: sequence number to use for request (for testing purposes)
func (s *Session) SendFastRequest(data []byte, keys []Key, resp *ClientResponse) {
	seqNo := s.rpcSeqNo
	s.rpcSeqNo++
	s.SendFastRequestWithSeqNo(data, keys, resp, seqNo)
}

// Make request to Raft cluster following CURP protocol. Send to witnesses and
// master simultaneously to complete in 1 RTT. Specify sequence number for testing
// purposes. Only use SendFastRequest in production!
// Params:
//   - data: client request to send to cluster
//   - keys: array of keys that request updates, used in commutativity checks
//   - resp: pointer to response that will be populated
//   - seqno: sequence number to use for request (for testing purposes)
func (s *Session) SendFastRequestWithSeqNo(data []byte, keys []Key, resp *ClientResponse, seqNo uint64) {
	req := ClientRequest{
		RPCHeader: RPCHeader{
			ProtocolVersion: ProtocolVersionMax,
		},
		Entry: &Log{
			Type:     LogCommand,
			Data:     data,
			Keys:     keys,
			ClientID: s.clientID,
			SeqNo:    seqNo,
		},
	}

	// Repeat until success.
	// TODO: only retry limited number of times
	for true {
		resultCh := make(chan bool, len(s.addrs))
		go func(s *Session, req *ClientRequest, resp *ClientResponse, resultCh *chan bool) {
			err := s.sendToActiveLeader(req, resp, rpcClientRequest)
			//fmt.Println("err sending to leader: ", err)
			if err != nil {
				*resultCh <- false
			} else {
				*resultCh <- true
			}
		}(s, &req, resp, &resultCh)
		s.sendToAllWitnesses(req.Entry, &resultCh)

		success := true

		for i := 0; i <= len(s.addrs); i += 1 { // TODO: should this be len + 1?
			result := <-resultCh
			//fmt.Println("result is ", result)
			success = success && result
			// TODO: if synced, automatically succeed, otherwise if not success need to retry
		}
		if success || resp.Synced {
			return
		}
		// If fail to record at witnesses and not synced, issue sync request.
		sync := &SyncRequest{
			RPCHeader: RPCHeader{
				ProtocolVersion: ProtocolVersionMax,
			},
		}
		var syncResp *SyncResponse
		err := s.sendToActiveLeader(sync, syncResp, rpcSyncRequest)
		if err == nil && syncResp.Success {
			return
		}
		// Failed to sync. Try everything again
	}

}

// Send log entry to all witnesses in parallel and put results (success
// or failure) into channel. Get all values from channel to ensure that
// RPCs to witnesses have completed.
// Params:
//   - entry: Log entry to send to all witnesses.
//   - resultCh: channel to put completion status into.
func (s *Session) sendToAllWitnesses(entry *Log, resultCh *chan bool) {
	req := &RecordRequest{
		RPCHeader: RPCHeader{
			ProtocolVersion: ProtocolVersionMax,
		},
		Entry: entry,
	}

	// Send to all witnesses.
	for i := range s.conns {
		go func(req *RecordRequest, resultCh *chan bool) {
			*resultCh <- s.sendToWitness(i, req)
		}(req, resultCh)
	}
}

// Send request to a witness specified by id. Synchronous.
// Params:
//   - id: ID of witness sending request to
//   - req: RecordRequest to send to witness
// Returns: success or failure of RPC.
func (s *Session) sendToWitness(id int, req *RecordRequest) bool {
	var err error
	s.conns[id].lock.Lock()
	if s.conns[id].conn == nil {
		s.conns[id].conn, err = s.trans.getConn(s.addrs[id])
		if err != nil {
			s.conns[id].lock.Unlock()
			return false
		}
	}
	err = sendRPC(s.conns[id].conn, rpcRecordRequest, req)
	if err != nil {
		s.conns[id].lock.Unlock()
		return false
	}
	resp := &RecordResponse{}
	_, err = decodeResponse(s.conns[id].conn, resp)
	s.conns[id].lock.Unlock()
	if err != nil || !resp.Success {
		return false
	}
	return true
}

// Send a RPC to the active leader. Try to use the currently cached active leader, and
// if there is no cached leader or it is unreachable, try other Raft servers until a
// leader is found. If no active Raft server is found, return an error.
// Params:
//   - request: JSON representation of request
//   - response: client response that contains a leader address to help find an active leader
//   - rpcType: type of RPC being sent.
func (s *Session) sendToActiveLeader(request interface{}, response GenericClientResponse, rpcType uint8) error {
	sendFailures := 0
	var err error

	s.leaderLock.Lock()
	defer s.leaderLock.Unlock()

	// Continue trying to send until have tried contacting all servers.
	for sendFailures < len(s.addrs) {
		// If no open connection to guessed leader, try to open one.
		s.conns[s.leader].lock.Lock()
		if s.conns[s.leader].conn == nil {
			s.conns[s.leader].conn, err = s.trans.getConn(s.addrs[s.leader])
			if err != nil {
				s.conns[s.leader].lock.Unlock()
				sendFailures += 1
				s.leader = (s.leader + 1) % len(s.conns)
				continue
			}
		}
		err = sendRPC(s.conns[s.leader].conn, rpcType, request)

		// Failed to send RPC - try next server.
		if err != nil {
			s.conns[s.leader].lock.Unlock()
			sendFailures += 1
			s.leader = (s.leader + 1) % len(s.conns)
			continue
		}

		// Try to decode response.
		_, err = decodeResponse(s.conns[s.leader].conn, &response)
		s.conns[s.leader].lock.Unlock()

		// If failure, use leader hint or wait for election to complete.
		if err != nil {
			if response != nil && response.GetLeaderAddress() != "" {
				s.leader = (s.leader + 1) % len(s.conns)
				for i, addr := range s.addrs {
					if addr == response.GetLeaderAddress() {
						s.leader = i
						break
					}
				}
			} else {
				time.Sleep(100 * time.Millisecond)
			}
		} else {
			return nil
		}
	}

	return ErrNoActiveLeader
}
