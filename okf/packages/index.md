# Packages

Go backend. Entry point plus four `internal/` packages.

* [cmd/herdr-bridge](cmd-herdr-bridge.md) - main: flags, lifecycle, HTTP server
* [internal/herdr](herdr.md) - Herdr socket client (one connection per call + event subscription)
* [internal/server](server.md) - Hub: Herdr event subscription, WS fan-out, REST endpoints
* [internal/protocol](protocol.md) - canonical UI model + Herdr snapshot normalizer
* [internal/config](config.md) - `[web]` table read/write in Herdr's config.toml
* [internal/webui](webui.md) - go:embed of the built SvelteKit assets with SPA fallback
