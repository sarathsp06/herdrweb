# Log

## 2026-09-03

**Initial bundle.**

Created the OKF bundle for herderweb from the codebase and `docs/ARCHITECTURE_RECOMMENDATIONS.md`. Documents the Go bridge (`cmd/herdr-bridge`, `internal/{herdr,server,protocol,config,webui}`), the SvelteKit frontend (`web/src`), the HTTP/WS service surface, the session-snapshot model, and configuration.

Reflects the post-`bridge-ipc-hardening` backend: `herdr.Client` uses a persistent, id-multiplexed socket connection and `server.Hub` bounds in-flight WebSocket RPC goroutines per connection.
