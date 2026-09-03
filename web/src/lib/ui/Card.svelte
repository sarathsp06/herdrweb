<script lang="ts">
  import type { Snippet } from 'svelte';
  let { children, onclick, blocked = false, interactive = false }:
    { children: Snippet; onclick?: () => void; blocked?: boolean; interactive?: boolean } = $props();
</script>
{#if onclick}
  <button class="card" class:blocked class:interactive {onclick}>{@render children()}</button>
{:else}
  <div class="card" class:blocked class:interactive>{@render children()}</div>
{/if}
<style>
  .card {
    display: block; width: 100%; text-align: left; background: var(--card);
    border: 1px solid var(--hairline); border-radius: var(--r-card); padding: 12px; color: inherit;
  }
  .interactive:hover { background: var(--card-hover); border-color: var(--control); }
  .blocked { background: var(--blocked-grad); border-color: var(--blocked-border); }
</style>
