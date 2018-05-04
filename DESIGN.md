### CURP

## Completed for CURP
* Record and sync RPCs
* Keys to track commutativity in client operations
* Accept records only if don't conflict
* GC records at witnesses when done applying
* Send to witnesses in parallel

### RIFL

## Completed for RIFL
* Added client IDs and sequence numbers to client RPCs
* Assign client ID at master using global nextClientId
* Replicate nextClientId counter to other servers with LogNextClientId operation
* Store responses to client RPCs in cache that is periodically garbage-collected based on configurable timeout
* Check for duplicate before applying to state machine
* Make nextClientId and cache of client responses persistent.

## RIFL Code Base
* `raft.go`: Support for ClientId RPC handling, incrementing nextClientId at all replicas
* `client_response_cache.go`: Stores state about the response to a client RPC along with a timestamp. Cache is periodically garbage collected.
* `config.go`: Set interval at which to garbage collect cache and how long responses to client RPCs should remain cached
* `snapshot.go`: Support for snapshotting the client response cache and the next client ID (must be stored persistently).
* `commands.go`: RPC format for ClientRequest and ClientResponse updated to contain Client ID and sequence number, new RPC format ClientIdRequest and ClientIdResponse
* `session.go`: Starting a client session requires getting a new client ID, use that client ID and assign monotonically increasing sequence numbers for cliient RPCs.

Run tests for RIFL: `src/test/runTests.sh`
