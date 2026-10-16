# Packages

Go backend. Entry point plus four `internal/` packages.

* [cmd/herdr-bridge](cmd-herdr-bridge.md) - main: flags, lifecycle, HTTP server
* [internal/herdr](herdr.md) - Herdr Unix-socket client (persistent mux + event subscription)
* [internal/server](server.md) - Hub: one Herdr connection, WS fan-out, REST endpoints
* [internal/protocol](protocol.md) - canonical UI model + Herdr snapshot normalizer
* [internal/config](config.md) - UI settings file read/write, `[web]` migration out of Herdr's config.toml
* [internal/webui](webui.md) - go:embed of the built SvelteKit assets with SPA fallback
