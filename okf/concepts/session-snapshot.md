---
type: Concept
title: Session snapshot
description: The normalized space→tab→pane→agent model broadcast whole to browsers
tags: [snapshot, model, spaces, agents]
timestamp: 2026-09-03T00:00:00Z
---

# What it is

The single unit of state the bridge shares. Herdr's raw `session.snapshot` is [normalized](/packages/protocol.md) into `Snapshot{ spaces, focus }` and broadcast **in full** on every change (coarse re-snapshotting — no deltas on the wire).

# Shape

- **Space** (Herdr workspace) → **Tabs** → **Panes**. A **Pane** has `status` (`working|blocked|done|idle|unknown`), an `agent` flag, and a `tail` of recent output.
- `focus` records the focused space/tab/pane ids.
- Wire form is flat: `{"type":"snapshot","spaces":[…],"focus":{…}}` — the frontend [SessionModel](/frontend/session-store.md) replaces state from it.

# Lifecycle

`produced` by [server.Hub](/packages/server.md) (subscribe + poll → debounce → normalize → broadcast); `consumed` by the [SessionModel](/frontend/session-store.md); rendered by the [inbox/sidebar](/frontend/navigation.md) with blocked-first sorting and per-space status rollups.

# Citations

* [internal/protocol/protocol.go](/internal/protocol/protocol.go)
* [web/src/lib/session/model.ts](/web/src/lib/session/model.ts)
