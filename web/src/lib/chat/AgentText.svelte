<script lang="ts">
  import { showToast } from '$lib/ui/state';
  let { text }: { text: string } = $props();
  async function copy() { await navigator.clipboard?.writeText(text).catch(() => {}); showToast('copied'); }
</script>
<div class="agent">
  <p class="prose">{text}</p>
  <div class="acts">
    <button class="g" title="copy" onclick={copy}>⎘</button>
    <button class="g" title="retry">↻</button>
  </div>
</div>
<style>
  .agent p { margin: 0; font-size: 13.5px; line-height: 1.6; color: var(--text-2); }
  .acts { display: flex; gap: 6px; margin-top: 6px; }
  .g { width: 30px; height: 30px; border-radius: var(--r-chip); background: none; border: 1px solid transparent; color: var(--text-4); }
  .g:hover { border-color: var(--control); color: var(--text-2); }
</style>
