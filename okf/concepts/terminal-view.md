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
- Long lines **soft-wrap under 880px** (a desktop-width terminal can't be reflowed to a phone — the bridge can't set Herdr's PTY width), and keep true columns with horizontal scroll on wider screens.
- Operator control is the [Composer](/frontend/composer.md): `agent.prompt` for text, `agent.send_keys` for arrows/enter/esc/ctrl+c and y/n approvals.

# Citations

* [web/src/routes/pane/[id]/+page.svelte](/web/src/routes/pane/%5Bid%5D/+page.svelte)
* [docs/ARCHITECTURE_RECOMMENDATIONS.md](/docs/ARCHITECTURE_RECOMMENDATIONS.md)
