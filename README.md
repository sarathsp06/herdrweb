# Handoff: Herdr Web Interface (personal, mobile-first)

## Overview
A personal web client for **Herdr**, a terminal/agent multiplexer. The interface lets one operator
watch every coding agent running across their spaces, answer the ones that are blocked, read diffs,
manage spaces/tabs/panes, and change server settings — from a phone, with a desktop layout for wide
viewports.

The design's organizing idea: **agents, not terminals, are the primary objects.** The default screen
is an agent inbox sorted by who needs the human. Raw terminal scrollback is available but secondary.

## About the Design Files
The files in this bundle are **design references created in HTML** — prototypes showing intended look
and behavior, not production code to copy. The task is to **recreate these designs in the target
codebase's existing environment** (the real client is SvelteKit + a Go bridge to the Herdr socket)
using its established patterns, stores, and component conventions. If no environment exists yet,
pick the most appropriate framework and implement there.

All data in the prototype is mocked (see `Mock data` below). Every socket call is *narrated* in the
UI as a small monospace caption (e.g. `agent.send_keys → w1:p2`) so the intended API surface is
visible; in production those captions are a debug affordance behind a setting, not core chrome.

## Fidelity
**High-fidelity.** Colors, typography, spacing, radii, and interaction states are final and should be
recreated closely. Layout is fluid/responsive, not fixed-canvas — treat the numbers as the intended
values at the two breakpoints described under *Responsive behavior*.

## Design Tokens

### Colors
| Role | Value |
|---|---|
| App background | `#0a0a0a` |
| Page background (outer, phone shell) | `#08080a` |
| Sidebar background | `#0c0c0e` |
| Raised card | `#141416` |
| Card, code/tool block | `#121214` |
| Code editor surface | `#0d0d0f` |
| Card hover | `#1b1b1f` / `#1e1e22` |
| Toast background | `#1e1e22` |
| Text primary | `#fafafa` |
| Text on light buttons | `#18181b` |
| Text secondary | `#e4e4e7` / `#d4d4d8` |
| Text tertiary | `#a3a3a3` / `#8a8a92` |
| Text quaternary (meta) | `#737378` |
| Text disabled / gutters | `#5c5c63`, `#4a4a52`, `#52525b` |
| Hairline border | `rgba(255,255,255,.07)` |
| Control border | `rgba(255,255,255,.12)` — `.14` on inputs, `.16`/`.28` when selected |
| Surface tint (selected/hover) | `rgba(255,255,255,.05)` → `.08` |
| Status: working | `oklch(0.769 0.188 70.08)` (amber) |
| Status: blocked / destructive | `oklch(0.645 0.246 16.439)` (red) |
| Blocked text tint | `oklch(0.72 0.19 16.4)`, badge text `oklch(0.75 0.16 16.4)` |
| Status: done / connected | `oklch(0.696 0.17 162.48)` (green) |
| Status: idle | `#737378`; unknown `#52525b` |
| Worktree badge | text `oklch(0.75 0.14 250)` on `oklch(0.6 0.15 250 / 16%)` |

Blocked surfaces use `linear-gradient(180deg, oklch(0.645 0.246 16.439 / 9-10%), rgba(255,255,255,.02))`
with a `oklch(0.645 0.246 16.439 / 28-32%)` border.

### Syntax highlighting
| Token | Value |
|---|---|
| keyword | `oklch(0.78 0.14 300)` |
| string | `oklch(0.82 0.13 92)` |
| number | `oklch(0.78 0.13 250)` |
| function call | `oklch(0.8 0.11 195)` |
| component / TitleCase | `oklch(0.8 0.12 165)` |
| comment | `#5c5c63` |
| punctuation | `#8a8a92` |
| default | `#d4d4d8` |

Diff rows: added row background `oklch(0.696 0.17 162.48 / 9%)` with a 2px green left mark; removed
row `oklch(0.645 0.246 16.439 / 9%)` with a 2px red left mark; context rows transparent.

### Typography
- UI: **Geist** (400/500/600/700).
- Identifiers, IDs, code, terminal, all numerics: **Geist Mono** (400/500/600).
- Screen titles: 21px/1.1, 600, `letter-spacing:-.02em`.
- Chat bar title 14px/600; section labels 11–11.5px/600 uppercase `letter-spacing:.04–.05em`.
- Body copy in messages 13.5px/1.6; card titles 13.5px/600; card subtitles 11–12px.
- Meta/captions 10–11.5px mono. Code and diffs 11.5px/1.7 mono. Terminal tails 10.5px/1.6 mono.
- Composer textarea 14px/1.45.
- `text-wrap: pretty` on all prose blocks.

