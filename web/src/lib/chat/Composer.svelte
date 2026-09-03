<script lang="ts">
  import { session } from '$lib/session/live';
  import { draft, config, showToast } from '$lib/ui/state';
  let { paneId, blocked }: { paneId: string; blocked: boolean } = $props();
  const s = session();
  let ta: HTMLTextAreaElement | undefined = $state();

  const chips = $derived(
    blocked ? ['y', 'n', 'esc', 'ctrl+c', '/status'] : ['continue', 'run tests', 'git diff', 'esc', 'ctrl+c']
  );

  function autogrow() {
    if (!ta) return;
    ta.style.height = 'auto';
    ta.style.height = Math.min(ta.scrollHeight, 96) + 'px';
  }
  async function chip(c: string) {
    if (['y', 'n', 'esc', 'ctrl+c', 'a'].includes(c)) {
      await s.request({ method: 'agent.send_keys', params: { target: paneId, keys: c } }).catch(() => {});
      showToast(`sent ${c} → ${paneId}`);
    } else {
      await send(c);
    }
  }
  async function send(text?: string) {
    const t = (text ?? $draft).trim();
    if (!t) return;
    await s
      .request({ method: 'agent.prompt', params: { target: paneId, text: t, wait: { until: ['idle', 'blocked'], timeout_ms: 900000 } } })
      .catch(() => {});
    draft.set('');
    if (ta) ta.style.height = 'auto';
    showToast('prompt sent');
  }
</script>
<div class="composer">
  <div class="chips">
    {#each chips as c}
      <button class="chip mono" onclick={() => chip(c)}>{c}</button>
    {/each}
  </div>
  <div class="group">
    <textarea
      bind:this={ta}
      bind:value={$draft}
      oninput={autogrow}
      placeholder="Message the agent"
      rows="1"
    ></textarea>
    <div class="foot">
      <span class="hint mono">{blocked ? 'blocked — answer above' : $config.devCaptions ? 'agent.prompt' : ''}</span>
      <button class="send" class:ready={$draft.trim().length > 0} aria-label="send" onclick={() => send()}>↑</button>
    </div>
  </div>
</div>
<style>
  .composer { border-top: 1px solid var(--hairline); background: var(--app-bg); padding: 8px 14px calc(10px + env(safe-area-inset-bottom)); }
  .chips { display: flex; gap: 6px; overflow-x: auto; padding-bottom: 8px; }
  .chip { flex: none; min-height: 34px; padding: 0 12px; border-radius: var(--r-chip); border: 1px solid var(--control); background: var(--card); color: var(--text-2); font-size: 12px; }
  .group { display: flex; flex-direction: column; border: 1px solid var(--control-input); border-radius: var(--r-composer); padding: 8px 10px; background: var(--code-surface); }
  textarea { width: 100%; resize: none; border: none; background: none; color: var(--text-1); font-family: var(--font-ui); font-size: 14px; line-height: 1.45; max-height: 96px; }
  textarea:focus { outline: none; }
  .foot { display: flex; align-items: center; gap: 8px; margin-top: 4px; }
  .hint { font-size: 11px; color: var(--text-4); }
  .send { margin-left: auto; width: 34px; height: 34px; border-radius: 50%; border: none; background: var(--control); color: var(--text-on-light); font-size: 16px; }
  .send.ready { background: var(--text-1); }
</style>
