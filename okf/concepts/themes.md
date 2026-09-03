---
type: Concept
title: Themes & text size
description: CSS-variable token themes (incl. Solarized Light) and a document-zoom text-size control
tags: [theme, tokens, css, accessibility]
timestamp: 2026-09-03T00:00:00Z
---

# Theming

All colors are CSS custom properties in `web/src/lib/tokens.css`. Themes override the token set via `[data-theme='…']`: `herdr-dark` (default), `ash`, `gruvbox`, and `solarized-light` (the one light theme — a full palette override, with a couple of components made theme-agnostic so dark assumptions invert). `+layout.svelte` sets `data-theme` from the [config](/config/settings.md) store.

# Text size

`font_scale` (S/M/L/XL → 0.9/1/1.15/1.3) is applied as document `zoom`, scaling the whole UI. Persisted in the `[web]` config and set from Settings.

# Typography

UI font Geist; mono font **Fira Code** (contextual ligatures), self-hosted via `@fontsource` and embedded (no CDN).

# Citations

* [web/src/lib/tokens.css](/web/src/lib/tokens.css)
* [web/src/routes/settings/+page.svelte](/web/src/routes/settings/+page.svelte)
