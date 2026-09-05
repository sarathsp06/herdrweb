<script lang="ts">
  import { session } from '$lib/session/live';
  import { draft, config, showToast } from '$lib/ui/state';
  import type { Call } from '$lib/protocol';
  import { filterSlash } from './slash';
  import { uploadImage } from './upload';
  let { paneId, blocked, agent = true }: { paneId: string; blocked: boolean; agent?: boolean } = $props();
  const s = session();
  let ta: HTMLTextAreaElement | undefined = $state();

  // ---- Slash-command palette (agent panes only) ----------------------------
  // Typing "/foo" surfaces a filtered command list. Arrow keys move the
  // highlight, Tab/Enter accept (filling the draft, never auto-sending), Esc
  // dismisses. The palette only edits the draft — send stays a deliberate act.
  let slashSel = $state(0);
  let slashDismissed = $state(false);
  const slashMatches = $derived(agent ? filterSlash($draft) : []);
  const slashOpen = $derived(slashMatches.length > 0 && !slashDismissed);
  // Reset the highlight to the top whenever the match set changes (on typing);
  // arrow navigation mutates slashSel without touching slashMatches.
  $effect(() => {
    slashMatches;
    slashSel = 0;
  });
  function acceptSlash(cmd: string) {
    draft.set(cmd + ' ');
    slashDismissed = false;
    queueMicrotask(() => {
      ta?.focus();
      autogrow();
    });
  }

  // ---- Image attach (agent panes only) -------------------------------------
  // Coding agents read images by path, so an uploaded/pasted image is written
  // host-side by the bridge and its absolute path dropped into the draft — the
  // operator reviews and sends. Attach via the button or paste into the box.
  let fileInput: HTMLInputElement | undefined = $state();
  function insertIntoDraft(text: string) {
    draft.update((d) => d + (d && !d.endsWith(' ') ? ' ' : '') + text + ' ');
    queueMicrotask(() => {
      ta?.focus();
      autogrow();
    });
  }
  async function attachImage(file: Blob) {
    showToast('uploading image…');
    try {
      insertIntoDraft(await uploadImage(file));
      showToast('image attached');
    } catch {
      showToast('image upload failed');
    }
  }
  function onPickFile(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (file && file.type.startsWith('image/')) void attachImage(file);
    input.value = '';
  }
  function onPaste(e: ClipboardEvent) {
    if (!agent) return;
    const item = Array.from(e.clipboardData?.items ?? []).find((i) => i.type.startsWith('image/'));
    const file = item?.getAsFile();
    if (!file) return;
    e.preventDefault();
    void attachImage(file);
  }

  function autogrow() {
    slashDismissed = false;
    if (!ta) return;
    ta.style.height = 'auto';
    ta.style.height = Math.min(ta.scrollHeight, 96) + 'px';
  }
  function send(text?: string) {
    const t = (text ?? $draft).trim();
    if (!t) return;
    // Clear the composer as soon as input is submitted. For agents, agent.prompt's
    // wait resolves only when the agent next goes idle/blocked (long-lived), so the
    // UI must not block on it. Terminals are not agents: type the literal text then
    // press Enter via the pane.* API (agent.* is a silent no-op on a plain pane).
    draft.set('');
    if (ta) ta.style.height = 'auto';
    if (agent) {
      showToast('prompt sent');
      void s
        .request({ method: 'agent.prompt', params: { target: paneId, text: t, wait: { until: ['idle', 'blocked'], timeout_ms: 900000 } } })
        .catch(() => {});
      return;
    }
    showToast('sent to terminal');
    void s.request({ method: 'pane.send_text', params: { pane_id: paneId, text: t } }).catch(() => {});
    void s.request({ method: 'pane.send_keys', params: { pane_id: paneId, keys: ['enter'] } }).catch(() => {});
  }
  const NAV: { k: string; glyph: string; label: string }[] = [
    { k: 'up', glyph: '↑', label: 'up' },
    { k: 'down', glyph: '↓', label: 'down' },
    { k: 'left', glyph: '←', label: 'left' },
    { k: 'right', glyph: '→', label: 'right' },
    { k: 'tab', glyph: '⇥', label: 'tab' },
    { k: 'shift+tab', glyph: '⇧⇥', label: 'shift+tab' },
    { k: 'enter', glyph: '⏎', label: 'enter' },
    { k: 'esc', glyph: 'esc', label: 'esc' },
    { k: 'ctrl+c', glyph: '⌃C', label: 'ctrl+c' },
    { k: 'ctrl+d', glyph: '⌃D', label: 'ctrl+d' }
  ];
  function sendKey(k: string) {
    const call: Call = agent
      ? { method: 'agent.send_keys', params: { target: paneId, keys: [k] } }
      : { method: 'pane.send_keys', params: { pane_id: paneId, keys: [k] } };
    void s.request(call).catch(() => {});
    showToast(`sent ${k} → ${paneId}`);
  }
  function onKeydown(e: KeyboardEvent) {
    if (slashOpen) {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        slashSel = (slashSel + 1) % slashMatches.length;
        return;
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault();
        slashSel = (slashSel - 1 + slashMatches.length) % slashMatches.length;
        return;
      }
      if (e.key === 'Tab' || (e.key === 'Enter' && !e.shiftKey && !e.isComposing)) {
        e.preventDefault();
        acceptSlash(slashMatches[slashSel].cmd);
        return;
      }
      if (e.key === 'Escape') {
        e.preventDefault();
        slashDismissed = true;
        return;
      }
    }
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
  {#if slashOpen}
    <ul class="slash" role="listbox" aria-label="slash commands">
      {#each slashMatches as m, i}
        <li>
          <button
            type="button"
            class="slashrow"
            class:sel={i === slashSel}
            role="option"
            aria-selected={i === slashSel}
            onclick={() => acceptSlash(m.cmd)}
          >
            <span class="scmd mono">{m.cmd}</span>
            <span class="sdesc">{m.desc}</span>
          </button>
        </li>
      {/each}
    </ul>
  {/if}
  <div class="group">
    <textarea
      bind:this={ta}
      bind:value={$draft}
      oninput={autogrow}
      onpaste={onPaste}
      onkeydown={onKeydown}
      placeholder={agent ? 'Message the agent' : 'Type into terminal'}
      rows="1"
    ></textarea>
    <div class="foot">
      {#if agent}
        <button class="attach" type="button" aria-label="attach image" title="attach image" onclick={() => fileInput?.click()}>
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="3" y="3" width="18" height="18" rx="2" /><circle cx="8.5" cy="8.5" r="1.5" /><path d="m21 15-5-5L5 21" /></svg>
        </button>
        <input bind:this={fileInput} class="hidden-file" type="file" accept="image/*" onchange={onPickFile} />
      {/if}
      <span class="hint mono">{blocked ? 'agent blocked — reply or use the keys above' : $config.devCaptions ? (agent ? 'agent.prompt' : 'pane.send_text') : ''}</span>
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
  .attach { flex: none; width: 34px; height: 34px; display: inline-flex; align-items: center; justify-content: center; border-radius: 50%; border: 1px solid var(--control); background: var(--card); color: var(--text-2); padding: 0; }
  .attach:hover { background: var(--surface-tint); }
  .hidden-file { display: none; }
  .slash { list-style: none; margin: 0 0 8px; padding: 4px; display: flex; flex-direction: column; gap: 2px; border: 1px solid var(--control-input); border-radius: var(--r-composer); background: var(--code-surface); max-height: 180px; overflow-y: auto; }
  .slashrow { display: flex; align-items: baseline; gap: 10px; width: 100%; text-align: left; padding: 7px 10px; border: none; border-radius: var(--r-chip); background: none; color: var(--text-1); }
  .slashrow.sel { background: var(--surface-tint); }
  .scmd { flex: none; color: var(--text-1); font-size: 13px; }
  .sdesc { color: var(--text-4); font-size: 12px; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
</style>
