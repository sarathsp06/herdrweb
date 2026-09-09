---
type: UI Component
title: Navigation & Layout
description: Bottom tab bar shell, per-screen headers, confirm/action sheet, theme + text-size application
tags: [navigation, layout, tab-bar, sidebar, theme, frontend]
timestamp: 2026-09-03T00:00:00Z
---

# Bottom tab bar (`screens/BottomNav.svelte`)

Primary phone navigation: a fixed, safe-area-aware tab bar (Agents · Spaces · Settings) with a blocked-agent count badge on Agents. Hidden on `/pane/*` (the pane is a fullscreen push; back lives in `PaneHeader`) and on desktop (≥ 880px), where the persistent sidebar takes over.

# Pane header (`screens/PaneHeader.svelte`)

Compact chrome on `/pane/*`: 44px back chevron, pane title + space sub-label (tap opens a tab-switcher bottom sheet), status pill, and a `⋯` overflow sheet for pane actions (diff, rename, close).

# Sidebar (`screens/Sidebar.svelte`)

Desktop-only inbox column: spaces + agents (blocked-first). On phones the inbox is the `/` route itself (urgency sections + space chips); there is no drawer.

# BottomSheet (`ui/BottomSheet.svelte`)

Two modes: a confirmation barrier for every mutating action (create/rename/close workspace/tab/pane, split) — nothing mutates on a single tap — and an action-list mode (tap-to-pick rows) used for the pane tab switcher, pane `⋯` actions, and space card overflow.

# Theme & text size

`+layout.svelte` applies `data-theme` and a document `zoom` from the [config](/config/settings.md) store, and keeps the `theme-color` meta in live sync with the active theme so OS/browser chrome tracks it immediately (the `apple-mobile-web-app-status-bar-style` meta is also written but only takes visible effect on an installed PWA's next cold launch — see [Theming](/concepts/themes.md)). Themes: `herdr-dark`, `gruvbox`, `solarized-light`, `paper` — full palettes in `web/src/lib/tokens.css`. Mono font is Fira Code.

# Citations

* [web/src/routes/+layout.svelte](/web/src/routes/+layout.svelte)
* [web/src/lib/screens/BottomNav.svelte](/web/src/lib/screens/BottomNav.svelte)
* [web/src/lib/screens/PaneHeader.svelte](/web/src/lib/screens/PaneHeader.svelte)
* [web/src/lib/tokens.css](/web/src/lib/tokens.css)
