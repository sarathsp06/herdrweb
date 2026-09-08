# Log

## 2026-09-03

**Initial bundle.**

Created the OKF bundle for herderweb from the codebase and `docs/ARCHITECTURE_RECOMMENDATIONS.md`. Documents the Go bridge (`cmd/herdr-bridge`, `internal/{herdr,server,protocol,config,webui}`), the SvelteKit frontend (`web/src`), the HTTP/WS service surface, the session-snapshot model, and configuration.

Reflects the post-`bridge-ipc-hardening` backend: `herdr.Client` uses a persistent, id-multiplexed socket connection and `server.Hub` bounds in-flight WebSocket RPC goroutines per connection.

## 2026-09-08

**Direct-control swipe + theme/chrome-sync fixes.**

Documented the pane view's `◎ direct control` toggle (agent panes: swipe the transcript to send one arrow key per swipe via `agent.send_keys`, replacing native scroll; gated off if the pane stops being an agent) and the `theme-color`/`apple-mobile-web-app-status-bar-style` meta sync that now tracks the active theme pre-paint and — for `theme-color` — on every live switch, fixing a stuck-dark-chrome bug on light themes (the status-bar-style meta is kept correct too, but iOS only applies it on an installed PWA's next cold launch). Corrected a stale theme list (`herdr-dark`/`ash`/`gruvbox`/`solarized-light`) across `themes.md`, `settings.md`, and `navigation.md` to the actual four themes (`herdr-dark`, `gruvbox`, `solarized-light`, `paper`) — `ash` never shipped; `paper` was missing. Fixed the same list in `README.md` and `AGENTS.md`. Corrected `terminal-view.md`'s stale "soft-wrap under 880px" claim to the actual `use:fitToWidth` font-shrink behaviour.
