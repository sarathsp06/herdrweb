<script lang="ts">
  import { goto } from '$app/navigation';
  import type { DiffFile } from '$lib/protocol';
  let { paneId, files, add, del, preview }: { paneId: string; files: DiffFile[]; add: number; del: number; preview: string } = $props();
</script>
<button class="diff" onclick={() => goto(`/pane/${encodeURIComponent(paneId)}/diff`)}>
  <div class="top">
    <span class="glyph mono">◫</span>
    <span class="mono path">{files[0]?.path}</span>
    <span class="count">{files.length} file{files.length === 1 ? '' : 's'} changed</span>
    <span class="delta mono"><span class="add">+{add}</span> <span class="del">−{del}</span></span>
    <span class="chev mono">›</span>
  </div>
  <div class="preview mono">{preview}</div>
</button>
<style>
  .diff { width: 100%; text-align: left; background: var(--card); border: 1px solid var(--hairline); border-radius: var(--r-chip); overflow: hidden; }
  .top { display: flex; align-items: center; gap: 8px; padding: 10px 12px; }
  .glyph { color: var(--text-3); }
  .path { font-size: 12px; color: var(--text-2); }
  .count { font-size: 11px; color: var(--text-4); }
  .delta { font-size: 11px; margin-left: auto; }
  .add { color: var(--done); } .del { color: var(--blocked-badge-text); }
  .chev { color: var(--text-4); }
  .preview { font-size: 11px; color: var(--text-3b); padding: 6px 12px; border-top: 1px solid var(--hairline); background: var(--code-surface); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
</style>
