---
type: Go Package
title: internal/config
description: Read/write the UI-owned [web] table inside Herdr's config.toml, preserving foreign tables
tags: [config, toml, settings]
timestamp: 2026-09-03T00:00:00Z
---

# Responsibilities

Persists UI preferences under a `[web]` table in `~/.config/herdr/config.toml`. `Load` falls back to `Default()` for missing keys; `Save` re-encodes the merged document so other tables (`[keys]`, `[theme]`, `[ui]`) survive, writing atomically via a temp file + `os.Rename`.

# Settings

| Field (toml / json) | Type | Default |
|---|---|---|
| `theme` | string | `herdr-dark` |
| `notify` | bool | `true` |
| `follow` | bool | `true` |
| `ansi` | bool | `true` |
| `dev_captions` / `devCaptions` | bool | `false` |
| `font_scale` / `fontScale` | float | `1` |

See the [settings reference](/config/settings.md). Written by [server.handleConfig](/packages/server.md), which then calls `server.reload_config`.

# Citations

* [internal/config/config.go](/internal/config/config.go)
