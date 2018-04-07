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
func CreateClient(trans *raft.NetworkTransport, servers []raft.ServerAddress) (*Client) {
    return &Client {
        trans:      trans,
        servers:    servers,
        session:    nil,
    }
}

// Cleanup associated with client.
func (c *Client) DestroyClient() error {
    if c.session != nil {
        return c.session.CloseClientSession()
    } else {
        return nil
    }
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
    session, session_err := c.getSession()
    if session_err != nil {
        return session_err
    }
    return session.SendRequest(data, &raft.ClientResponse{})
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
    session, sessionErr := c.getSession()
    if sessionErr != nil {
        return "", sessionErr
    }
    sendErr := session.SendRequest(data, &resp)
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

// Private method for opening new session if one doesn't exist.
func (c *Client) getSession() (*raft.Session, error) {
    existing := c.session
    if existing != nil {
        return existing, nil
    }
    newSession, err := raft.CreateClientSession(c.trans, c.servers)
    c.session = newSession
    return c.session, err
}
