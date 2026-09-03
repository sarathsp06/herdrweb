---
type: UI Route
title: SvelteKit Routes
description: The app's screens — inbox, pane terminal, diff, spaces, space detail, settings
tags: [routes, sveltekit, ui, frontend]
timestamp: 2026-09-03T00:00:00Z
---

# Routes (`web/src/routes`)

| Path | Screen | Notes |
|---|---|---|
| `/` | Inbox / home | desktop redirects to the last/first agent pane; mobile shows the embedded sidebar inbox |
| `/pane/[id]` | Pane terminal | raw scrollback (`pane.read`, `recent_unwrapped`) + [Composer](/frontend/composer.md); soft-wraps < 880px; bottom-pinned autoscroll |
| `/pane/[id]/diff` | Diff viewer | Shiki-highlighted unified diff (fixture-backed) |
| `/spaces` | Spaces list | cards; body opens chat, "Tabs" opens detail |
| `/spaces/[id]` | Space detail | tab strip + pane cards; add-tab / split-pane via the confirm sheet |
| `/settings` | Settings | theme picker, text-size, behaviour toggles, server card |

The root `+layout.svelte` is a column: a persistent [breadcrumb bar](/frontend/navigation.md) over scrolling content, with the sidebar as a push (desktop) / drawer (mobile). Every mutating action routes through a [BottomSheet](/frontend/navigation.md) confirmation.

# Citations

* [web/src/routes](/web/src/routes)
