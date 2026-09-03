<script lang="ts">
  import type { Rollup, Status } from '$lib/protocol';
  let { status }: { status: Status | Rollup } = $props();
  const color: Record<string, string> = {
    working: 'var(--working)', blocked: 'var(--blocked)', done: 'var(--done)',
    idle: 'var(--idle)', unknown: 'var(--unknown)', none: 'var(--unknown)'
  };
</script>
<span class="dot" style="--c: {color[status] ?? 'var(--unknown)'}" data-status={status} aria-hidden="true"></span>
<style>
  .dot { width: 8px; height: 8px; border-radius: 50%; background: var(--c); display: inline-block; flex: none; }
  .dot[data-status='idle'], .dot[data-status='unknown'], .dot[data-status='none'] {
    background: transparent; box-shadow: inset 0 0 0 1.5px var(--c);
  }
</style>
