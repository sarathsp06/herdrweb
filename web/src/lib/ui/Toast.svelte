<script lang="ts">
  import { toast } from '$lib/ui/state';
</script>
{#if $toast}
  <div class="toast" role="status">
    <span class="dot"></span>
    <span class="mono">{$toast}</span>
  </div>
{/if}
<style>
  .toast {
    position: fixed; left: 14px; right: 14px; bottom: 92px; z-index: 60;
    background: var(--toast-bg); border: 1px solid var(--hairline); border-radius: var(--r-btn);
    padding: 11px 14px; display: flex; gap: 9px; align-items: center; font-size: 12px;
    box-shadow: 0 12px 30px -8px rgba(0, 0, 0, 0.8); max-width: 520px; margin: 0 auto;
  }
  .dot { width: 7px; height: 7px; border-radius: 50%; background: var(--done); flex: none; }
</style>
