---
type: Concept
title: Terminal view
description: Live agent panes render raw terminal scrollback; there is no chat-transcript parser
tags: [terminal, scrollback, pane, agent]
timestamp: 2026-09-03T00:00:00Z
---

# Decision

Against a live Herdr instance every agent pane is **raw terminal scrollback** via `pane.read` (`source: recent_unwrapped`). An early mock had chat-bubble/reasoning/tool-call components, but no structured transcript feed exists, so that layer was removed — the ground truth is the terminal.

# Behaviour

- The pane view fetches `pane.read` and re-fetches as the snapshot changes (live tailing), with **bottom-pinned autoscroll** (scrolling up to read history is not yanked down).
- Lines never wrap (`white-space: pre`): the `use:fitToWidth` action instead shrinks the monospace font size so a base-14px-wide line still fits the viewport, keeping true terminal columns readable on any phone without reflowing content the bridge didn't render at that width. Content still too wide at the floor size scrolls horizontally.
- Operator control is the [Composer](/frontend/composer.md): `agent.prompt` for text, `agent.send_keys` for arrows/enter/esc/ctrl+c and y/n approvals. Agent panes also get a **◎ direct control** toggle: while on, swiping the transcript sends one arrow key per swipe via `agent.send_keys` instead of scrolling — a faster way to drive an agent's own interactive picker (e.g. `/model`) than tapping the key row. It resets off on every pane switch and drops automatically if the pane stops being an agent mid-session.

# Citations

* [web/src/routes/pane/[id]/+page.svelte](/web/src/routes/pane/%5Bid%5D/+page.svelte)
* [docs/ARCHITECTURE_RECOMMENDATIONS.md](/docs/ARCHITECTURE_RECOMMENDATIONS.md)
* [web/src/lib/layout/fit.ts](/web/src/lib/layout/fit.ts)
* [web/src/lib/layout/swipe.ts](/web/src/lib/layout/swipe.ts)
