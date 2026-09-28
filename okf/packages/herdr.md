---
type: Go Package
title: internal/herdr
description: Client for the Herdr local socket — JSON-RPC on one connection per call, plus a streaming event subscription
tags: [ipc, socket, json-rpc, client]
timestamp: 2026-09-03T00:00:00Z
---

# Responsibilities

Speaks Herdr's local-socket protocol: newline-delimited JSON, id-correlated request/response, plus a long-lived event stream. Consumed by [server.Hub](/packages/server.md). The transport is a Unix domain socket, or on Windows the named pipe `\\.\pipe\<socket path>` (the `herdr.sock` file there only records the server PID); `socket_{windows,other}.go` hold the platform split.

# Surface

| Symbol | Purpose |
|---|---|
| `New(socketPath)` | construct (empty path → `DefaultSocketPath`) |
| `Call(ctx, method, params) (json.RawMessage, error)` | one RPC on its own connection |
| `Snapshot(ctx, out)` | `session.snapshot`, unwrapping the `{snapshot:…}` envelope; rides `Call` |
| `Subscribe(ctx, types, onEvent)` | long-lived events connection; blocks until drop/ctx |
| `ConfigDir()` | Herdr's per-user dir: `~/.config/herdr`, `%APPDATA%\herdr` on Windows |
| `DefaultSocketPath()` / `SessionSocketPath(name)` | `<ConfigDir>/herdr.sock` / `<ConfigDir>/sessions/<name>/herdr.sock` |
| `Listen(path)` | serve a Herdr-style socket (Unix socket or named pipe); test fakes use it |
| `AllEventTypes` | global resource events subscribed by default |

# One connection per call

Herdr answers exactly one request per connection and then closes it, as the CLI assumes. Measured against Herdr 0.9.1-preview on Windows: a second request written after the first reply finds the pipe closing, and requests written while one is in progress wait behind it and then fail with EOF.

- `Call` dials, writes one request line, reads one reply line, and closes. `context.AfterFunc` closes the connection when `ctx` ends, which unblocks the read.
- A call Herdr holds open (such as `agent.prompt` with a `wait`) therefore never delays other calls. Until 2026-09-28 the client multiplexed every call over one persistent connection (`bridge-ipc-hardening`); behind a prompt's wait, each `pane.read` and snapshot stalled until the agent went idle, and concurrent calls failed with EOF.
- `Subscribe` keeps its **own** long-lived connection (streaming, no id correlation).

`TestSlowCallDoesNotStallOthers` in `internal/herdr/client_test.go` runs against a one-request-per-connection fake.

# Citations

* [internal/herdr/client.go](/internal/herdr/client.go)
* [Herdr socket API](/references/herdr-socket-api.md)
