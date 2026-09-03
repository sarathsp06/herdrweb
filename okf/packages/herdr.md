---
type: Go Package
title: internal/herdr
description: Client for the Herdr local socket — persistent id-multiplexed JSON-RPC plus a streaming event subscription
tags: [ipc, socket, json-rpc, client]
timestamp: 2026-09-03T00:00:00Z
---

# Responsibilities

Speaks Herdr's Unix-socket protocol: newline-delimited JSON, id-correlated request/response, plus a long-lived event stream. Consumed by [server.Hub](/packages/server.md).

# Surface

| Symbol | Purpose |
|---|---|
| `New(socketPath)` | construct (empty path → `DefaultSocketPath`) |
| `Call(ctx, method, params) (json.RawMessage, error)` | one RPC over the persistent connection |
| `Snapshot(ctx, out)` | `session.snapshot`, unwrapping the `{snapshot:…}` envelope; rides `Call` |
| `Subscribe(ctx, types, onEvent)` | long-lived events connection; blocks until drop/ctx |
| `DefaultSocketPath()` | `~/.config/herdr/herdr.sock` |
| `AllEventTypes` | global resource events subscribed by default |

# Persistent multiplexing

Post `bridge-ipc-hardening`, `Call` no longer dials per request. The client keeps one connection guarded by `mu` with a `pending map[string]chan call`:

- `Call` registers a waiter under `id`, writes the request under `mu`, then `select`s on the waiter channel vs `ctx.Done()` (per-call timeout via ctx — a shared conn can't set per-call deadlines).
- A single `readLoop` dispatches each reply to its waiter by `id`; unknown ids (already cancelled) are dropped.
- On read/write error `dropConn` closes the connection, fails all outstanding waiters (`errClosed`), and bumps a generation counter; the next `Call` lazily reconnects.
- `Subscribe` keeps its **own** connection (streaming, no id correlation).

Concurrency is exercised by `internal/herdr/mux_test.go` (correlation, out-of-order, ctx-cancel, reconnect).

# Citations

* [internal/herdr/client.go](/internal/herdr/client.go)
* [Herdr socket API](/references/herdr-socket-api.md)
