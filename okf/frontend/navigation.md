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

Compact chrome on `/pane/*`: 44px back chevron on phones, pane title + space sub-label (tap opens a tab-switcher bottom sheet), status pill, and a `⋯` overflow sheet for pane actions (diff, rename, close). On desktop with the sidebar collapsed, the chevron's place holds `☰`, which shows the sidebar over the terminal.

# Sidebar (`screens/Sidebar.svelte`)

Desktop-only inbox column: spaces + agents (blocked-first). On phones the inbox is the `/` route itself (urgency sections + space chips); there is no drawer.

On `/pane/*` the header's **hide** sets `sidebarCollapsed` (`ui/state.ts`, persisted in `localStorage`), and the terminal takes the full width. Collapse applies only to `/pane/*`: other routes keep the sidebar docked, so they never lose navigation. While collapsed, `sidebarPeek` renders the sidebar over the pane (absolute inside the app shell, not `fixed`, because an installed iOS app lays out fixed elements against a shorter viewport); a backdrop tap or any navigation closes it, and its **pin** clears the collapse.

# BottomSheet (`ui/BottomSheet.svelte`)

Two modes: a confirmation barrier for every mutating action (create/rename/close workspace/tab/pane, split) — nothing mutates on a single tap — and an action-list mode (tap-to-pick rows) used for the pane tab switcher, pane `⋯` actions, and space card overflow.

# Theme & text size

`+layout.svelte` applies `data-theme` and a document `zoom` from the [config](/config/settings.md) store, and keeps the `theme-color` meta in live sync with the active theme so OS/browser chrome tracks it immediately (the `apple-mobile-web-app-status-bar-style` meta is also written but only takes visible effect on an installed PWA's next cold launch — see [Theming](/concepts/themes.md)). Themes: `herdr-dark`, `gruvbox`, `solarized-light`, `paper` — full palettes in `web/src/lib/tokens.css`. Mono font is Fira Code.

# Citations

* [web/src/routes/+layout.svelte](/web/src/routes/+layout.svelte)
* [web/src/lib/screens/BottomNav.svelte](/web/src/lib/screens/BottomNav.svelte)
* [web/src/lib/screens/PaneHeader.svelte](/web/src/lib/screens/PaneHeader.svelte)
* [web/src/lib/screens/Sidebar.svelte](/web/src/lib/screens/Sidebar.svelte)
* [web/src/lib/ui/state.ts](/web/src/lib/ui/state.ts)
* [web/src/lib/tokens.css](/web/src/lib/tokens.css)
