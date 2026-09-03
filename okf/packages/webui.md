---
type: Go Package
title: internal/webui
description: go:embed of the built SvelteKit assets with SPA index.html fallback routing
tags: [embed, spa, static, assets]
timestamp: 2026-09-03T00:00:00Z
---

# Responsibilities

Embeds `internal/webui/dist/` (the compiled SvelteKit output, produced by `make web`) and returns an `http.Handler` that serves static assets and falls back to `index.html` for client-side history routes. Mounted at `/` by [server.Handler](/packages/server.md).

# Build coupling

`make web` runs the SvelteKit build and copies `web/build/` → `internal/webui/dist/`, then `go build` embeds it. `dist/.gitkeep` keeps the directory tracked; the built assets are gitignored.

# Citations

* [internal/webui/webui.go](/internal/webui/webui.go)
* [Makefile](/Makefile)
