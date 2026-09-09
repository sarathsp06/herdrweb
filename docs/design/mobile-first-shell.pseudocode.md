# Feature: Mobile-first shell rewrite

## Context

Herdr Web is operated daily from a phone, yet its shell is a desktop sidebar retrofitted to
mobile: navigation hides behind a corner FAB and a drawer, a breadcrumb bar burns vertical
space, the pane screen stacks ~140px of chrome above the terminal, and the most-tapped
controls sit below the 44px touch minimum. Rewrite the navigation shell and screen chrome
mobile-first while keeping the protocol mirror, the thin WebSocket pass-through, the raw
terminal renderer, composer routing, bottom-sheet confirmations, themes, and PWA/push intact.

## Design

- the phone gets permanent thumb-reach navigation
  - `✎` new `BottomNav.svelte` — fixed bottom tab bar: Inbox · Spaces · Settings; targets ≥48px, safe-area aware; Inbox tab badges the count of blocked agents
  - the corner FAB, mobile drawer, and the sidebar's foot buttons are deleted — with them the `nav_corner` setting end-to-end (web config, Go `[web] nav_corner`, settings screen, README)
  - hidden on `/pane/*` — the pane is a fullscreen push, back lives in its header, terminal keeps the height
- breadcrumbs are replaced by per-screen headers
  - `✎` new `PaneHeader.svelte` — back chevron (44px) + pane label + space sub-label + status pill + overflow (⋯) for pane actions (rename, close, diff)
  - list screens (Inbox, Spaces, Settings) keep titling their own headers; `Breadcrumbs.svelte` deleted
  - desktop reuses the same pane header as a slim top bar next to the sidebar
- the inbox becomes a real mobile screen, not an embedded sidebar
  - agents sectioned by urgency: **Needs you** (blocked) → **Working** → **Idle/Done**; status stays glyph + colour + word
  - whole row ≥56px opens the agent's pane; space · branch as the sub-line
  - spaces collapse to a horizontal chip strip on top (rollup glyph + label), chip tap opens that space's chat pane, long-label ellipsis; a trailing "manage" chip goes to /spaces
  - `Sidebar.svelte` becomes desktop-only and reuses the same inbox list component — one rendering of an agent row, two shells
- the pane screen sheds chrome so the terminal owns the height
  - the tab chip strip leaves the screen: tapping the header title opens a tab/pane switcher bottom sheet (reuses `BottomSheet`), blocked-dot per tab preserved
  - direct-control moves into the key row as a key-sized toggle (◎/◉) — same reset-on-pane-switch semantics
  - key row, send, and attach grow to 44px; key row stays horizontally scrollable
- management actions get out of the primary path
  - Spaces list: card tap opens the space chat; Tabs / Rename / Close move behind one ⋯ overflow per card (action list in a bottom sheet, confirmations unchanged)
  - space detail keeps explicit management (it is the management screen) with 44px targets
- verification against the real surface
  - Playwright e2e updated: tab bar visible under 880px, drawer/FAB gone, pane push hides tab bar; fixtures mode still exercises every flow
  - screenshots re-taken; README screens section updated

## Open Questions

None — navigation model (bottom tab bar), pane tab switching (header-tap sheet), and full
`nav_corner` removal were all approved 2026-09-09.
