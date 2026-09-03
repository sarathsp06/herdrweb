<script lang="ts">
  let { name, arg, result, ok, output }: { name: string; arg: string; result: string; ok: boolean; output: string } = $props();
  let open = $state(false);
</script>
<div class="tool">
  <button class="head" onclick={() => (open = !open)}>
    <span class="chev mono">{open ? '▾' : '▸'}</span>
    <span class="name mono">{name}</span>
    <span class="arg mono">{arg}</span>
    <span class="res mono" class:ok>{result}</span>
  </button>
  {#if open}<pre class="out mono">{output}</pre>{/if}
</div>
<style>
  .tool { border: 1px solid var(--hairline); border-radius: var(--r-chip); background: var(--card); }
  .head { width: 100%; display: flex; align-items: center; gap: 8px; padding: 8px 10px; background: none; border: none; text-align: left; }
  .chev { color: var(--text-4); font-size: 11px; }
  .name { font-size: 11.5px; color: var(--text-2); background: var(--surface-tint); padding: 1px 6px; border-radius: var(--r-badge); }
  .arg { font-size: 11.5px; color: var(--text-3b); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; flex: 1; }
  .res { font-size: 11px; color: var(--text-3); margin-left: auto; }
  .res.ok { color: var(--done); }
  .out { margin: 0; padding: 8px 12px; border-top: 1px solid var(--hairline); font-size: 11.5px; color: var(--text-3); white-space: pre-wrap; }
</style>
