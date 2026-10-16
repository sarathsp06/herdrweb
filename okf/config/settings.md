---
type: Configuration
title: Settings
description: CLI flags for the bridge and the UI-owned settings file (herdrweb-settings.json) beside Herdr's config.toml
tags: [config, flags, json, settings]
timestamp: 2026-09-28T00:00:00Z
---

# CLI flags (`herdrweb`)

| Flag | Default |
|---|---|
| `-addr` | `127.0.0.1:7331` |
| `-socket` | `~/.config/herdr/herdr.sock` |
| `-config` | `~/.config/herdr/config.toml`; the bridge keeps its own files in the same directory |
| `-version` | print and exit |

# Settings file (`herdrweb-settings.json`)

Lives in the directory of `-config`, next to the push keys (`webpush.json`) and subscriptions (`push-subs.json`). Written by the bridge on a settings save and mirrored in the browser's `localStorage`; Herdr never reads it. Managed by [internal/config](/packages/config.md).

```json
{
  "theme": "herdr-dark",
  "notify": true,
  "follow": true,
  "ansi": true,
  "devCaptions": false,
  "fontScale": 1
}
```

`theme` is `herdr-dark` | `gruvbox` | `solarized-light` | `paper`; `notify` pushes when an agent needs you; `follow` follows the focused pane; `ansi` keeps ANSI colours in raw mode; `devCaptions` shows socket-call captions; `fontScale` is the UI text-size multiplier (document zoom).

# Migration from `[web]`

Earlier versions kept these settings in a `[web]` table inside Herdr's `config.toml`, which Herdr reports as `unknown config section [web]`. On startup `config.Migrate` copies that table into the settings file (unless one already exists) and deletes only the table's lines, leaving comments, order and formatting of the rest untouched. A `config.toml` that does not parse is left alone.

# Citations

* [cmd/herdr-bridge/main.go](/cmd/herdr-bridge/main.go)
* [internal/config/config.go](/internal/config/config.go)
