<script lang="ts">
  import type { Status } from '$lib/protocol';
  let { status }: { status: Status } = $props();
</script>
<span class="pill mono" data-status={status}>{status}</span>
<style>
  .pill {
    font-size: 10.5px; font-weight: 600; padding: 2px 7px; border-radius: var(--r-badge);
    border: 1px solid var(--control); color: var(--text-3);
    text-transform: lowercase; letter-spacing: 0.02em; white-space: nowrap;
  }
  .pill[data-status='blocked'] { color: var(--blocked-badge-text); border-color: var(--blocked-border); background: var(--blocked-grad); }
  .pill[data-status='working'] { color: var(--working); border-color: color-mix(in oklch, var(--working) 40%, transparent); }
  .pill[data-status='done'] { color: var(--done); border-color: color-mix(in oklch, var(--done) 40%, transparent); }
</style>
