package keyValStore 

import (
    "raft"
    "encoding/json"
)

type Client struct {
    raftClient      *raft.Client
}

// Create new client for sending RPCs.
func CreateClient(trans *raft.NetworkTransport, servers []raft.ServerAddress) (*Client) {
    return &Client {
        raftClient:     raft.CreateClient(trans, servers),
    }
}

// Cleanup associated with client.
func (c *Client) DestroyClient() error {
    return c.raftClient.DestroyClient()
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
    _, err := c.raftClient.SendRpc(data)
    return err
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
    resp, sendErr := c.raftClient.SendRpc(data)
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

