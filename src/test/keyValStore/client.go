package keyValStore 

import (
    "raft"
    "encoding/json"
)

type Client struct {
    // Client transport layer.
    trans       *raft.NetworkTransport
    // Addresses in cluster.
    servers     []raft.ServerAddress
    // Open session with cluster leader.
    session     *raft.Session
}

// Create new client for sending RPCs.
func CreateClient(trans *raft.NetworkTransport, servers []raft.ServerAddress) (*Client, error) {
    newSession, err := raft.CreateClientSession(trans, servers)
    if err != nil {
        return nil, err
    }
    return &Client {
        trans:      trans,
        servers:    servers,
        session:    newSession,
    }, nil
}

// Cleanup associated with client.
func (c *Client) DestroyClient() {
    c.session.CloseClientSession()
}

func (c *Client) Inc() (uint64, error) {
    args := make(map[string]string)
    args[FunctionArg] = IncCommand
    data, marshal_err := json.Marshal(args)
    if marshal_err != nil {
        return 0, marshal_err
    }
    resp := raft.ClientResponse{}
    keys := []raft.Key{raft.Key([]byte{1})}
    sendErr := c.session.SendRequest(data, keys, &resp)
    if sendErr != nil {
        return 0, sendErr
    }
    var response IncResponse
    recvErr := json.Unmarshal(resp.ResponseData, &response)
    if recvErr != nil {
        return 0, recvErr
    }
    return response.Value, nil
}

func (c *Client) IncWithSeqno(seqno uint64) (uint64, error) {
    args := make(map[string]string)
    args[FunctionArg] = IncCommand
    data, marshal_err := json.Marshal(args)
    if marshal_err != nil {
        return 0, marshal_err
    }
    resp := raft.ClientResponse{}
    keys := []raft.Key{raft.Key([]byte{1})}
    sendErr := c.session.SendRequestWithSeqNo(data, keys, &resp, seqno)
    if sendErr != nil {
        return 0, sendErr
    }
    var response IncResponse
    recvErr := json.Unmarshal(resp.ResponseData, &response)
    if recvErr != nil {
        return 0, recvErr
    }
    return response.Value, nil
}


// Send RPC to set the value of a key. No expected response.
func (c *Client) Set(key string, value string) error {
    args := make(map[string]string)
    args[FunctionArg] = SetCommand
    args[KeyArg] = key
    args[ValueArg] = value
    data, marshal_err := json.Marshal(args)
    if marshal_err != nil  {
        return marshal_err
    }
    keys := []raft.Key{raft.Key([]byte(key))}
    return c.session.SendRequest(data, keys, &raft.ClientResponse{})
}

// Send RPC to get the value of a key. Empty string if error. 
func (c *Client) Get(key string) (string, error) {
    args := make(map[string]string)
    args[FunctionArg] = GetCommand
    args[KeyArg] = key
    data, marshal_err := json.Marshal(args)
    if marshal_err != nil {
        return "", marshal_err
    }
    resp := raft.ClientResponse{}
    keys := []raft.Key{raft.Key([]byte(key))}
    sendErr := c.session.SendRequest(data, keys, &resp)
    if sendErr != nil {
        return "", sendErr
    }
    var response GetResponse
    recvErr := json.Unmarshal(resp.ResponseData, &response)
    if recvErr != nil {
        return "", recvErr
    }
    return response.Value, nil
}
