<script lang="ts">
  import type { AgentRef } from '$lib/session/derive';
  import StatusGlyph from '$lib/ui/StatusGlyph.svelte';

  let { agent, sub, onopen }: { agent: AgentRef; sub: string; onopen: () => void } = $props();

  const wordColor: Record<string, string> = {
    blocked: 'text-(--blocked-badge-text)',
    working: 'text-working',
    done: 'text-done'
  };
</script>

<button
  class="arow flex min-h-[52px] w-full items-center gap-2.5 rounded-(--r-chip) px-2 py-1.5 text-left hover:bg-muted"
  onclick={onopen}
>
  <StatusGlyph status={agent.pane.status} />
  <span class="flex min-w-0 flex-1 flex-col gap-px">
    <span class="mono text-[13.5px] font-medium"
      >{agent.pane.label} <span class="text-[11px] text-muted-foreground">{agent.pane.id}</span></span
    >
    <span class="truncate text-[11.5px] text-(--text-3b)">{sub}</span>
  </span>
  <span class="mono flex-none text-[11px] {wordColor[agent.pane.status] ?? 'text-(--text-3)'}" data-status={agent.pane.status}
    >{agent.pane.status}</span
  >
</button>