### Spacing, radius, motion
- Card padding 11–14px; screen gutters 14px (phone).
- Radii: 4–5px micro badges, 6–8px chips/small buttons, 9–11px buttons and cards, 12px prominent
  cards, 16px chat bubble (`16px 16px 4px 16px` for the user bubble), 18px composer,
  `22px 22px 34px 34px` bottom sheet.
- Touch targets: never below 44px on phone (list rows 44–54px, primary buttons 46–50px).
- Toast shadow `0 12px 30px -8px rgba(0,0,0,.8)`.
- Animations: `hpulse` 2.6s ease-in-out infinite (connection dot, opacity 1 → .3); `hsheet` .22s
  `cubic-bezier(.32,.72,0,1)` (sheet slides up from `translateY(100%)`); `hbounce` 1.2s ease-in-out
  infinite with .15s stagger across three 5px dots (agent working indicator); toggle knob
  `transition:left .16s`.
- Scrollbars hidden (`::-webkit-scrollbar{width:0;height:0}`); scroll containers still scroll.

## Screens / Views

### 1. Agent inbox (`view: "inbox"`) — phone only
**Purpose:** triage. Answer whoever is blocked, then scan everything else by space.

Layout: fixed header (title `Agents` + a green pulsing dot and the line
`default session · 4 spaces · connected`), then a scroll region with `padding:14px 14px 96px` (bottom
padding clears the floating tab bar).

- **"Needs you" section.** Red 7×7px rounded square + uppercase label, count right-aligned in mono.
  Each blocked agent is a full-width card: a `blocked` pill, agent name, pane id right-aligned, the
  approval question at 13px/1.45, then a footer line `<space> · waiting 4m`. Tapping opens chat.
- **Per-space groups.** Header row: space label (mono 600), branch in a bordered 5px-radius chip, then
  a 1px hairline filling the remaining width. Rows below show only *agent* panes: status dot,
  name + pane id, one-line truncated title, and a right column with status word (colored by status)
  over relative time.

### 2. Chat (`view: "chat"`) — the core screen
**Purpose:** read what an agent did and reply or approve.

Top bar: back chevron (32×32, phone only), agent name + status pill, second line `<space> · <pane id>`,
and a 2-segment `chat / raw` switch (active segment `#fafafa` on `#18181b` track).

Transcript (scrolls, auto-pinned to bottom on pane/mode/view change). Message kinds:
- **User prompt** — right-aligned white bubble, dark text, max-width 84%, caption `agent.prompt · 6m`.
- **Reasoning** — collapsed button "Thought for 8s" with ▸/▾; expanded body is indented behind a 2px
  left rule in `#8a8a92`.
- **Agent text** — plain 13.5px/1.6 `#e4e4e7`, no bubble, with a row of 30×30 ghost icon buttons
  (copy, retry) underneath.
- **Tool call** — bordered card: chevron, tool-name chip, truncated argument in mono, right-side
  result label (green `ok`, or neutral `44 lines`); expands to show output.
- **Code block** — header with file path, language, and a copy button that flips to `copied`; body is
  horizontally scrollable with a 32px right-aligned line-number gutter.
- **Diff summary** — tappable card: ◫ glyph, file path, "3 files changed", `+34 −12`, chevron; footer
  strip previews one changed line. Opens the diff viewer.
- **Blocked prompt** — the most prominent element: red-tinted gradient card with a `blocked` pill and
  "approval requested 4m", the question at 13.5px, the literal terminal text in a
  `rgba(0,0,0,.4)` inset block, then answer buttons **Yes** (solid white), **Yes, don't ask**
  (subtle), **No** (red outline), **esc** (ghost). Caption: `agent.send_keys → <pane id>`.
- **Working indicator** — three bouncing dots + `working · editing PaneCell.svelte`.

**Raw mode** replaces the transcript with monospace scrollback, preceded by
`pane.read · source=recent-unwrapped · lines=80`. Lines are individually colored (diff lines green/red,
status amber, commentary grey) and never wrap.

