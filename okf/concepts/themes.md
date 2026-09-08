---
type: Concept
title: Themes & text size
description: CSS-variable token themes (incl. Solarized Light) and a document-zoom text-size control
tags: [theme, tokens, css, accessibility]
timestamp: 2026-09-03T00:00:00Z
---

# Theming

All colors are CSS custom properties in `web/src/lib/tokens.css`. Themes override the token set via `[data-theme='…']`: `herdr-dark` (default), `gruvbox` (alternate dark accents), `solarized-light`, and `paper` (a pure-white light theme) — the two light themes are full palette overrides, with a couple of components made theme-agnostic so dark assumptions invert. Each light theme sets `color-scheme: light`; `:root` pins `color-scheme: dark` so native controls in the dark themes don't follow a light OS preference even though the page's `<meta name="color-scheme">` advertises both. `+layout.svelte` sets `data-theme` from the [config](/config/settings.md) store.

OS/browser chrome (the `theme-color` meta and, for an installed PWA, the iOS status-bar style) tracks the active theme: `app.html` has an inline pre-paint script that reads the persisted theme from `localStorage` and sets both before first paint (killing the flash-of-herdr-dark and the stuck-dark-chrome bug), and `+layout.svelte`'s effect keeps them in sync on live theme switches. `manifest.webmanifest`'s own `theme_color`/`background_color` (the OS splash/task-switcher colour for a cold PWA launch) is a static file read once by the OS before any JS runs, so it always matches `herdr-dark` regardless of the persisted theme — a known, accepted gap.

# Text size

`font_scale` (S/M/L/XL → 0.9/1/1.15/1.3) is applied as document `zoom`, scaling the whole UI. Persisted in the `[web]` config and set from Settings.

# Typography

UI font Geist; mono font **Fira Code** (contextual ligatures), self-hosted via `@fontsource` and embedded (no CDN).

# Citations

* [web/src/lib/tokens.css](/web/src/lib/tokens.css)
* [web/src/routes/settings/+page.svelte](/web/src/routes/settings/+page.svelte)
* [web/src/app.html](/web/src/app.html)
* [web/src/routes/+layout.svelte](/web/src/routes/+layout.svelte)
