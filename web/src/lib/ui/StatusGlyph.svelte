<script lang="ts">
  import type { Rollup, Status } from '$lib/protocol';
  import { glyph } from '$lib/session/derive';
  let { status }: { status: Status | Rollup } = $props();
  const color: Record<string, string> = {
    working: 'var(--working)', blocked: 'var(--blocked)', done: 'var(--done)',
    idle: 'var(--idle)', unknown: 'var(--unknown)', none: 'var(--unknown)'
  };
</script>
<span class="glyph mono" style="color: {color[status] ?? 'var(--unknown)'}" data-status={status} title={status}>{glyph(status)}</span>
<style>
  .glyph { font-size: 12px; width: 16px; text-align: center; flex: none; }
</style>
