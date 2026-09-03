<script lang="ts">
  import type { Snippet } from 'svelte';
  let { children }: { children: Snippet } = $props();
</script>
<div class="row">{@render children()}</div>
<style>
  .row { display: flex; border-top: 1px solid var(--hairline); }
  .row :global(button) {
    flex: 1; min-height: 44px; background: none; border: none; border-left: 1px solid var(--hairline);
    color: var(--text-2); font-size: 13px; font-weight: 500;
  }
  .row :global(button:first-child) { border-left: none; }
  .row :global(button:hover) { background: var(--surface-tint); }
  .row :global(button.danger) { color: var(--blocked-badge-text); }
</style>
