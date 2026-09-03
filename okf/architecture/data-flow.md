---
type: Concept
title: Data Flow
description: Snapshot fan-out (Herdr → browser) and id-correlated call pass-through (browser → Herdr)
tags: [architecture, data-flow, websocket, snapshot]
timestamp: 2026-09-03T00:00:00Z
---

# Inbound: Herdr → browser (snapshot fan-out)

[server.Hub.Run](/packages/server.md) maintains one Herdr connection and a normalized snapshot cache:

1. Bootstrap: `refresh` pulls `session.snapshot`, [normalizes](/packages/protocol.md) it, caches the JSON, and broadcasts.
2. Live: `Client.Subscribe` streams Herdr resource events; each event calls `markDirty`. A 1.5s `poller` also marks dirty (Herdr emits no global agent-status event).
3. `debouncer` coalesces dirty signals over 120ms, then `refresh` re-snapshots and `broadcast`s the fresh normalized snapshot to every connected browser.
4. New browsers get the cached snapshot immediately on connect (`addBrowser`).

The wire message is flat: `{"type":"snapshot","spaces":[…],"focus":{…}}`. The frontend [SessionModel](/concepts/session-snapshot.md) replaces its state from it.

# Outbound: browser → Herdr (call pass-through)

Browser RPCs arrive as `{id, method, params}` frames on `/ws`. [handleWS](/packages/server.md) acquires a per-connection semaphore token (bounded concurrency), then `handleCall` forwards the method to [herdr.Client.Call](/packages/herdr.md) and returns `{id, result}` or `{id, error}` on the same socket. The browser [SocketTransport](/frontend/transport.md) correlates replies by `id`.

Methods used by the UI: `pane.read` (raw scrollback), `agent.prompt`, `agent.send_keys`, `workspace.*`, `tab.*`, `pane.*`, `server.reload_config`.

# Resilience

- Herdr connection drop → `Hub.Run` reconnects with backoff and self-heals via a fresh snapshot.
- RPC connection drop → `herdr.Client` fails in-flight calls and lazily reconnects on the next call.
- Browser send races (disconnect mid-reply) are guarded by a per-browser mutex + `closed` flag in [server.Hub](/packages/server.md).

# Citations

* [internal/server/server.go](/internal/server/server.go)
* [internal/herdr/client.go](/internal/herdr/client.go)
