<script lang="ts">
  import { config } from '$lib/ui/state';
  let { text, caption }: { text: string; caption: string } = $props();
</script>
<div class="wrap">
  <div class="bubble prose">{text}</div>
  {#if $config.devCaptions}<div class="cap mono">{caption}</div>{/if}
</div>
<style>
  .wrap { display: flex; flex-direction: column; align-items: flex-end; }
  .bubble { max-width: 84%; background: var(--text-1); color: var(--text-on-light); padding: 9px 13px; border-radius: 16px 16px 4px 16px; font-size: 13.5px; line-height: 1.5; }
  .cap { font-size: 10.5px; color: var(--text-4); margin-top: 4px; }
</style>