Composer (pinned): a horizontally scrolling row of quick-send chips — `y n esc ctrl+c /status` when
blocked, otherwise `continue / run tests / git diff / esc / ctrl+c` — then an 18px-radius input group
with an auto-growing textarea (`max-height:96px`), three 32×32 utility buttons, a mono hint
(`agent.prompt`, or `blocked — answer above`), and a 34×34 send button that goes from
`rgba(255,255,255,.12)` to solid `#fafafa` once the draft is non-empty.

### 3. Diff viewer (`view: "diff"`)
Top bar: back, file path (mono, truncated), `+21 −7 svelte`, and a `wrap on/off` toggle. Below it a
horizontally scrolling file strip; each chip shows basename + per-file `+add −del`, selected chip gets
a brighter border and tinted background.

Body: unified diff at 11.5px/1.7 mono. Each row = 34px right-aligned line number, 13px sign column,
then syntax-highlighted tokens; row background and 2px left mark encode add/remove. Wrap off means
rows scroll horizontally; wrap on lets tokens flow.

Footer, over a fade to `#0a0a0a`: **Back to chat** (outlined, flex:1) and **Copy path** (solid white,
flips to `Copied`).

### 4. Spaces (`view: "spaces"`) — phone only
Header `Spaces` + `4 spaces · 6 tabs` + a solid white **＋ New**. Each space is a card: a 30×30
2-letter monogram, label (mono 600), optional `worktree` badge, cwd truncated, and a right column with
a rollup status word (blocked / working / idle / —) over `3 tabs · 4 panes`. A 3-button footer row
divided by hairlines: **Tabs**, **Rename**, **Close** (red).

### 5. Space detail (`view: "space"`)
Top bar: back, space label, cwd, **Rename**. Then a horizontal tab strip; each tab shows label, a red
6px dot if any pane inside is blocked, and its pane count; a dashed **＋** button appends a tab.

Body: a row with the active tab's label (uppercase) and id plus **Rename** / **Close tab** (red
outline), then one card per pane: status dot, pane label (mono), subtitle (`claude · editing
PaneGrid.svelte`), pane id, a monospace tail of recent output (horizontally scrollable, whitespace
preserved), and a footer row **Chat** (or **Read** for non-agent panes) / **Rename** / **Close**.
Ends with a dashed **＋ Split a new pane** button.

### 6. Settings (`view: "settings"`)
Three labelled sections:
- **Theme** — three cards (`herdr-dark`, `ash`, `gruvbox`), each three 14px swatches over the name;
  selection is border + background tint. Caption: `writes [theme] in config.toml → server.reload_config`.
- **Behavior** — three toggle rows (name, description, 38×22 track with an 18px knob; on = green track,
  dark knob at `left:18px`): *Push when blocked*, *Follow focused pane*, *Keep ANSI colors in raw*.
- **Server** — a read-only key/value card (bridge `go · :7331`, socket path, version `0.8.2`, protocol
  `ok`, uptime), the caption `svelte → go bridge → herdr socket`, and a **Reload config** button.

### 7. Bottom tab bar — phone only
Three tabs — `◉ Agents`, `⌗ Spaces`, `⚙ Settings` — over a gradient fade, 50px min height, active
tab gets `rgba(255,255,255,.08)` and `#fafafa`. Hidden on chat and diff (those are full-screen
pushes) and on desktop.

### 8. Bottom sheet
Every mutating action routes through one sheet component: 36×4px grab handle, title, explanatory
body, optional text input(s) (label; plus a cwd field when creating a space), the socket call it will
make in mono, then **Cancel** / a primary CTA. Destructive variants use a red CTA and spell out the
consequence ("Running processes are killed. A worktree checkout on disk is left alone."). Confirming
closes the sheet and fires a toast.

Sheet variants and their calls:
| Action | Call |
|---|---|
| New space | `workspace.create { cwd, label, focus: false }` |
| New tab | `tab.create { workspace_id, label }` |
| Split pane | `pane.split { direction: "right", ratio: 0.5 }` |
| Rename space / tab / pane | `workspace.rename` / `tab.rename` / `pane.rename` |
| Close space / tab / pane | `workspace.close` / `tab.close` / `pane.close` |

### 9. Toast
Bottom-anchored (`bottom:92px`, 14px insets), `#1e1e22`, green dot + monospace text, auto-dismiss
after 2.2s, single-slot (a new toast replaces the current one).

