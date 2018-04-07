package raft

type Client struct {
    // Addresses of servers in cluster.
    servers     []ServerAddress
    // Transport layer for communications.
    trans       *NetworkTransport
    // Current session with cluster leader.
    session     *Session
    // ID of client determined by cluster for use in RIFL. 
    clientID    int64
    // Sequence number for next RPC for use in RIFL.
    rpcSeqNo    int64
}

func CreateClient(trans *NetworkTransport, servers []ServerAddress) (*Client) {
    c := &Client {
        trans:      trans,
        servers:    servers,
        session:    nil,
        rpcSeqNo:   0,
    }
    c.clientID = 0     // TODO: get this value from server
    return c
}

func (c *Client) DestroyClient() error {
    if c.session != nil {
        return c.session.CloseClientSession()
    } else {
        return nil 
    }
}

func (c *Client) SendRpc(rpc []byte) (*ClientResponse, error) {
    session, session_err := c.getSession()
    if session_err != nil {
        return &ClientResponse{}, session_err
    }
    resp := &ClientResponse{}
    send_err := session.SendRequest(rpc, resp)
    if send_err != nil {
        return &ClientResponse{}, send_err
    }
    return resp, nil
}

// Private function for opening new session if one doesn't exist.
func (c *Client) getSession() (*Session, error) {
    existing := c.session
    if existing != nil {
        return existing, nil
    }
    newSession, err := CreateClientSession(c.trans, c.servers, nil)
    c.session = newSession
    return c.session, err
}
