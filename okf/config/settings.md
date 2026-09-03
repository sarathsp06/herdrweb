---
type: Configuration
title: Settings
description: CLI flags for the bridge and the UI-owned [web] table persisted in Herdr's config.toml
tags: [config, flags, toml, settings]
timestamp: 2026-09-03T00:00:00Z
---

# CLI flags (`herdr-bridge`)

| Flag | Default |
|---|---|
| `-addr` | `127.0.0.1:7331` |
| `-socket` | `~/.config/herdr/herdr.sock` |
| `-config` | `~/.config/herdr/config.toml` |
| `-version` | print and exit |

# `[web]` table (config.toml)

Written by the bridge on a settings save (then `server.reload_config`), and mirrored in the browser's `localStorage`. Managed by [internal/config](/packages/config.md).

```toml
[web]
theme = "herdr-dark"   # herdr-dark | ash | gruvbox | solarized-light
notify = true          # push when an agent is blocked
follow = true          # follow the focused pane
ansi = true            # keep ANSI colours in raw mode
dev_captions = false   # show socket-call captions (developer)
font_scale = 1.0       # UI text-size multiplier (document zoom)
```

Other tables in the file (`[keys]`, `[theme]`, `[ui]`) are preserved on write.

# Citations

* [cmd/herdr-bridge/main.go](/cmd/herdr-bridge/main.go)
* [internal/config/config.go](/internal/config/config.go)
