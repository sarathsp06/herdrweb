---
type: API Endpoint
title: Herdr socket (upstream)
description: The Herdr daemon's Unix-socket JSON-RPC surface the bridge consumes
tags: [herdr, socket, json-rpc, upstream]
resource: https://herdr.dev/docs/socket-api/
timestamp: 2026-09-03T00:00:00Z
---

# Protocol

Newline-delimited JSON over `~/.config/herdr/herdr.sock`, id-correlated request/response, plus `events.subscribe` streaming. The bridge's [herdr.Client](/packages/herdr.md) is the only consumer.

# Methods used by Herdr Web

| Method | Use |
|---|---|
| `session.snapshot` | bootstrap + every refresh (fanned to browsers) |
| `events.subscribe` | live resource events → mark cache dirty |
| `pane.read` | raw terminal scrollback (`source: recent_unwrapped`) |
| `agent.prompt` | submit a prompt to an agent pane |
| `agent.send_keys` | send key tokens (arrows/enter/esc/ctrl+c, y/n approvals) |
| `workspace.* / tab.* / pane.*` | create/rename/close/split/focus |
| `server.reload_config` | applied after a settings write |

Full protocol version is Herdr-owned (observed protocol 20 / v0.8.2). See the [reference](/references/herdr-socket-api.md).

# Citations

* [Herdr socket API](/references/herdr-socket-api.md)
