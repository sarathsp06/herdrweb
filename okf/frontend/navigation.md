---
type: UI Component
title: Navigation & Layout
description: Persistent breadcrumb bar, toggleable sidebar/drawer, confirm sheet, theme + text-size application
tags: [navigation, layout, breadcrumbs, sidebar, theme, frontend]
timestamp: 2026-09-03T00:00:00Z
---

# Breadcrumbs bar (`screens/Breadcrumbs.svelte`)

Always visible on every route (including fullscreen panes). Holds the `☰` nav toggle, a clickable path (`⌂ / space / pane`), the pane status pill, and the `⚙` settings gear. Replaced the bottom TabBar (which vanished on the pane screen).

# Sidebar (`screens/Sidebar.svelte`)

The inbox: spaces + agents (blocked-first). A push column on desktop, a slide-in drawer + backdrop on mobile, toggled via `navOpen`. Selecting a space opens its primary agent pane; a per-space `tabs` button opens the space detail. Footer nav links Agents/Spaces/Settings.

# BottomSheet (`ui/BottomSheet.svelte`)

Confirmation barrier for every mutating action (create/rename/close workspace/tab/pane, split) — nothing mutates on a single tap.

# Theme & text size

`+layout.svelte` applies `data-theme` and a document `zoom` from the [config](/config/settings.md) store. Themes: `herdr-dark`, `ash`, `gruvbox`, `solarized-light` — full palettes in `web/src/lib/tokens.css`. Mono font is Fira Code.

# Citations

* [web/src/routes/+layout.svelte](/web/src/routes/+layout.svelte)
* [web/src/lib/screens/Breadcrumbs.svelte](/web/src/lib/screens/Breadcrumbs.svelte)
* [web/src/lib/tokens.css](/web/src/lib/tokens.css)
