---
type: TypeScript Type
title: Transport
description: Transport abstraction — SocketTransport (live WebSocket) and FixtureTransport (mock dataset)
tags: [transport, websocket, fixtures, frontend]
timestamp: 2026-09-03T00:00:00Z
---

# Interface

`Transport` (in `web/src/lib/transport`): `subscribe(handler)` for session events, `onConnection(handler)` for `ConnState`, `request(call)` for id-correlated RPCs, `start()`/`stop()`.

# SocketTransport

`consumes` the bridge `/ws`. Opens a WebSocket, reconnects with exponential backoff. Inbound frames with an `id` resolve pending `request` promises (`c1`,`c2`,…); frames with a `type` are session events (the flat `snapshot`) dispatched to subscribers. `request` sends `{id, method, params}`.

# FixtureTransport

Selected by `?fixtures=1` (and during SSR/e2e/screenshots). Emits `{type:'snapshot', ...FIXTURE_SNAPSHOT}` and answers `pane.read` from each pane's `tail`. Backs deterministic first paint and Playwright screenshots without a live daemon.

`pickTransport()` chooses fixture vs socket; `session()` builds the singleton and `start()`s it.

# Citations

* [web/src/lib/transport/socket.ts](/web/src/lib/transport/socket.ts)
* [web/src/lib/transport/fixture.ts](/web/src/lib/transport/fixture.ts)
* [web/src/lib/session/live.ts](/web/src/lib/session/live.ts)
