# AGENTS.md

Herdr Web: one Go binary (`herdr-bridge`) that embeds a SvelteKit SPA and bridges the browser to the Herdr daemon's Unix socket. See `README.md` for the screen tour and `okf/` for a structured knowledge bundle.

## Build & run

- `make build` — builds the SPA, copies it to `internal/webui/dist/`, then compiles the binary. **`go build` alone does NOT refresh the UI** (it's `go:embed`ed from `dist/`); always `make build` after web changes. Node is required at build time.
- `make run` — build + run on `http://127.0.0.1:7331`.
- `make dev` — Vite dev server with HMR; it proxies `/ws` + `/api` to a bridge on `:7331`, so run `make run` in another shell for live data.
- Keep `internal/webui/dist/.gitkeep` (dir must stay tracked; built assets are gitignored). `make build` deletes it — restore before committing.

## Test / lint

- `make check` = `make lint` (`go vet` + `svelte-check`) + `make test` (Go + Vitest). Don't finish red.
- Single Go test: `go test ./internal/herdr/ -run TestMux…`. Use `-race` for `internal/herdr` and `internal/server` (both have concurrency-sensitive code + race tests).
- Web unit: `cd web && npm run test:unit` (Vitest). Typecheck: `cd web && npm run check`.
- E2e: `cd web && npm run test:e2e` — Playwright self-hosts `build`+`preview` on `:4173` in **fixtures mode** (`?fixtures=1`), no live bridge needed; run `npx playwright install chromium` first.
- Screenshots: `make screenshots` — needs a bridge running on `:7331`.
- Go: `gofmt`; this repo requires Go 1.22+ `for range n` loops (avoid `for i := 0; i < n; i++`).

## Load-bearing invariants (easy to break)

- **Protocol is mirrored, not generated.** `internal/protocol/*.go` is canonical; `web/src/lib/protocol/index.ts` is a hand-written twin — change both together, and keep `web/src/lib/transport/fixture.ts` in sync.
- **Snapshot wire shape is flat**: `{"type":"snapshot","spaces":[…],"focus":{…}}` — no nested `snapshot` wrapper. Fixtures and `SessionModel.apply` must match, or the UI renders "0 spaces".
- **The browser↔bridge WebSocket is a thin pass-through**: the frontend calls Herdr socket methods directly, so params must match Herdr exactly:
  - `pane.read` → `{ pane_id, source: 'recent_unwrapped', lines }`; the reply is `result.read.text` (a string), not `lines`.
  - `agent.send_keys` → `keys` is a **string array** (e.g. `['esc']`), tokens `up/down/left/right/tab/shift+tab/enter/esc/ctrl+c/ctrl+d` (Herdr accepts any key-combo string: printable keys, `enter`/`esc`, `ctrl+`/`alt+`/`shift+` chords, `f1`…).
  - `agent.prompt` → `wait.until` is an array; don't block the UI on its (long) resolution.
- **Live agent panes are raw terminal only** — there is no chat/transcript view or chat/raw toggle. Don't reintroduce a transcript parser.

## Conventions & gotchas

- No auth; binds loopback by design. Expose over a tailnet with `./bin/herdr-bridge -addr $(tailscale ip -4):7331` (not `0.0.0.0`).
- Fonts are self-hosted via `@fontsource` and embedded (no CDN). Colours are CSS variables in `web/src/lib/tokens.css`; themes override via `[data-theme]` (`herdr-dark`/`ash`/`gruvbox`/`solarized-light`).
- `/.worktrees/` holds Rx-pipeline worktrees (gitignored); don't commit them.
- Go tests use in-process fake Unix-socket servers (`internal/herdr/*_test.go`) — no real daemon needed.
- `okf/` is a generated OKF bundle; validate with `python3 ~/.claude/skills/okf-repo/scripts/validate_okf.py okf --strict` after edits.
