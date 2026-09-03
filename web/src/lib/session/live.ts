import { browser } from '$app/environment';
import { FixtureTransport } from '$lib/transport/fixture';
import { SocketTransport } from '$lib/transport/socket';
import type { Transport } from '$lib/transport';
import { createSession, type Session } from './store';

/** Pick a transport. Default is the live Go bridge; `?fixtures=1` forces the
 *  mocked dataset (used for e2e and screenshots). Server-side render gets
 *  fixtures so the first paint is deterministic. */
export function pickTransport(): Transport {
  if (!browser) return new FixtureTransport();
  const params = new URLSearchParams(location.search);
  if (params.has('fixtures')) return new FixtureTransport();
  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  return new SocketTransport(`${proto}://${location.host}/ws`);
}

let singleton: Session | null = null;

export function session(): Session {
  if (!singleton) {
    singleton = createSession(pickTransport());
    singleton.start();
  }
  return singleton;
}
