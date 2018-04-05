package keyValStore

// Client RPCs.
const GetCommand string = "Get"
const SetCommand string = "Set"
const FunctionArg string = "function"
const KeyArg string = "key"
const ValueArg string = "value"

// Response to Get RPC. 
type GetResponse struct {
    Value string
}
