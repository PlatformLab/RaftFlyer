## CURP

### Completed for CURP
* Record and sync RPCs.
* Keys sent with client requests to track commutativity in client operations.
* Accept records only if operations stored in witnesses don't commute.
* Master tries to apply command only locally if commutative. If not commutative, replicates synchronously and responds that it synced. 
* Master synchronously replicates commands sent in Sync RPCs.
* GC records at witnesses when done applying.
* Send to witnesses and master in parallel, check for success or sync. If failure, send sync to master.

### CURP Code Base
* `raft.go`: Witness state defined. Garbage collect at witnesses when operation completed. Support for handling record requests: accept and record if keys commutative and not leader, reject otherwise. Master syncs if operation not commutative, support for sync operation at master.
* `commands.go`: Sync and Record RPCs, add Synced field to ClientResponse to know if master synced. Add keys to ClientRequests.
* `session.go`: Sending to all witnesses and master in parallel. If all succeeded or synced at master, succeed. Otherwise, send Sync RPC to master. Keep repeating until success.  
* `log.go`: Update log entry to contain keys for commutativity checks.
* `api.go`: Add witness state to raft nodes.
* `net_transport.go`: Add new RPC types.

## RIFL

### Completed for RIFL
* Added client IDs and sequence numbers to client RPCs
* Assign client ID at master using global nextClientId
* Replicate nextClientId counter to other servers with LogNextClientId operation
* Store responses to client RPCs in cache that is periodically garbage-collected based on configurable timeout
* Check for duplicate before applying to state machine
* Make nextClientId and cache of client responses persistent.

### RIFL Code Base
* `raft.go`: Support for ClientId RPC handling, incrementing nextClientId at all replicas
* `fsm.go`: Before applying a command locally, check for cached response.
* `client_response_cache.go`: Stores state about the response to a client RPC along with a timestamp. Cache is periodically garbage collected.
* `session.go`: Starting a client session requires getting a new client ID, use that client ID and assign monotonically increasing sequence numbers for client RPCs.
* `commands.go`: RPC format for ClientRequest and ClientResponse updated to contain Client ID and sequence number, new RPC format ClientIdRequest and ClientIdResponse. GenericClientRequest for sending a request to a Raft leader.
* `log.go`: Update log entry to contain client IDs and sequence numbers.
* `config.go`: Set interval at which to garbage collect cache and how long responses to client RPCs should remain cached
* `api.go`: client response cache and next client ID state added to each raft node and snapshot restoring operations.
* `snapshot.go`: Support for snapshotting the client response cache and the next client ID (must be stored persistently).
* `file_snapshot.go`: Support for snapshotting the client response cache and the next client ID.
* `inmem_snapshot.go`: Support for snapshotting the client response cache and the next client ID.
* `net_transport.go`: Add new RPC types.

Run tests for RIFL: `src/test/runTests.sh`
Currently has a race condition
