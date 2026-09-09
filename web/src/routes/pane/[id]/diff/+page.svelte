<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { highlight, escapeHtml } from '$lib/highlight';
  import { wrap, showToast } from '$lib/ui/state';
  import { FIXTURE_DIFFS } from '$lib/transport/fixture';
  import { Button } from '$lib/components/ui/button';

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

  const rowCls: Record<string, string> = {
    add: 'bg-(--diff-add-bg) shadow-[inset_2px_0_0_var(--done)]',
    del: 'bg-(--diff-del-bg) shadow-[inset_2px_0_0_var(--blocked)]',
    hunk: 'text-muted-foreground',
    '': ''
  };

  async function copyPath() {
    await navigator.clipboard?.writeText(file?.path ?? '').catch(() => {});
    copied = true; showToast('copied'); setTimeout(() => (copied = false), 1600);
  }
</script>

<section class="flex h-[calc(100vh/var(--font-scale,1))] flex-col">
  <header class="flex items-center gap-2.5 border-b border-(--hairline) px-3.5 py-2.5">
    <button
      class="h-8 w-8 text-[22px] text-(--text-2)"
      onclick={() => goto(`/pane/${encodeURIComponent(paneId)}`)}
      aria-label="back">‹</button
    >
    <span class="mono min-w-0 flex-1 truncate text-xs text-(--text-2)">{file?.path ?? 'no diff'}</span>
    {#if file}<span class="mono text-[11px]"
        ><span class="text-done">+{file.add}</span> <span class="text-(--blocked-badge-text)">−{file.del}</span>
        {file.lang}</span
      >{/if}
    <button
      class="mono rounded-(--r-badge) border border-border px-2 py-0.5 text-[11px] text-(--text-3)"
      onclick={() => wrap.update((v) => !v)}>wrap {$wrap ? 'on' : 'off'}</button
    >
  </header>

  <div class="flex gap-1.5 overflow-x-auto border-b border-(--hairline) px-3.5 py-2">
    {#each diff.files as f, i}
      <button
        class="mono flex-none rounded-(--r-badge) border px-2 py-1 text-[11px] {i === selected
          ? 'border-ring bg-accent text-foreground'
          : 'border-border bg-card text-(--text-3)'}"
        onclick={() => (selected = i)}
      >
        {f.path.split('/').pop()} <span class="text-done">+{f.add}</span>
        <span class="text-(--blocked-badge-text)">−{f.del}</span>
      </button>
    {/each}
  </div>

  <div class="mono flex-1 overflow-auto py-2 text-[11.5px] leading-[1.7]">
    {#each rows as r, i}
      <div class="flex items-start pl-0.5 {rowCls[r.cls]}">
        <span class="w-[34px] flex-none pr-2.5 text-right text-(--gutter) select-none">{r.cls === 'hunk' ? '' : i + 1}</span>
        <span class="w-[13px] flex-none text-muted-foreground">{r.sign === ' ' ? '' : r.sign}</span>
        <span class={$wrap ? 'whitespace-pre-wrap [word-break:break-word]' : 'whitespace-pre'}>{@html r.html}</span>
      </div>
    {/each}
    {#if rows.length === 0}<div class="p-[30px] text-center text-muted-foreground">no changes</div>{/if}
  </div>

  <footer
    class="flex gap-2.5 bg-[linear-gradient(180deg,transparent,var(--app-bg)_40%)] px-3.5 pt-3 pb-[calc(12px+env(safe-area-inset-bottom))]"
  >
    <Button
      variant="outline"
      class="min-h-[46px] flex-1 rounded-lg font-semibold text-(--text-2)"
      onclick={() => goto(`/pane/${encodeURIComponent(paneId)}`)}>Back to chat</Button
    >
    <Button class="min-h-[46px] flex-1 rounded-lg font-semibold" onclick={copyPath}>{copied ? 'Copied' : 'Copy path'}</Button>
  </footer>
</section>
