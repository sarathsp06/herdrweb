<script lang="ts">
  import { sheet, closeSheet, type SheetAction } from '$lib/ui/state';
  let label = $state('');
  let cwd = $state('');
  $effect(() => {
    if ($sheet) { label = ''; cwd = ''; }
  });
  const actions = $derived($sheet && 'actions' in $sheet ? ($sheet.actions as SheetAction[]) : null);
  function confirm() {
    if ($sheet && !('actions' in $sheet)) $sheet.onConfirm(label, cwd);
    closeSheet();
  }
  function pick(a: SheetAction) {
    closeSheet();
    a.onSelect();
  }
</script>
{#if $sheet}
  <div class="scrim" onclick={closeSheet} role="presentation"></div>
  <div class="sheet" role="dialog" aria-label={$sheet.title}>
    <div class="handle"></div>
    <h2>{$sheet.title}</h2>
    {#if actions}
      <ul class="menu" role="menu">
        {#each actions as a (a.label)}
          <li>
            <button class="item" class:danger={a.destructive} class:active={a.active} role="menuitem" onclick={() => pick(a)}>
              {#if a.glyph}<span class="glyph mono" aria-hidden="true">{a.glyph}</span>{/if}
              <span class="ilabel">{a.label}</span>
              {#if a.hint}<span class="hint mono">{a.hint}</span>{/if}
            </button>
          </li>
        {/each}
      </ul>
      <div class="actions">
        <button class="cancel" onclick={closeSheet}>Cancel</button>
      </div>
    {:else if !('actions' in $sheet)}
      <p class="prose body">{$sheet.body}</p>
      {#if $sheet.hasInput}
        <label class="field">
          <span>{$sheet.inputLabel ?? 'Label'}</span>
          <input class="mono" bind:value={label} placeholder="name" />
        </label>
      {/if}
      {#if $sheet.hasCwd}
        <label class="field">
          <span>Working directory</span>
          <input class="mono" bind:value={cwd} placeholder="~/code/project" />
        </label>
      {/if}
      <div class="call mono">{$sheet.call}</div>
      <div class="actions">
        <button class="cancel" onclick={closeSheet}>Cancel</button>
        <button class="cta" class:danger={$sheet.destructive} onclick={confirm}>{$sheet.cta}</button>
      </div>
    {/if}
  </div>
{/if}
<style>
  .scrim { position: fixed; inset: 0; background: rgba(0, 0, 0, 0.55); z-index: 70; }
  .sheet {
    position: fixed; left: 0; right: 0; bottom: 0; z-index: 71; margin: 0 auto; max-width: 560px;
    background: var(--raised); border: 1px solid var(--hairline);
    border-radius: 22px 22px 34px 34px; padding: 10px 18px calc(18px + env(safe-area-inset-bottom));
  animation: hsheet 0.22s cubic-bezier(0.32, 0.72, 0, 1);
  }
  .handle { width: 36px; height: 4px; border-radius: 999px; background: var(--control); margin: 6px auto 14px; }
  h2 { font-size: 16px; font-weight: 600; margin: 0 0 6px; }
  .body { color: var(--text-3); font-size: 13px; margin: 0 0 14px; }
  .field { display: block; margin-bottom: 12px; }
  .field span { display: block; font-size: 11px; color: var(--text-3b); margin-bottom: 6px; }
  .field input {
    width: 100%; background: var(--code-surface); border: 1px solid var(--control-input);
    border-radius: var(--r-chip); padding: 10px 12px; color: var(--text-1); font-size: 14px;
  }
  .call { font-size: 11px; color: var(--text-4); background: var(--code-surface); border: 1px solid var(--hairline); border-radius: var(--r-chip); padding: 8px 10px; margin-bottom: 16px; }
  .actions { display: flex; gap: 10px; }
  .actions button { flex: 1; min-height: 46px; border-radius: var(--r-btn); font-weight: 600; font-size: 14px; border: 1px solid var(--control); background: none; color: var(--text-2); }
  .cta { background: var(--text-1); color: var(--text-on-light); border-color: transparent; }
  .cta.danger { background: var(--blocked); color: var(--text-1); }
  .menu { list-style: none; margin: 6px 0 14px; padding: 0; max-height: 50dvh; overflow-y: auto; display: flex; flex-direction: column; gap: 2px; }
  .item {
    width: 100%; display: flex; align-items: center; gap: 10px; min-height: 48px;
    padding: 0 12px; text-align: left; border: none; border-radius: var(--r-chip);
    background: none; color: var(--text-1); font-size: 14px;
  }
  .item:hover { background: var(--surface-tint); }
  .item.active { background: var(--surface-tint-2); }
  .item.danger { color: var(--blocked-badge-text); }
  .glyph { flex: none; width: 20px; text-align: center; color: var(--text-3); }
  .item.danger .glyph { color: inherit; }
  .ilabel { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .hint { flex: none; font-size: 11px; color: var(--text-4); }
</style>
