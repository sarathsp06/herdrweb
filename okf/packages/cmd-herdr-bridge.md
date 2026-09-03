---
type: Go Package
title: cmd/herdr-bridge
description: Binary entry point — parses flags, wires the Hub + client, serves UI/ws/api with graceful shutdown
tags: [main, cli, lifecycle]
timestamp: 2026-09-03T00:00:00Z
---

# Responsibilities

`main` builds a [herdr.Client](/packages/herdr.md) and a [server.Hub](/packages/server.md), starts `hub.Run` in the background, and serves `hub.Handler()` over `net/http` with `signal.NotifyContext` for graceful shutdown.

# Flags

| Flag | Default | Purpose |
|---|---|---|
| `-addr` | `127.0.0.1:7331` | listen address (loopback only by default) |
| `-socket` | `~/.config/herdr/herdr.sock` | path to the Herdr socket |
| `-config` | `~/.config/herdr/config.toml` | path to Herdr config.toml |
| `-version` | — | print version and exit |

`version` is injected at release via `-ldflags -X main.version=…`.

# Notes

- Bind to a Tailscale IP (`-addr $(tailscale ip -4):7331`) to expose over a tailnet; there is no auth (single operator on loopback by design).
- `configures` the [HTTP service](/services/bridge-http.md).

# Citations

* [cmd/herdr-bridge/main.go](/cmd/herdr-bridge/main.go)
