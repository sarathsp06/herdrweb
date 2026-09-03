import type { Call, CallResult, ConnState, SessionEvent } from '$lib/protocol';
import type { Transport } from './index';

interface Pending {
  resolve: (r: CallResult) => void;
  reject: (e: unknown) => void;
}

/** Talks to the Go bridge over a WebSocket. The bridge sends `snapshot` and
 *  patch events; calls are id-correlated request/response frames. */
export class SocketTransport implements Transport {
  private ws: WebSocket | null = null;
  private handlers = new Set<(ev: SessionEvent) => void>();
  private conn = new Set<(s: ConnState) => void>();
  private connState: ConnState = 'closed';
  private pending = new Map<string, Pending>();
  private seq = 0;
  private backoff = 500;
  private stopped = false;
  private retry: ReturnType<typeof setTimeout> | null = null;

  constructor(private url: string) {}

  subscribe(handler: (ev: SessionEvent) => void): () => void {
    this.handlers.add(handler);
    return () => this.handlers.delete(handler);
  }

  onConnection(handler: (s: ConnState) => void): () => void {
    this.conn.add(handler);
    handler(this.connState);
    return () => this.conn.delete(handler);
  }

  private setConn(s: ConnState): void {
    this.connState = s;
    for (const h of this.conn) h(s);
  }

  start(): void {
    this.stopped = false;
    this.open();
  }

  private open(): void {
    this.setConn(this.connState === 'closed' ? 'connecting' : 'reconnecting');
    const ws = new WebSocket(this.url);
    this.ws = ws;
    ws.onopen = () => {
      this.backoff = 500;
      this.setConn('open');
    };
    ws.onmessage = (e) => this.onMessage(e.data as string);
    ws.onclose = () => {
      this.ws = null;
      for (const [, p] of this.pending) p.reject(new Error('socket closed'));
      this.pending.clear();
      if (this.stopped) {
        this.setConn('closed');
        return;
      }
      this.setConn('reconnecting');
      this.retry = setTimeout(() => this.open(), this.backoff);
      this.backoff = Math.min(this.backoff * 2, 10000);
    };
    ws.onerror = () => ws.close();
  }

  private onMessage(raw: string): void {
    let msg: any;
    try {
      msg = JSON.parse(raw);
    } catch {
      return;
    }
    if (msg.id && this.pending.has(msg.id)) {
      const p = this.pending.get(msg.id)!;
      this.pending.delete(msg.id);
      if (msg.error) p.reject(new Error(String(msg.error)));
      else p.resolve((msg.result ?? { ok: true }) as CallResult);
      return;
    }
    // Otherwise it is a session event (snapshot / patch).
    if (msg.type) for (const h of this.handlers) h(msg as SessionEvent);
  }

  request(call: Call): Promise<CallResult> {
    return new Promise((resolve, reject) => {
      const ws = this.ws;
      if (!ws || ws.readyState !== WebSocket.OPEN) {
        reject(new Error('socket not open'));
        return;
      }
      const id = `c${++this.seq}`;
      this.pending.set(id, { resolve, reject });
      ws.send(JSON.stringify({ id, method: call.method, params: call.params }));
    });
  }

  stop(): void {
    this.stopped = true;
    if (this.retry) clearTimeout(this.retry);
    this.ws?.close();
  }
}
