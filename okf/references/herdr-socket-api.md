---
type: Concept
title: Herdr socket API
description: External reference — Herdr's local socket protocol for scripts, tools, and agents
tags: [reference, herdr, socket, external]
resource: https://herdr.dev/docs/socket-api/
timestamp: 2026-09-03T00:00:00Z
---

# Summary

Herdr exposes a local Unix-socket API (newline-delimited JSON, id-correlated request/response, plus event subscriptions) at `~/.config/herdr/herdr.sock`. Herdr Web's [herdr.Client](/packages/herdr.md) is a minimal consumer of this surface; the [upstream service concept](/services/herdr-socket.md) lists the specific methods used.

Canonical concept model: **session → workspace → tab → pane → agent**, with agent states `working|blocked|done|idle|unknown`.

# Source

- Socket API: <https://herdr.dev/docs/socket-api/>
- Agent guide: <https://herdr.dev/agent-guide.md>
- The installed CLI can print the schema: `herdr api schema --json`.

# Citations

* <https://herdr.dev/docs/socket-api/>
