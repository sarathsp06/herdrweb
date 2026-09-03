import { writable, get, type Readable } from 'svelte/store';
import type { Call, CallResult, ConnState, SessionEvent, Space } from '$lib/protocol';
import type { Transport } from '$lib/transport';
import { SessionModel } from './model';

export interface Session {
  spaces: Readable<Space[]>;
  focus: Readable<SessionModel['focus']>;
  connection: Readable<ConnState>;
  request(call: Call): Promise<CallResult>;
  start(): void;
  stop(): void;
}

export function createSession(transport: Transport): Session {
  const model = new SessionModel();
  const spaces = writable<Space[]>([]);
  const focus = writable<SessionModel['focus']>({});
  const connection = writable<ConnState>('connecting');

  const publish = () => {
    spaces.set(model.spaces);
    focus.set(model.focus);
  };

  let unEv: (() => void) | null = null;
  let unConn: (() => void) | null = null;

  return {
    spaces,
    focus,
    connection,
    request: (call) => transport.request(call),
    start() {
      unEv = transport.subscribe((ev: SessionEvent) => {
        model.apply(ev);
        publish();
      });
      unConn = transport.onConnection((s) => connection.set(s));
      transport.start();
    },
    stop() {
      unEv?.();
      unConn?.();
      transport.stop();
    }
  };
}

// Convenience for reading a store's current value outside reactive scopes.
export { get };
