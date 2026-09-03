---
type: UI Component
title: Composer.svelte
description: Prompt input plus terminal nav keys; dispatches agent.prompt / agent.send_keys
tags: [composer, input, agent, keys, frontend]
timestamp: 2026-09-03T00:00:00Z
---

# Behaviour

`web/src/lib/chat/Composer.svelte` — the pane's input row.

- **Send**: Enter submits (Shift+Enter = newline). Clears the draft immediately and fires `agent.prompt` non-blocking (`void request(...)`) — the prompt's `wait` resolves only when the agent next goes idle/blocked, so the UI must not await it.
- **Nav keys**: a row of `↑ ↓ ← → ⏎ esc ⌃C` sends `agent.send_keys` with a single-element `keys` array (Herdr tokens `up/down/left/right/enter/esc/ctrl+c`) — this is how the operator drives TUI selection menus and answers blocked prompts.

`routes to` [herdr.Client](/packages/herdr.md) via the [SocketTransport](/frontend/transport.md). Covered by `web/src/lib/chat/Composer.test.ts`.

# Citations

* [web/src/lib/chat/Composer.svelte](/web/src/lib/chat/Composer.svelte)
