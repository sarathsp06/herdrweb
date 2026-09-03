<script lang="ts">
  let { summary, body }: { summary: string; body: string } = $props();
  let open = $state(false);
</script>
<div class="reason">
  <button class="head mono" onclick={() => (open = !open)}>{open ? '▾' : '▸'} {summary}</button>
  {#if open}<div class="body prose">{body}</div>{/if}
</div>
<style>
  .head { background: none; border: none; color: var(--text-3b); font-size: 12px; padding: 2px 0; }
  .body { border-left: 2px solid var(--control); padding-left: 12px; margin: 6px 0 2px; color: var(--text-3b); font-size: 13px; }
</style>
