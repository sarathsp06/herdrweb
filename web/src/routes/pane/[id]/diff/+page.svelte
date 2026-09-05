<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { highlight, escapeHtml } from '$lib/highlight';
  import { wrap, showToast } from '$lib/ui/state';
  import { FIXTURE_DIFFS } from '$lib/transport/fixture';

  const paneId = $derived(decodeURIComponent($page.params.id ?? ''));
  const diff = $derived(FIXTURE_DIFFS[paneId] ?? { files: [] });
  let selected = $state(0);
  let copied = $state(false);
  const file = $derived(diff.files[selected]);

  interface Row { sign: string; cls: string; html: string; text: string }
  let rows: Row[] = $state([]);

  $effect(() => {
    const f = file;
    if (!f) { rows = []; return; }
    const lines = f.body.split('\n');
    Promise.all(
      lines.map(async (ln): Promise<Row> => {
        const sign = ln[0] === '+' ? '+' : ln[0] === '-' ? '-' : ln[0] === '@' ? '@' : ' ';
        const cls = sign === '+' ? 'add' : sign === '-' ? 'del' : sign === '@' ? 'hunk' : '';
        const code = sign === ' ' || sign === '@' ? ln : ln.slice(1);
        const html = sign === '@' ? escapeHtml(code) : await highlight(code, f.lang).catch(() => escapeHtml(code));
        return { sign, cls, html, text: ln };
      })
    ).then((r) => (rows = r));
  });

  async function copyPath() {
    await navigator.clipboard?.writeText(file?.path ?? '').catch(() => {});
    copied = true; showToast('copied'); setTimeout(() => (copied = false), 1600);
  }
</script>

<section class="diffview">
  <header class="bar">
    <button class="back" onclick={() => goto(`/pane/${encodeURIComponent(paneId)}`)} aria-label="back">‹</button>
    <span class="path mono">{file?.path ?? 'no diff'}</span>
    {#if file}<span class="delta mono"><span class="add">+{file.add}</span> <span class="del">−{file.del}</span> {file.lang}</span>{/if}
    <button class="wrap mono" onclick={() => wrap.update((v) => !v)}>wrap {$wrap ? 'on' : 'off'}</button>
  </header>

  <div class="strip">
    {#each diff.files as f, i}
      <button class="fchip mono" class:sel={i === selected} onclick={() => (selected = i)}>
        {f.path.split('/').pop()} <span class="a">+{f.add}</span> <span class="d">−{f.del}</span>
      </button>
    {/each}
  </div>

  <div class="body mono" class:wrapon={$wrap}>
    {#each rows as r, i}
      <div class="row {r.cls}">
        <span class="ln">{r.cls === 'hunk' ? '' : i + 1}</span>
        <span class="sign">{r.sign === ' ' ? '' : r.sign}</span>
        <span class="src">{@html r.html}</span>
      </div>
    {/each}
    {#if rows.length === 0}<div class="empty">no changes</div>{/if}
  </div>

  <footer class="foot">
    <button class="btn" onclick={() => goto(`/pane/${encodeURIComponent(paneId)}`)}>Back to chat</button>
    <button class="btn solid" onclick={copyPath}>{copied ? 'Copied' : 'Copy path'}</button>
  </footer>
</section>

<style>
  .diffview { display: flex; flex-direction: column; height: calc(100vh / var(--font-scale, 1)); }
  .bar { display: flex; align-items: center; gap: 10px; padding: 10px 14px; border-bottom: 1px solid var(--hairline); }
  .back { width: 32px; height: 32px; border: none; background: none; color: var(--text-2); font-size: 22px; }
  .path { font-size: 12px; color: var(--text-2); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; flex: 1; }
  .delta { font-size: 11px; } .delta .add { color: var(--done); } .delta .del { color: var(--blocked-badge-text); }
  .wrap { font-size: 11px; color: var(--text-3); background: none; border: 1px solid var(--control); border-radius: var(--r-badge); padding: 2px 8px; }
  .strip { display: flex; gap: 6px; overflow-x: auto; padding: 8px 14px; border-bottom: 1px solid var(--hairline); }
  .fchip { flex: none; font-size: 11px; padding: 4px 8px; border-radius: var(--r-badge); border: 1px solid var(--control); background: var(--card); color: var(--text-3); }
  .fchip.sel { border-color: var(--control-selected-2); background: var(--surface-tint-2); color: var(--text-1); }
  .fchip .a { color: var(--done); } .fchip .d { color: var(--blocked-badge-text); }
  .body { flex: 1; overflow: auto; font-size: 11.5px; line-height: 1.7; padding: 8px 0; }
  .row { display: flex; align-items: flex-start; padding-left: 2px; }
  .row.add { background: var(--diff-add-bg); box-shadow: inset 2px 0 0 var(--done); }
  .row.del { background: var(--diff-del-bg); box-shadow: inset 2px 0 0 var(--blocked); }
  .row.hunk { color: var(--text-4); }
  .ln { width: 34px; text-align: right; color: var(--gutter); flex: none; padding-right: 10px; user-select: none; }
  .sign { width: 13px; color: var(--text-4); flex: none; }
  .src { white-space: pre; }
  .wrapon .src { white-space: pre-wrap; word-break: break-word; }
  .empty { padding: 30px; color: var(--text-4); text-align: center; }
  .foot { display: flex; gap: 10px; padding: 12px 14px calc(12px + env(safe-area-inset-bottom)); background: linear-gradient(180deg, transparent, var(--app-bg) 40%); }
  .btn { flex: 1; min-height: 46px; border-radius: var(--r-btn); border: 1px solid var(--control); background: none; color: var(--text-2); font-weight: 600; }
  .btn.solid { background: var(--text-1); color: var(--text-on-light); border-color: transparent; }
</style>