## Interactions & Behavior
- **Navigation is a view switch, not a router**, in the prototype. In production these should be real
  routes: `/` inbox, `/pane/:id` chat, `/pane/:id/diff`, `/spaces`, `/spaces/:id`, `/settings`.
- Tapping any agent row anywhere opens chat for that pane **and** syncs the selected space/tab.
- Non-agent panes open chat in **raw** mode directly (they have no transcript).
- Chat auto-scrolls to bottom on pane change, mode change, and view change (the prototype pins on
  rAF + 80ms + 220ms; in production, scroll after paint / on transcript mutation).
- Reasoning and tool blocks are independently collapsible; state persists while the view is mounted.
- Copy buttons show a transient `copied` label (1.6s) and also fire a toast.
- Every destructive or creating action opens the sheet first — nothing mutates on a single tap.
- Hover states exist for pointer users (background lifts to `rgba(255,255,255,.05–.06)`, borders
  brighten, primary buttons drop to `opacity:.86`) but no behavior depends on hover.

### Responsive behavior
Single breakpoint at **880px** (prototype also exposes a `layout` override: `auto | phone | desktop`).

- **< 880px (phone).** Status bar mock, single column, floating bottom tab bar, full-screen pushes for
  chat and diff, sheets from the bottom.
- **≥ 880px (desktop).** A 328px left sidebar appears (`#0c0c0e`, right hairline): brand mark, session
  chip, a 3-up nav row, then "Needs you" followed by every space with its agent rows inline —
  the sidebar *is* the inbox. The bottom tab bar and the phone status bar are hidden. `inbox` view
  resolves to `chat` and `spaces` resolves to `space`. Content in the right pane is centered by a
  fluid gutter: `max(28px, calc(50% - 430px))`, so text lines cap around 860px.

## State Management
Prototype state (all local):
```
view      "inbox" | "chat" | "diff" | "spaces" | "space" | "settings"
pane      selected pane id, e.g. "w1:p1"
mode      "chat" | "raw"
spaceId   selected space
tabId     selected tab
diff      selected diff file path
sheet     null | { kind, ctx, title, body, cta, hasInput, hasCwd, call, … }
sv, scwd  sheet text inputs (label, cwd)
draft     composer text
open      { r0, t0 }  collapse state per reasoning/tool block
wrap      diff soft-wrap
theme     "herdr-dark" | "ash" | "gruvbox"
tg        { notify, follow, ansi }  behavior toggles
toast     current toast string ("" = hidden)
w         window width (resize-listened, 8px threshold)
```

In production, replace the mock arrays with a live session store fed by the socket. Events the UI
already assumes exist: `workspace.updated`, `tab.updated`, `pane.updated`,
`pane.agent_status_changed`, `pane.output_matched`. Reads use
`pane.read { source: "recent-unwrapped", lines: 80 }`. Prompts use
`agent.prompt { target, text, wait: { until: ["idle","blocked"], timeout_ms: 900000 } }`; answering a
blocked prompt uses `agent.send_keys`. Collapse/scroll/draft state stays view-local; theme and
toggles belong in `config.toml` via `server.reload_config`.

Agent status is a 5-value enum — `working | blocked | done | idle | unknown` — and drives every dot,
status word, rollup, and sort. Space rollup = blocked if any agent pane is blocked, else working if
any is working, else idle if it has agent panes, else `—`.

### Mock data
Four spaces: `hedr-web` (feat/chat-ui; agents claude=working, codex=blocked; plus dev server and log
tail panes), `api` (main; codex=done, psql), `api/billing` (a worktree at
`~/.herdr/worktrees/api/billing`; claude=blocked), `dotfiles` (a plain zsh pane). Three diff files with
hand-tokenized lines. Replace all of it with live data; keep the shapes.

## Assets
None. No images or icon files — status is color + shape, and the few glyphs (`◉ ⌗ ⚙ ◫ ‹ ↑ ＋ ▸ ▾`) are
Unicode characters rendered in Geist / Geist Mono. Fonts come from Google Fonts (Geist, Geist Mono);
self-host them in production.

## Files
- `Herdr Mobile.dc.html` — the design source. All six screens, the sheet, and the toast in one file;
  markup at the top, mock data and behavior in the script at the bottom.
- `herdr-web-standalone.html` — the same design bundled into one self-contained offline file. Open it
  in a browser to click through the prototype; resize past 880px for the desktop layout.
