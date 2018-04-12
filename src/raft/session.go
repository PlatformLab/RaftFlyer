package raft

import (
    "time"
    "fmt"
    "errors"
)

type Session struct {
    trans               *NetworkTransport
    currConn            *netConn
    raftServers         []ServerAddress
    stopCh              chan bool
    active              bool
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
        raftServers: addrs,
        active: true,
        stopCh : make(chan bool, 1),
        rpcSeqNo: 0,
    }
    var err error
    session.currConn, err = findActiveServerWithTrans(addrs, trans)
    if err != nil {
        return nil ,err
    }
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
    fmt.Println("New client id is ", session.clientID)
    return session, nil
}


/* Make request to open session. */
func (s *Session) SendRequest(data []byte, resp *ClientResponse) error {
    if !s.active {
        return errors.New("Inactive client session.")
    }
    if resp == nil {
        return errors.New("Response is nil")
    }
    req := ClientRequest {
        RPCHeader: RPCHeader {
            ProtocolVersion: ProtocolVersionMax,
        },
        Entries: []*Log{
            &Log {
                Type: LogCommand,
                Data: data,
            },
        },
        ClientID: s.clientID,
        SeqNo: s.rpcSeqNo,
    }
    s.rpcSeqNo++
    return s.sendToActiveLeader(&req, resp, rpcClientRequest)
}


/* Close client session. Kill heartbeat go routine. */
func (s *Session) CloseClientSession() error {
    if !s.active {
        return errors.New("Inactive client session")
    }
    s.stopCh <- true
    fmt.Println("closed client session")
    return nil
}

func (s *Session) sendToActiveLeader(request interface{}, response GenericClientResponse, rpcType uint8) error {
    var err error = errors.New("")
    retries := 5
    /* Send heartbeat to active leader. Connect to active leader if connection no longer to active leader. */
    for err != nil {
        if retries <= 0 {
            s.active = false
            return errors.New("Failed to find active leader.")
        }
        if s.currConn == nil {
            s.active = false
            return errors.New("No current connection.")
        }
        err = sendRPC(s.currConn, rpcType, request)
        /* Try another server if server went down. */
        for err != nil {
            if retries <= 0 {
                s.active = false
                return errors.New("Failed to find active leader.")
            }
            s.currConn, err = findActiveServerWithTrans(s.raftServers, s.trans)
            if err != nil || s.currConn == nil {
                s.active = false
                return errors.New("No active server found.")
            }
            retries--
            err = sendRPC(s.currConn, rpcType, request)
        }
        /* Decode response if necesary. Try new server to find leader if necessary. */
        if (s.currConn == nil) {
            return errors.New("Failed to find active leader.")
        }
        _, err = decodeResponse(s.currConn, &response)
        if err != nil {
            if response != nil && response.GetLeaderAddress() != "" {
                s.currConn, _ = s.trans.getConn(response.GetLeaderAddress())
             } else {
                /* Wait for leader to be elected. */
                time.Sleep(1000*time.Millisecond)
            }
        }
        retries--
    }
    return nil
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
