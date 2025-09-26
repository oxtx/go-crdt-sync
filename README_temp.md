# go-crdt-sync

Minimal CRDT sync service in Go with:
- LWW register (last-writer-wins)
- OR-Set (Observed-Remove Set)
- In-memory store
- Simple HTTP API (chi)

Requirements
- Go 1.21+

Getting started
1) Clone and setup
   git clone https://github.com/oxtx/go-crdt-sync.git
   cd go-crdt-sync
   make tidy

2) Run (foreground)
   make run
   # or change port
   make run ADDR=:9090

3) Run (background) and stop
   make run-bg
   make stop

4) Test
   make test
   # verbose
   make testv
   # static analysis
   make vet

5) Build
   make build
   # binary at bin/server
   ADDR=:8080 ./bin/server

API quick reference
- PUT /v1/docs/{docID}
  - LWW body:
    { "type": "lww", "snapshot": { "value": "hello", "ts": 1, "nodeId": "n1" } }
    or
    { "type": "lww", "snapshot": "hello" }
  - OR-Set body:
    { "type": "orset", "snapshot": ["a","b"] }
  - Response:
    { "version": 0, "type": "lww|orset", "snapshot": ... }

- GET /v1/docs/{docID}
  - Response:
    { "version": N, "type": "lww|orset", "snapshot": ... }

- POST /v1/docs/{docID}/ops
  - LWW write:
    { "since": 0, "ops": [ { "lww_write": { "value": "blue", "ts": 2, "nodeId": "nA" } } ] }
  - OR-Set ops:
    { "since": 0, "ops": [
        { "orset_add":    { "item": "tag1", "tag": "u1-123" } },
        { "orset_remove": { "item": "tag1", "seen": ["u1-123"] } }
      ] }
  - Response:
    { "version": N, "snapshot": ... }

Detailed run and test results

A) Run the server
- Foreground:
  make run
  Expected output (example):
  listening on :8080

- Background:
  make run-bg
  Creates:
  - bin/server (binary)
  - bin/server.out (logs)
  - .server.pid (PID file)
  Stop:
  make stop
  Expected:
  "Server stopped." and .server.pid removed.

B) Manual verification with curl
- Create LWW doc:
  curl -s -X PUT localhost:8080/v1/docs/colors \
    -H 'content-type: application/json' \
    -d '{ "type":"lww", "snapshot":"red" }'
  Expected (example):
  { "version":0, "type":"lww", "snapshot":{"value":"red","ts":169...,"nodeId":"server"} }

- Apply LWW write:
  curl -s -X POST localhost:8080/v1/docs/colors/ops \
    -H 'content-type: application/json' \
    -d '{ "since":0, "ops":[ { "lww_write": { "value":"blue", "ts": 2, "nodeId":"n1" } } ] }'
  Expected:
  { "version":1, "snapshot":{"value":"blue","ts":2,"nodeId":"n1"} }

- Read back:
  curl -s localhost:8080/v1/docs/colors
  Expected:
  { "version":1, "type":"lww", "snapshot":{"value":"blue","ts":2,"nodeId":"n1"} }

- Create OR-Set doc:
  curl -s -X PUT localhost:8080/v1/docs/tags \
    -H 'content-type: application/json' \
    -d '{ "type":"orset", "snapshot":["a","b"] }'
  Expected:
  { "version":0, "type":"orset", "snapshot":["a","b"] }

- Add item:
  curl -s -X POST localhost:8080/v1/docs/tags/ops \
    -H 'content-type: application/json' \
    -d '{ "since":0, "ops":[ { "orset_add": { "item":"c", "tag":"u1-1" } } ] }'
  Expected:
  { "version":1, "snapshot":["a","b","c"] }   // order may vary

- Remove item:
  curl -s -X POST localhost:8080/v1/docs/tags/ops \
    -H 'content-type: application/json' \
    -d '{ "since":1, "ops":[ { "orset_remove": { "item":"c", "seen":["u1-1"] } } ] }'
  Expected:
  { "version":2, "snapshot":["a","b"] }

Notes
- This scaffold uses in-memory storage; data resets on restart.
- For production, add durable storage, op IDs, causal metadata (HLC/vector clocks), auth, and pagination.
