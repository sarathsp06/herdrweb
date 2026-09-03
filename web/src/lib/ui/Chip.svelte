<script lang="ts">
  let { text, selected = false }: { text: string; selected?: boolean } = $props();
</script>
<span class="chip mono" class:selected>{text}</span>
<style>
  .chip {
    font-size: 11px; padding: 2px 6px; border-radius: var(--r-badge);
    border: 1px solid var(--control); color: var(--text-3); white-space: nowrap;
  }
  .selected { border-color: var(--control-selected-2); background: var(--surface-tint-2); color: var(--text-2); }
</style>
