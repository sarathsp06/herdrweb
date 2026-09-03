---
type: Go Package
title: internal/protocol
description: Canonical Go UI model and the Herdr snapshot → UI normalizer; mirrored by web/src/lib/protocol
tags: [protocol, normalizer, model, canonical]
timestamp: 2026-09-03T00:00:00Z
---

# Responsibilities

`defines` the normalized model the bridge sends to browsers and maps Herdr's flat snapshot onto it. These Go types are **canonical**; `web/src/lib/protocol/index.ts` is the hand-mirrored TypeScript twin.

# Model

- `Snapshot{ Type "snapshot"; Spaces []Space; Focus }` — marshals **flat**: `{"type":"snapshot","spaces":[…],"focus":{…}}`.
- `Space{ id,label,cwd,branch,worktree,tabs }` → `Tab{ id,label,panes }` → `Pane{ id,label,sub,status,agent,tail }`.
- `Status` — 5-value enum: `working | blocked | done | idle | unknown`.
- `Normalize(*HerdrSnapshot) Snapshot`, `normalizePane`, `normStatus`, `base(path)`.

⚠ Flatness matters: the frontend once expected a nested `{snapshot:…}` wrapper and rendered "0 spaces"; the contract is flat and must stay mirrored in TS.

# Citations

* [internal/protocol/protocol.go](/internal/protocol/protocol.go)
* [session snapshot concept](/concepts/session-snapshot.md)
