---
type: Go Package
title: internal/server
description: Hub — owns one Herdr connection, caches a normalized snapshot, fans it to browsers over WebSocket, and serves REST endpoints
tags: [hub, websocket, broadcast, http]
timestamp: 2026-09-03T00:00:00Z
---

# Responsibilities

`Hub` is the coordinator. It `depends on` [herdr.Client](/packages/herdr.md), `produces` normalized snapshots via [protocol.Normalize](/packages/protocol.md), and `manages` connected browsers.

# Lifecycle (`Hub.Run`)

Bootstrap snapshot → `Subscribe` to Herdr events (each `markDirty`) → 1.5s `poller` also marks dirty → `debouncer` coalesces over 120ms → `refresh` re-snapshots, caches JSON, `broadcast`s to browsers. Reconnects to Herdr with backoff on drop. See [data flow](/architecture/data-flow.md).

# WebSocket (`handleWS` / `handleCall`)

- Each `/ws` connection gets a write-pump goroutine draining `browser.send`, plus a read loop.
- `const maxInflightPerConn = 8`: the read loop acquires a semaphore token via `dispatch` before spawning `handleCall`, applying per-connection backpressure (no unbounded goroutines).
- `handleCall` forwards `{id,method,params}` to `Client.Call` and replies `{id,result|error}` via `trySend`.
- `browser` has `mu` + `closed` so an in-flight reply can't `send on closed channel` when the read loop removes the browser (`internal/server/browser_test.go`). Concurrency bound covered by `internal/server/server_test.go`.

# REST

- `GET/PUT /api/config` — read/write the [web] settings; a write persists then calls `server.reload_config`.
- `GET /api/health` — `{ok, herdr:<state>, version, socket}`.

`Handler()` wires `/ws`, `/api/config`, `/api/health`, and `/` → [webui](/packages/webui.md). See the [HTTP service](/services/bridge-http.md).

# Citations

* [internal/server/server.go](/internal/server/server.go)
