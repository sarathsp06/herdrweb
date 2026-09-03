import type { Call, CallResult, ConnState, SessionEvent, Snapshot } from '$lib/protocol';
import type { Transport } from './index';

// README mock: four spaces. Agent panes carry transcripts; the diff fixture
// backs the diff viewer.

function pane(id: string, label: string, sub: string, status: string, agent: boolean, tail: string[] = []) {
  return { id, label, sub, status: status as any, agent, tail };
}

export const FIXTURE_SNAPSHOT: Snapshot = {
  focus: { spaceId: 'w1', tabId: 'w1:t1', paneId: 'w1:p2' },
  spaces: [
    {
      id: 'w1',
      label: 'hedr-web',
      cwd: '~/code/hedr-web',
      branch: 'feat/chat-ui',
      worktree: null,
      tabs: [
        {
          id: 'w1:t1',
          label: 'agents',
          panes: [
            pane('w1:p1', 'claude', 'claude · editing PaneCell.svelte', 'working', true, [
              'running vitest…',
              'PASS  src/lib/session/model.test.ts (12)',
              'watching for changes'
            ]),
            pane('w1:p2', 'codex', 'codex · approval requested', 'blocked', true, [
              '? Apply patch to src/routes/+layout.svelte? (y/n)',
              '  1 file changed, 24 insertions(+)'
            ])
          ]
        },
        {
          id: 'w1:t2',
          label: 'server',
          panes: [
            pane('w1:p3', 'dev', 'vite dev', 'idle', false, [
              'VITE v6.1.0  ready in 412 ms',
              'Local:   http://localhost:5173/'
            ]),
            pane('w1:p4', 'logs', 'tail -f app.log', 'idle', false, [
              '12:04:11 GET / 200 4ms',
              '12:04:13 GET /pane/w1:p2 200 6ms'
            ])
          ]
        }
      ]
    },
    {
      id: 'w2',
      label: 'api',
      cwd: '~/code/api',
      branch: 'main',
      worktree: null,
      tabs: [
        {
          id: 'w2:t1',
          label: 'agents',
          panes: [
            pane('w2:p1', 'codex', 'codex · migration written', 'done', true, [
              'wrote migrations/0042_add_index.sql',
              'done'
            ]),
            pane('w2:p2', 'psql', 'psql api_dev', 'idle', false, ['api_dev=#'])
          ]
        }
      ]
    },
    {
      id: 'w3',
      label: 'api/billing',
      cwd: '~/.herdr/worktrees/api/billing',
      branch: 'feat/billing',
      worktree: '~/.herdr/worktrees/api/billing',
      tabs: [
        {
          id: 'w3:t1',
          label: 'agents',
          panes: [
            pane('w3:p1', 'claude', 'claude · confirm destructive migration', 'blocked', true, [
              '? Drop column invoices.legacy_total? This is irreversible. (y/n)'
            ])
          ]
        }
      ]
    },
    {
      id: 'w4',
      label: 'dotfiles',
      cwd: '~/dotfiles',
      branch: 'main',
      worktree: null,
      tabs: [{ id: 'w4:t1', label: 'shell', panes: [pane('w4:p1', 'zsh', 'zsh', 'idle', false, ['~/dotfiles ❯'])] }]
    }
  ]
};

// Diff fixture for the diff viewer (hand-tokenized-ish; Shiki re-highlights live).
export const FIXTURE_DIFFS: Record<string, { files: { path: string; add: number; del: number; lang: string; body: string }[] }> = {
  'w1:p2': {
    files: [
      {
        path: 'src/routes/+layout.svelte',
        add: 24,
        del: 0,
        lang: 'svelte',
        body: "@@ -1,3 +1,5 @@\n <script lang=\"ts\">\n+  import { session } from '$lib/session/store';\n+  import Sidebar from '$lib/screens/Sidebar.svelte';\n   let { children } = $props();\n </script>"
      },
      {
        path: 'src/lib/session/store.ts',
        add: 10,
        del: 2,
        lang: 'ts',
        body: "@@ -8,2 +8,10 @@\n-export const spaces = [];\n+export const session = createSession(transport);\n+session.start();"
      }
    ]
  }
};

type Handler = (ev: SessionEvent) => void;

export class FixtureTransport implements Transport {
  private handlers = new Set<Handler>();
  private conn = new Set<(s: ConnState) => void>();
  private timers: ReturnType<typeof setTimeout>[] = [];

  subscribe(handler: Handler): () => void {
    this.handlers.add(handler);
    handler({ type: 'snapshot', ...FIXTURE_SNAPSHOT });
    return () => this.handlers.delete(handler);
  }

  onConnection(handler: (s: ConnState) => void): () => void {
    this.conn.add(handler);
    handler('open');
    return () => this.conn.delete(handler);
  }

  emit(ev: SessionEvent): void {
    for (const h of this.handlers) h(ev);
  }

  async request(call: Call): Promise<CallResult> {
    if (call.method === 'pane.read') {
      const found = FIXTURE_SNAPSHOT.spaces
        .flatMap((s) => s.tabs)
        .flatMap((t) => t.panes)
        .find((p) => p.id === call.params.pane_id);
      return { ok: true, read: { text: (found?.tail ?? []).join('\n') } };
    }
    return { ok: true };
  }

  start(): void {}
  stop(): void {
    this.timers.forEach(clearTimeout);
  }
}
