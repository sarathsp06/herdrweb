# Implementation prompt — Herdr Web (personal client)

Paste this to Claude Code (or any coding agent) in the repo you want it built in, with
`design_handoff_herdr_web/` present in the working directory.

---

## Task

Implement the Herdr Web personal client from the design references in
`design_handoff_herdr_web/`.

Read `design_handoff_herdr_web/README.md` first — it is the specification. Open
`design_handoff_herdr_web/herdr-web-standalone.html` in a browser and click through every screen
before writing code; the interaction details matter more than the markup. `Herdr Mobile.dc.html`
is the design source if you need an exact value.

The HTML files are **design references, not code to copy.** Recreate them in this repo's existing
stack, using its established component conventions, state management, and styling approach. Match
the visuals closely — this is a high-fidelity spec, so colors, type, spacing, radii and states should
land on the documented values. Do not import the prototype's inline-style approach; translate it to
whatever this codebase uses.

## What this is

A single-operator web client for Herdr, a terminal/agent multiplexer. The organizing idea:
**agents, not terminals, are the primary objects.** The default screen is an inbox of agents sorted
by who needs the human. Raw terminal scrollback exists but is secondary.

## Stack

Target: SvelteKit front end talking to a Go bridge, which talks to the Herdr socket at
`~/.config/herdr/herdr.sock`. If this repo already has a front end, use it and ignore that
suggestion. If the repo is empty, set up SvelteKit + TypeScript and say so before scaffolding.

## Build order

Work in these steps. **Stop after each step, run it, and show me the result before continuing.**

### 1. Socket layer and session store
No UI yet. Build the transport and the store the whole app reads from.

- Connect through the bridge; expose connection state (`connecting | open | reconnecting | closed`).
- Model the data as: `Space { id, label, cwd, branch, worktree, tabs[] }`,
  `Tab { id, label, panes[] }`,
  `Pane { id, label, sub, status, agent, tail }`.
- `status` is a 5-value enum: `working | blocked | done | idle | unknown`. It drives every dot,
  status word, rollup and sort order in the UI, so make it a real type, not a string.
- Bootstrap from a session snapshot, then live-patch from events:
  `workspace.updated`, `tab.updated`, `pane.updated`, `pane.agent_status_changed`,
  `pane.output_matched`.
- Derive space rollup status: blocked if any agent pane is blocked, else working if any is working,
  else idle if it has agent panes, else none.
- Calls to implement: `pane.read { source: "recent-unwrapped", lines: 80 }`,
  `agent.prompt { target, text, wait: { until: ["idle","blocked"], timeout_ms: 900000 } }`,
  `agent.send_keys`, `workspace.create/rename/close`, `tab.create/rename/close`,
  `pane.split/rename/close`, `server.reload_config`.
- Reconnect with backoff; re-snapshot on reconnect rather than replaying missed events.

Write it with the transport behind an interface and a fixture implementation, so the UI can be built
and tested against the README's mock data (four spaces: `hedr-web`, `api`, `api/billing` worktree,
`dotfiles`) before the bridge is ready.

### 2. Tokens and primitives
Put the README's color, type, spacing and radius tokens into this repo's normal styling mechanism.
Self-host Geist and Geist Mono — do not ship a Google Fonts link. Then build:
status dot, status pill, mono id/badge chip, card, hairline-divided button row, bottom sheet, toast,
and the three keyframe animations (`hpulse`, `hsheet`, `hbounce`).

Rules that apply everywhere: dark only; UI text in Geist, every identifier / id / number / code /
terminal string in Geist Mono; touch targets never below 44px; `text-wrap: pretty` on prose;
scrollbars hidden but containers still scroll.

### 3. Chat screen — `/pane/:id`
The largest piece. Implement each transcript message kind as its own component: user prompt bubble,
collapsible reasoning, agent text with copy/retry actions, collapsible tool call, code block with
line-number gutter and copy button, diff summary card, and the **blocked approval card** — that last
one is the highest-value element in the product, so get its prominence right.

Also: the `chat / raw` mode switch, raw scrollback (colored lines, no wrap, preceded by the
`pane.read` caption), the three-dot working indicator, and the composer — quick-send chips that
change when the pane is blocked (`y n esc ctrl+c /status`) vs not
(`continue / run tests / git diff / esc / ctrl+c`), auto-growing textarea capped at 96px, and a send
button that only goes solid once there's a draft.

Auto-scroll to bottom on pane change, mode change, and transcript mutation — after paint, not on a
timer chain like the prototype does.

Answering a blocked prompt sends `agent.send_keys` with the chosen key (`y`, `a`, `n`, `esc`).
Free-text sends use `agent.prompt`.

### 4. Inbox and desktop sidebar
Same data, two presentations, one breakpoint at **880px**.

- Phone: `/` is the inbox — "Needs you" section first, then per-space agent groups. Floating
  3-tab bar (Agents / Spaces / Settings), hidden on chat and diff.
- Desktop: a 328px left sidebar becomes the inbox and stays visible; the tab bar and the phone status
  bar mock disappear; `/` resolves to the last-selected pane's chat. Right-pane content is centered
  with a fluid gutter of `max(28px, calc(50% - 430px))`.
- Tapping any agent row anywhere opens its chat *and* syncs selected space/tab. Non-agent panes open
  chat directly in raw mode.

Drop the phone status-bar mock — it exists only to make the prototype read as a phone.

### 5. Spaces and space detail
`/spaces` (phone) and `/spaces/:id`. Space cards with monogram, worktree badge, cwd, rollup status
and counts; space detail with a horizontal tab strip (red dot when a pane inside is blocked), pane
cards showing recent output tails, and the dashed add-tab / split-pane affordances.

**Every mutation goes through the bottom sheet first — nothing mutates on a single tap.** The sheet
shows the consequence in plain language and, for destructive actions, a red CTA. Confirming fires a
toast. See the README's action/call table.

### 6. Diff viewer — `/pane/:id/diff`
Unified diff, syntax highlighted, with a per-file chip strip, a soft-wrap toggle, and add/remove
encoded as both row tint and a 2px left mark. Do not hand-roll the tokenizer the prototype uses —
use this repo's existing highlighter, or Shiki, and map its theme onto the README's syntax colors.

### 7. Settings
Theme picker (`herdr-dark`, `ash`, `gruvbox`), three behavior toggles (push when blocked, follow
focused pane, keep ANSI colors in raw), and a read-only server info card. Writes land in
`config.toml` and then call `server.reload_config`.

## Deliberate design decisions — keep these

- Blocked agents outrank everything. They are the first thing on every screen that can show them, and
  they are the only place red is used besides destructive actions.
- Terminal output is never hidden, but never the default view either.
- Ids (`w1:p2`) are always visible in mono. This is a power-user tool; the operator uses them.
- Confirmation sheets state consequences literally ("Running processes are killed. A worktree
  checkout on disk is left alone.") rather than asking "Are you sure?".
- Colour carries status, but so do shape and text — never colour alone.

## Things the design does not cover — ask me before inventing

Disconnect and reconnect UI, agent error states, empty states (no spaces, no agents), push
notification permission flow, and auth. Flag them when you reach them; do not design them yourself.

The socket-call captions in the prototype (`agent.send_keys → w1:p2` and similar) are a debug
affordance. Ship them behind a developer setting, off by default — not as permanent chrome.

## Definition of done

Every screen and state in the standalone prototype is reachable in the real app, on a phone-width
viewport and at desktop width, driven by live socket data, with real routes rather than a view enum.
