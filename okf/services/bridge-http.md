---
type: HTTP Service
title: Bridge HTTP/WS
description: The loopback surface the browser uses — WebSocket fan-out/RPC plus config and health REST
tags: [http, websocket, api, service]
resource: internal/server/server.go
timestamp: 2026-09-03T00:00:00Z
---

# Endpoints (`Hub.Handler`)

| Method | Path | Purpose |
|---|---|---|
| WS | `/ws` | receive `{type:"snapshot",…}` broadcasts; send `{id,method,params}` RPCs, receive `{id,result|error}` |
| GET | `/api/config` | current `[web]` settings |
| PUT/POST | `/api/config` | persist settings, then `server.reload_config` |
| GET | `/api/health` | `{ok, herdr:<state>, version, socket}` |
| GET | `/` (+ assets) | embedded SPA with `index.html` fallback |

# WebSocket RPC

The browser [SocketTransport](/frontend/transport.md) opens one WebSocket; RPCs are id-correlated. `handleCall` `delegates to` [herdr.Client.Call](/packages/herdr.md); concurrency is bounded per connection (`maxInflightPerConn`). Only the UI's own methods (`browserMethods`, mirroring the web `Call` union) pass through to Herdr - see the [upstream socket](/services/herdr-socket.md); others get an error reply.

Bound to `127.0.0.1:7331` by default, no auth (single operator). Every route sits behind `hostGuard`: a Host that is not an IP literal, `localhost`, `*.ts.net` or an `-allow-host` name gets 403 (DNS rebinding), as does an `Origin` that is neither the request's own host nor loopback (cross-site WebSocket). See [main flags](/packages/cmd-herdr-bridge.md).

# Citations

* [internal/server/server.go](/internal/server/server.go)
