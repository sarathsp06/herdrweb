<script lang="ts">
  import type { SendKey } from '$lib/protocol';
  import { session } from '$lib/session/live';
  import { config, showToast } from '$lib/ui/state';
  let { paneId, question, terminal, age }: { paneId: string; question: string; terminal: string; age: string } = $props();
  const s = session();
  async function answer(key: SendKey) {
    await s.request({ method: 'agent.send_keys', params: { target: paneId, keys: [key] } }).catch(() => {});
    showToast(`sent ${key} → ${paneId}`);
  }
</script>
<div class="blocked">
  <div class="hd">
    <span class="pill mono">blocked</span>
    <span class="age mono">approval requested {age}</span>
  </div>
  <p class="q prose">{question}</p>
  <pre class="term mono">{terminal}</pre>
  <div class="btns">
    <button class="yes" onclick={() => answer('y')}>Yes</button>
    <button class="subtle" onclick={() => answer('a')}>Yes, don't ask</button>
    <button class="no" onclick={() => answer('n')}>No</button>
    <button class="ghost" onclick={() => answer('esc')}>esc</button>
  </div>
  {#if $config.devCaptions}<div class="cap mono">agent.send_keys → {paneId}</div>{/if}
</div>
<style>
  .blocked { background: var(--blocked-grad); border: 1px solid var(--blocked-border); border-radius: var(--r-card); padding: 14px; }
  .hd { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
  .pill { font-size: 10.5px; font-weight: 600; color: var(--blocked-badge-text); border: 1px solid var(--blocked-border); border-radius: var(--r-badge); padding: 2px 7px; }
  .age { font-size: 11px; color: var(--blocked-text); margin-left: auto; }
  .q { font-size: 13.5px; margin: 0 0 10px; color: var(--text-1); }
  .term { background: var(--code-surface); border: 1px solid var(--hairline); border-radius: var(--r-chip); padding: 10px 12px; font-size: 11.5px; color: var(--text-2b); white-space: pre-wrap; margin: 0 0 12px; }
  .btns { display: flex; gap: 8px; flex-wrap: wrap; }
  .btns button { min-height: 44px; padding: 0 16px; border-radius: var(--r-btn); font-weight: 600; font-size: 13.5px; border: 1px solid var(--control); background: none; color: var(--text-2); }
  .yes { background: var(--text-1); color: var(--text-on-light); border-color: transparent; }
  .no { color: var(--blocked-badge-text); border-color: var(--blocked-border); }
  .subtle { background: var(--surface-tint); }
  .ghost { border-color: transparent; color: var(--text-3b); }
</style>
