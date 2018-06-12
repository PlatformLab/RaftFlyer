package keyValStore 

import (
    "raft"
    "encoding/json"
)

// Client library for sending RPCs to keyValStore Raft servers. Allows you
// to set a key, get a value, and increment an integer.

// Handle given to client to make reqeuests.
type Client struct {
    // Client transport layer.
    trans       *raft.NetworkTransport
    // Addresses in cluster.
    servers     []raft.ServerAddress
    // Open session with cluster leader.
    session     *raft.Session
}

// Create new client for sending RPCs.
// Params:
//   - trans: transport layer for client.
//   - servers: list of Raft server addresses.
// Returns: client handle, error if any.
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

// Increment integer and return value. Not idempotent!
// Returns: updated value of integer.
func (c *Client) Inc() (uint64, error) {
    args := make(map[string]string)
    args[FunctionArg] = IncCommand
    data, marshal_err := json.Marshal(args)
    if marshal_err != nil {
        return 0, marshal_err
    }
    resp := raft.ClientResponse{}
    keys := []raft.Key{raft.Key([]byte{1})}
    c.session.SendFastRequest(data, keys, &resp)
    var response IncResponse
    recvErr := json.Unmarshal(resp.ResponseData, &response)
    if recvErr != nil {
        return 0, recvErr
    }
    return response.Value, nil
}

// Same as Inc() but specifies sequence number. Use for testing
// purposes only (in production only use Inc).
// Params:
//   - seqno: Sequence number to use when sending RPC.
// Returns: updated value of integer.
func (c *Client) IncWithSeqno(seqno uint64) (uint64, error) {
    args := make(map[string]string)
    args[FunctionArg] = IncCommand
    data, marshal_err := json.Marshal(args)
    if marshal_err != nil {
        return 0, marshal_err
    }
    resp := raft.ClientResponse{}
    keys := []raft.Key{raft.Key([]byte{1})}
    c.session.SendFastRequestWithSeqNo(data, keys, &resp, seqno)
    var response IncResponse
    recvErr := json.Unmarshal(resp.ResponseData, &response)
    if recvErr != nil {
        return 0, recvErr
    }
    return response.Value, nil
}


// Send RPC to set the value of a key. No expected response.
// Params:
//   - key: Key to access.
//   - value: Value to set with key.
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
    c.session.SendFastRequest(data, keys, &raft.ClientResponse{})
    return nil
}

// Send RPC to get the value of a key. 
// Params:
//   - key: Key to get value of.
// Returns: value of key, empty string if error not nil.
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
    c.session.SendFastRequest(data, keys, &resp)
    var response GetResponse
    recvErr := json.Unmarshal(resp.ResponseData, &response)
    if recvErr != nil {
        return "", recvErr
    }
    return response.Value, nil
}
