<script lang="ts">
  import { session } from '$lib/session/live';
  import { draft, config, showToast } from '$lib/ui/state';
  let { paneId, blocked }: { paneId: string; blocked: boolean } = $props();
  const s = session();
  let ta: HTMLTextAreaElement | undefined = $state();

  function autogrow() {
    if (!ta) return;
    ta.style.height = 'auto';
    ta.style.height = Math.min(ta.scrollHeight, 96) + 'px';
  }
  function send(text?: string) {
    const t = (text ?? $draft).trim();
    if (!t) return;
    // Clear the composer as soon as the prompt is submitted; agent.prompt's
    // wait resolves only when the agent next goes idle/blocked (long-lived), so
    // the UI must not block on it.
    draft.set('');
    if (ta) ta.style.height = 'auto';
    showToast('prompt sent');
    void s
      .request({ method: 'agent.prompt', params: { target: paneId, text: t, wait: { until: ['idle', 'blocked'], timeout_ms: 900000 } } })
      .catch(() => {});
  }
  const NAV: { k: string; glyph: string; label: string }[] = [
    { k: 'up', glyph: '↑', label: 'up' },
    { k: 'down', glyph: '↓', label: 'down' },
    { k: 'left', glyph: '←', label: 'left' },
    { k: 'right', glyph: '→', label: 'right' },
    { k: 'enter', glyph: '⏎', label: 'enter' },
    { k: 'esc', glyph: 'esc', label: 'esc' },
    { k: 'ctrl+c', glyph: '⌃C', label: 'ctrl+c' }
  ];
  function sendKey(k: string) {
    void s.request({ method: 'agent.send_keys', params: { target: paneId, keys: [k] } }).catch(() => {});
    showToast(`sent ${k} → ${paneId}`);
  }
  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey && !e.isComposing) {
      e.preventDefault();
      void send();
    }
  }
</script>
<div class="composer">
  <div class="keys">
    {#each NAV as n}
      <button class="key mono" aria-label={n.label} title={n.label} onclick={() => sendKey(n.k)}>{n.glyph}</button>
    {/each}
  </div>
  <div class="group">
    <textarea
      bind:this={ta}
      bind:value={$draft}
      oninput={autogrow}
      onkeydown={onKeydown}
      placeholder="Message the agent"
      rows="1"
    ></textarea>
    <div class="foot">
      <span class="hint mono">{blocked ? 'agent blocked — reply or use the keys above' : $config.devCaptions ? 'agent.prompt' : ''}</span>
      <button class="send" class:ready={$draft.trim().length > 0} aria-label="send" onclick={() => send()}>↑</button>
    </div>
  </div>
</div>
<style>
  .composer { border-top: 1px solid var(--hairline); background: var(--app-bg); padding: 8px 14px calc(10px + env(safe-area-inset-bottom)); }
  .group { display: flex; flex-direction: column; border: 1px solid var(--control-input); border-radius: var(--r-composer); padding: 8px 10px; background: var(--code-surface); }
  .keys { display: flex; gap: 6px; overflow-x: auto; padding-bottom: 8px; }
  .key { flex: none; min-width: 34px; height: 34px; padding: 0 10px; border-radius: var(--r-chip); border: 1px solid var(--control); background: var(--card); color: var(--text-2); font-size: 13px; }
  .key:hover { background: var(--surface-tint); }
  textarea { width: 100%; resize: none; border: none; background: none; color: var(--text-1); font-family: var(--font-ui); font-size: 14px; line-height: 1.45; max-height: 96px; }
  textarea:focus { outline: none; }
  .foot { display: flex; align-items: center; gap: 8px; margin-top: 4px; }
  .hint { font-size: 11px; color: var(--text-4); }
  .send { margin-left: auto; width: 34px; height: 34px; border-radius: 50%; border: none; background: var(--control); color: var(--text-on-light); font-size: 16px; }
  .send.ready { background: var(--text-1); }
</style>
