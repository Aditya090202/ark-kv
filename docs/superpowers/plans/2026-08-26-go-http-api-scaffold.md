# Go HTTP API Scaffold Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a learning-oriented Go project skeleton for the clstr HTTP API stage without implementing server, handler, or storage behavior.

**Architecture:** Keep the executable entry point under `cmd/server` and place non-exported application concerns under `internal`. Separate HTTP transport placeholders from the future concurrency-safe store so later walkthrough stages can evolve independently.

**Tech Stack:** Go 1.26.1, Go standard library, Docker

**Spec:** `docs/superpowers/specs/2026-08-26-go-http-api-scaffold-design.md`

## Global Constraints

- Module path: `github.com/Aditya090202/ark-kv`.
- Provide structural placeholders and learning-oriented TODO comments only.
- Do not implement HTTP routing, request handling, storage operations, persistence, or replication.
- Reserve TCP port `8080` in the Docker scaffold.
- Do not add third-party dependencies or tests.

---

### Task 1: Create the Go HTTP API project skeleton

**Files:**
- Create: `go.mod`
- Create: `cmd/server/main.go`
- Create: `internal/httpapi/server.go`
- Create: `internal/httpapi/handlers.go`
- Create: `internal/store/store.go`
- Create: `Dockerfile`

**Interfaces:**
- Produces: package `main` as the future executable entry point.
- Produces: placeholder type `httpapi.Server`.
- Produces: placeholder type `httpapi.Handlers`.
- Produces: placeholder type `store.Store`.
- No behavior or cross-package wiring is introduced.

- [ ] **Step 1: Declare the Go module**

Create `go.mod`:

```go
module github.com/Aditya090202/ark-kv

go 1.26.1
```

- [ ] **Step 2: Add the executable entry-point placeholder**

Create `cmd/server/main.go`:

```go
// Command server will run a single key-value store node.
package main

// TODO: Construct the store and HTTP server, then listen on port 8080.
```

- [ ] **Step 3: Add HTTP transport placeholders**

Create `internal/httpapi/server.go`:

```go
// Package httpapi will expose the key-value store over HTTP.
package httpapi

// Server will own HTTP server configuration and route registration.
type Server struct{}

// TODO: Add server construction, route registration, and startup behavior.
```

Create `internal/httpapi/handlers.go`:

```go
package httpapi

// Handlers will contain the HTTP endpoint dependencies.
type Handlers struct{}

// TODO: Add PUT, GET, DELETE, clear, and health handlers.
// TODO: Return the walkthrough's required 400, 404, and 405 responses.
```

- [ ] **Step 4: Add the storage placeholder**

Create `internal/store/store.go`:

```go
// Package store will provide concurrency-safe key-value storage.
package store

// Store will hold the node's in-memory key-value data.
type Store struct{}

// TODO: Add concurrency-safe read, write, delete, and clear operations.
```

- [ ] **Step 5: Add the container scaffold**

Create `Dockerfile`:

```dockerfile
FROM golang:1.26-alpine

WORKDIR /app
COPY . .

EXPOSE 8080

# TODO: Add build and startup instructions when the server is implemented.
```

- [ ] **Step 6: Format and inspect the scaffold**

Run:

```bash
gofmt -w cmd/server/main.go internal/httpapi/server.go internal/httpapi/handlers.go internal/store/store.go
go mod edit -json
go list ./...
```

Expected:

- `gofmt` exits successfully.
- `go mod edit -json` reports module path `github.com/Aditya090202/ark-kv` and Go version `1.26.1`.
- `go list ./...` reports the `cmd/server`, `internal/httpapi`, and `internal/store` packages.
- No server process, endpoint, or storage operation is functional yet.

- [ ] **Step 7: Review the resulting scope**

Confirm that only the six scaffold files were added, aside from the approved design and plan documents. Confirm that every functionally incomplete area is explicitly identified by a TODO and that no third-party dependency or endpoint implementation was introduced.
