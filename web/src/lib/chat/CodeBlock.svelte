<script lang="ts">
  import { highlight } from '$lib/highlight';
  import { showToast } from '$lib/ui/state';
  let { path, lang, code }: { path: string; lang: string; code: string } = $props();
  let html = $state('');
  let copied = $state(false);
  $effect(() => { highlight(code, lang).then((h) => (html = h)); });
  const lines = $derived(code.split('\n'));
  async function copy() {
    await navigator.clipboard?.writeText(code).catch(() => {});
    copied = true; showToast('copied'); setTimeout(() => (copied = false), 1600);
  }
</script>
<div class="code">
  <header>
    <span class="mono path">{path}</span>
    <span class="mono lang">{lang}</span>
    <button class="copy mono" onclick={copy}>{copied ? 'copied' : 'copy'}</button>
  </header>
  <div class="body mono">
    <div class="gutter" aria-hidden="true">{#each lines as _, i}<span>{i + 1}</span>{/each}</div>
    <pre class="src">{#if html}{@html html}{:else}{code}{/if}</pre>
  </div>
</div>
<style>
  .code { background: var(--code-surface); border: 1px solid var(--hairline); border-radius: var(--r-chip); overflow: hidden; }
  header { display: flex; align-items: center; gap: 8px; padding: 7px 10px; border-bottom: 1px solid var(--hairline); }
  .path { font-size: 11px; color: var(--text-3); }
  .lang { font-size: 10.5px; color: var(--text-4); }
  .copy { margin-left: auto; font-size: 10.5px; color: var(--text-3); background: none; border: 1px solid var(--control); border-radius: var(--r-badge); padding: 1px 7px; }
  .body { display: flex; overflow-x: auto; font-size: 11.5px; line-height: 1.7; }
  .gutter { display: flex; flex-direction: column; text-align: right; color: var(--gutter); padding: 8px 8px 8px 10px; width: 32px; flex: none; user-select: none; }
  .src { margin: 0; padding: 8px 12px; white-space: pre; }
  .src :global(span) { white-space: pre; }
</style>
