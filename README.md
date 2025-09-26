# go-crdt-sync

Local-first, CRDT-backed sync server with a Go client SDK. Enable offline-first apps to edit data locally and sync later without conflicts using well-defined CRDT merges.

## Status
- MVP scaffold: server, HTTP endpoints, basic CRDTs (LWW Register, OR-Set), Go client SDK, SQLite storage.
- Roadmap includes: op-log delta sync, WebSocket subscriptions, more CRDTs (PN-Counter, RGA), Postgres, auth, observability.

## Why
- Offline-first requires conflict-free merging when clients reconnect.
- CRDTs guarantee eventual consistency without central locking or complex conflict resolution.

## Features (MVP)
- Server: HTTP API for documents and ops, minimal auth headers.
- CRDTs: LWW Register, OR-Set.
- Storage: SQLite (WAL). Pluggable path to Postgres later.
- Client SDK: Go client with simple `Put` and `PushOps` methods.
- Example: Todo list using OR-Set.

## Quick Start
1) Prerequisites
   - Go 1.22+
2) Clone
   - `git clone https://github.com/yourname/go-crdt-sync.git`
   - `cd go-crdt-sync`
3) Run server
   - `make dev`
   - Server listens on `:8080`
4) Run example
   - `go run ./examples/todo`

## Configuration
- Environment
  - `ADDR`: server bind address (default `:8080`)
- Headers for requests
  - `X-Tenant-ID`: your tenant identifier (required)
  - `X-Node-ID`: client node identifier (required)
  - `X-API-Key`: optional for MVP (future: required)

## API (MVP)
- `PUT /v1/docs/{docID}`
  - Purpose: Create/overwrite a document with an LWW value (or initial OR-Set snapshot).
  - Headers: `X-Tenant-ID`, `X-Node-ID`
  - Body:
    ```json
    { "type": "lww" | "orset", "value": any }
    ```
- `POST /v1/docs/{docID}/ops`
  - Purpose: Apply operations and receive updated version and snapshot.
  - Headers: `X-Tenant-ID`, `X-Node-ID`
  - Body:
    ```json
    { "since": 0, "type": "lww" | "orset", "ops": [] }
    ```
  - Response:
    ```json
    { "version": 1, "snapshot": {} }
    ```

## Example Operations
- OR-Set
  - Add:
    ```json
    { "type": "add", "elem": "buy-milk" }
    ```
  - Remove:
    ```json
    { "type": "remove", "elem": "buy-milk" }
    ```
- LWW Register
  - Set:
    ```json
    { "set": { "name": "Ada" }, "ts_unix_ns": 0 }
    ```
    Note: Client typically sets a timestamp; server handles 0 when using `PUT`.

## Design Overview
- Identity: Tenant and node provided per request via headers.
- Versioning: Simple lamport-like counter per document for MVP (to evolve into op-log + causal metadata).
- CRDTs:
  - LWW Register: resolves by timestamp, then nodeID as tiebreaker.
  - OR-Set: observed-remove set with unique tags to prevent resurrection.

## Repository Layout (high level)
- `cmd/server`: server entrypoint
- `internal/api`: HTTP and WebSocket handlers
- `internal/crdt`: CRDT implementations (LWW, OR-Set)
- `internal/store`: persistence (SQLite), models
- `pkg/client`: Go client SDK
- `examples/todo`: sample app using OR-Set

## Development
- Run server: `make dev`
- Run tests: `make test`
- Lint: to be added (golangci-lint), plus govulncheck in CI

## Roadmap
- Delta sync with per-document op-log and `GET /ops?since`
- WebSocket subscriptions and backfill on reconnect
- Additional CRDTs: PN-Counter, ordered list (RGA/Logoot)
- Postgres storage with migrations and indexes
- AuthN/Z with API keys, rate limits, quotas
- Observability: Prometheus metrics, OpenTelemetry traces
- JS/TS client SDK and sample web app

## Security
- Do not expose the MVP publicly without proper authentication, rate limiting, and TLS. This is a work-in-progress.

## License
- MIT (placeholder; choose your preferred license)

## Credits
- Built with Go.
- Libraries: gorilla/mux, gorilla/websocket, GORM, zerolog.