package raft

import (
    "time"
    "errors"
    "sync"
)

type syncedConn struct {
    conn    *netConn
    lock    sync.Mutex
}

type Session struct {
    trans               *NetworkTransport
    conns               []syncedConn
    // Leader is index into conns or raftServers arrays.
    leader              int
    leaderLock          sync.RWMutex
    addrs               []ServerAddress
    // Client ID assigned by cluster for use in RIFL. 
    clientID            uint64
    // Sequence number of next RPC for use in RIFL.
    rpcSeqNo            uint64
}


/* Open client session to cluster. Takes clientID, server addresses for all servers in cluster, and returns success or failure.
   Start go routine to periodically send heartbeat messages and switch to new leader when necessary. */ 
func CreateClientSession(trans *NetworkTransport, addrs []ServerAddress) (*Session, error) {
    session := &Session{
        trans: trans,
        conns: make([]syncedConn, len(addrs)),
        leader: -1,
        addrs: addrs,
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
        RPCHeader: RPCHeader {
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

/* Make request to open session. */
func (s *Session) SendRequest(data []byte, keys []Key, resp *ClientResponse) error {
    if resp == nil {
        return errors.New("Response is nil")
    }
    req := ClientRequest {
        RPCHeader: RPCHeader {
            ProtocolVersion: ProtocolVersionMax,
        },
        Entry: &Log {
            Type: LogCommand,
            Data: data,
            Keys: keys,
        },
        ClientID: s.clientID,
        SeqNo: s.rpcSeqNo,
    }
    s.rpcSeqNo++
    return s.sendToActiveLeader(&req, resp, rpcClientRequest)
}

/* Make request to open session. Only use for testing purposes! */
func (s *Session) SendRequestWithSeqno(data []byte, keys []Key, resp *ClientResponse, seqno uint64) error {
    if resp == nil {
        return errors.New("Response is nil")
    }
    req := ClientRequest {
        RPCHeader: RPCHeader {
            ProtocolVersion: ProtocolVersionMax,
        },
        Entry: &Log {
            Type: LogCommand,
            Data: data,
            Keys: keys,
        },
        ClientID: s.clientID,
        SeqNo: seqno,
    }
    return s.sendToActiveLeader(&req, resp, rpcClientRequest)
}

/* Close client session. TODO: GC client request tables. */
func (s *Session) CloseClientSession() error {
    return nil
}

// TODO: send in parallel. If fail and haven't gotten a synced response from master, issue sync
func (s *Session) sendToAllWitnesses(entry *Log, leader int) bool {
    var err error
    request := &RecordRequest {
        Entry: entry,
    }

    // Send to all witnesses (excludes leader). 
    for i := range s.conns {
        if i == leader {
            continue
        }
        s.conns[i].lock.Lock()
        if s.conns[i].conn == nil {
            s.conns[i].conn, err = s.trans.getConn(s.addrs[i])
            if err != nil {
                s.conns[i].lock.Unlock()
                return false
            }
        }
        err = sendRPC(s.conns[i].conn, rpcRecordRequest, request)
        if err != nil {
            s.conns[i].lock.Unlock()
            return false
        }
        resp := &RecordResponse{}
        _, err = decodeResponse(s.conns[i].conn, resp)
        s.conns[i].lock.Unlock()
        if err != nil || !resp.Success {
            return false
        }
    }
    s.leaderLock.Lock()
    defer s.leaderLock.Unlock()
    if s.leader == leader {
        return true
    } else {
        // Active leader changed while sending to witness. Not sent
        // to f+1 distinct replicas.
        return false
    }
}

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
                time.Sleep(100*time.Millisecond)
            }
        } else {
            return nil
        }
    }

    return ErrNoActiveLeader
}

func findActiveServerWithTrans(addrs []ServerAddress, trans *NetworkTransport) (*netConn, error) {
    for _, addr := range(addrs) {
        conn, err := trans.getConn(addr)
        if err == nil {
            return conn, nil
        }
    }
    return nil, errors.New("No active raft servers.")
}
