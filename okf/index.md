---
okf_version: "0.1"
---

# Herdr Web

A single-operator web client for [Herdr](https://herdr.dev), the terminal/agent multiplexer. The organizing idea: **agents, not terminals, are the primary objects** — the default screen is an inbox of every coding agent across your spaces, sorted so the ones that need you (`blocked`) surface first.

It ships as **one self-contained Go binary** (`herdr-bridge`): a SvelteKit SPA embedded via `go:embed` into a Go bridge that owns connections to the Herdr Unix socket (`~/.config/herdr/herdr.sock`) and fans live session data to the browser over WebSocket on loopback (`127.0.0.1:7331`).

```
browser (SvelteKit)  ⇄  Go bridge (herdr-bridge)  ⇄  Herdr socket
        WebSocket + embedded static assets, one loopback origin
```

# Knowledge

* [Architecture](/architecture/index.md) - system shape, data flow, and key decisions
* [Packages](/packages/index.md) - Go backend package reference (`cmd/`, `internal/`)
* [Frontend](/frontend/index.md) - SvelteKit UI (`web/src/`): transport, routes, components
* [Services](/services/index.md) - the bridge's HTTP/WS surface and the upstream Herdr socket client
* [Concepts](/concepts/index.md) - snapshot model, terminal view, themes
* [Configuration](/config/index.md) - CLI flags and the `[web]` config.toml table
* [References](/references/index.md) - external APIs and design docs

# Citations

* [README.md](/README.md)
* [docs/ARCHITECTURE_RECOMMENDATIONS.md](/docs/ARCHITECTURE_RECOMMENDATIONS.md)
