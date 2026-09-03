---
type: TypeScript Type
title: Session store
description: SessionModel (pure state) bridged to Svelte stores; applies flat snapshots and derives agents/rollups
tags: [state, svelte, store, model, frontend]
timestamp: 2026-09-03T00:00:00Z
---

# SessionModel

`web/src/lib/session/model.ts` — framework-free class holding `spaces` + `focus`. `apply(ev)` handles the flat `snapshot` event (`applySnapshot(ev)` reads `ev.spaces`/`ev.focus` directly) plus granular `workspace.*`/`tab.*`/`pane.*` patch events (present but unused — the Go backend only broadcasts full snapshots). Derived helpers: `agents()` (blocked-first sort), `rollup(spaceId)`, `findPane`.

# Store wiring

`web/src/lib/session/store.ts` — `createSession(transport)` subscribes the model to transport events, publishes `spaces`/`focus`/`connection` as Svelte readable stores, and exposes `request`. `derive.ts` holds pure selectors (`agentsOf`, `rollupOf`, `primaryPaneOf`, `glyph`).

`implements` the [session snapshot](/concepts/session-snapshot.md) concept; `consumes` a [Transport](/frontend/transport.md).

# Citations

* [web/src/lib/session/model.ts](/web/src/lib/session/model.ts)
* [web/src/lib/session/store.ts](/web/src/lib/session/store.ts)
* [web/src/lib/session/derive.ts](/web/src/lib/session/derive.ts)
