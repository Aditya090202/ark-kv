# miniKV — Distributed Key-Value Store (Redis-compatible)

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

A Redis-protocol-compatible in-memory key-value store built from scratch in Go. Implements the RESP wire protocol with SET, GET, DEL, EXPIRE, TTL, and more — all backed by a concurrent-safe store with automatic TTL eviction.

Built as part of the [Coding Challenges](https://codingchallenges.fyi/) distributed systems series.

## Features

- **RESP protocol** — wire-compatible with `redis-cli` and any Redis client library
- **Concurrent-safe** — `sync.RWMutex` + background TTL eviction goroutine
- **TTL support** — per-key expiry with millisecond precision, auto-cleanup
- **Commands:** SET, GET, DEL, EXISTS, EXPIRE, TTL, KEYS, DBSIZE, FLUSHDB, PING, INFO
- **Dockerized** — single-container or multi-instance via Docker Compose

## Quick Start

```bash
# Build and run
make run

# Or with Docker
make docker-run
```

Then connect with any Redis client:

```bash
redis-cli -p 6379
miniKV:6379> SET name aditya
OK
miniKV:6379> GET name
"aditya"
miniKV:6379> EXPIRE name 10
(integer) 1
miniKV:6379> TTL name
(integer) 8
```

## Commands

| Command     | Args            | Description                          |
|-------------|-----------------|--------------------------------------|
| PING        | [message]       | Returns PONG or the message          |
| SET         | key value [EX seconds] [PX ms] | Store a key-value pair |
| GET         | key             | Retrieve a value                     |
| DEL         | key [key ...]   | Delete one or more keys              |
| EXISTS      | key             | Check if key exists                  |
| EXPIRE      | key seconds     | Set TTL on an existing key           |
| TTL         | key             | Get remaining TTL in seconds         |
| KEYS        | [pattern]       | List keys matching pattern           |
| DBSIZE      | —               | Number of keys in the store          |
| FLUSHDB     | —               | Remove all keys                      |
| INFO        | —               | Server statistics                    |

## Project Structure

```
cmd/server/main.go          — Entry point, TCP listener
internal/
  resp/reader.go            — RESP protocol decoder
  resp/reader_test.go       — RESP parser tests
  store/store.go            — Concurrent KV store with TTL
  store/store_test.go       — Store tests
  server/handler.go         — TCP server + command routing
go.mod
Makefile
Dockerfile
docker-compose.yml          — Multi-instance cluster
```

## Architecture

```
redis-cli -----> miniKV (:6379)
                    |
              ┌─────┴──────┐
              │  RESP Parser │
              └─────┬──────┘
                    │
              ┌─────┴──────┐
              │ Command    │
              │ Router     │
              └─────┬──────┘
                    │
              ┌─────┴──────┐
              │  KV Store   │  ← sync.RWMutex + TTL eviction goroutine
              └────────────┘
```

## Roadmap

- [x] RESP protocol (simple strings, errors, integers, bulk strings, arrays)
- [x] Thread-safe in-memory store with TTL
- [x] Basic Redis commands (SET, GET, DEL, EXISTS, EXPIRE, TTL, KEYS, PING, INFO)
- [ ] RDB snapshot persistence
- [ ] AOF append-only log
- [ ] Replication (master-replica)
- [ ] Cluster mode (sharding + gossip)

## Why Build This?

Building a Redis-compatible server from scratch teaches:

- **Wire protocols** — how Redis clients actually communicate (RESP)
- **Concurrency patterns** — read/write locks, goroutine lifecycle, background workers
- **Storage internals** — TTL expiration, eviction strategies, memory management
- **Systems thinking** — what "production-grade" actually means for a data store

## License

MIT