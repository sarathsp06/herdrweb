import type { Call, CallResult, ConnState, SessionEvent } from '$lib/protocol';

/** The whole UI reads from this. Two impls: FixtureTransport, SocketTransport. */
export interface Transport {
  /** Subscribe to the live event stream (snapshot first, then patches). */
  subscribe(handler: (ev: SessionEvent) => void): () => void;
  /** Readable connection state; calls handler immediately with current value. */
  onConnection(handler: (state: ConnState) => void): () => void;
  /** Fire a socket call and await its result. */
  request(call: Call): Promise<CallResult>;
  /** Begin connecting / bootstrapping. */
  start(): void;
  /** Tear down. */
  stop(): void;
}
