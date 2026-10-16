---
type: Go Package
title: internal/config
description: Read/write the UI-owned settings file (herdrweb-settings.json) and migrate the old [web] table out of Herdr's config.toml
tags: [config, json, toml, settings]
timestamp: 2026-09-28T00:00:00Z
---

# Responsibilities

Persists UI preferences as JSON in `FileName` (`herdrweb-settings.json`), in the directory of Herdr's `config.toml`. `Load` falls back to `Default()` for a missing file or missing keys; `Save` writes atomically via a temp file + `os.Rename`.

`Migrate(herdrConfig, settingsPath)` moves the `[web]` table (and any `[web.*]` sub-tables) that earlier versions wrote into Herdr's `config.toml`: it parses the file first and gives up on a parse error, copies `[web]` into the settings file unless that already exists, then removes the table's lines textually (`cutTable`) so every other line stays byte for byte. It writes through a symlinked `config.toml` rather than replacing the link.

# Settings

| Field (json / legacy toml) | Type | Default |
|---|---|---|
| `theme` | string | `herdr-dark` |
| `notify` | bool | `true` |
| `follow` | bool | `true` |
| `ansi` | bool | `true` |
| `devCaptions` / `dev_captions` | bool | `false` |
| `fontScale` / `font_scale` | float | `1` |

See the [settings reference](/config/settings.md). Written by [server.handleConfig](/packages/server.md); `Migrate` runs once at startup from `cmd/herdr-bridge`.

# Citations

* [internal/config/config.go](/internal/config/config.go)
