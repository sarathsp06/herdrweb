# Frontend

SvelteKit + TypeScript SPA in `web/src`, built with the static adapter and embedded into the binary. Mirrors the Go [protocol](/packages/protocol.md) types and talks to the bridge over `/ws` + `/api`.

* [Transport](transport.md) - SocketTransport (live) and FixtureTransport (`?fixtures=1`)
* [Session store](session-store.md) - SessionModel → Svelte stores wiring
* [Routes](routes.md) - the SvelteKit screens
* [Composer](composer.md) - prompt input + terminal nav keys
* [Navigation & layout](navigation.md) - breadcrumb bar, sidebar/drawer, themes/text-size
