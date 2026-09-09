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
| `/` | Inbox / home | desktop redirects to the last/first agent pane; mobile shows agents triaged by urgency (needs you → working → idle/done) under a space-chip strip |
| `/pane/[id]` | Pane terminal | raw scrollback (`pane.read`, `recent_unwrapped`) + [Composer](/frontend/composer.md); soft-wraps < 880px; bottom-pinned autoscroll |
| `/pane/[id]/diff` | Diff viewer | Shiki-highlighted unified diff (fixture-backed) |
| `/spaces` | Spaces list | cards; tap opens chat, `⋯` overflow sheet for tabs/rename/close |
| `/spaces/[id]` | Space detail | tab strip + pane cards; add-tab / split-pane via the confirm sheet |
| `/settings` | Settings | theme picker, text-size, behaviour toggles, server card |

The root `+layout.svelte` is a column of scrolling content over a fixed [bottom tab bar](/frontend/navigation.md) on phones (hidden on `/pane/*`); desktop (≥ 880px) swaps it for a persistent sidebar. Every mutating action routes through a [BottomSheet](/frontend/navigation.md) confirmation.

# Citations

* [web/src/routes](/web/src/routes)
