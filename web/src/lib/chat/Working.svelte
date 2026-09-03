<script lang="ts">
  let { label }: { label: string } = $props();
</script>
<div class="work">
  <span class="dots"><span></span><span></span><span></span></span>
  <span class="mono">{label}</span>
</div>
<style>
  .work { display: flex; align-items: center; gap: 10px; color: var(--text-3b); font-size: 12px; }
  .dots { display: inline-flex; gap: 4px; }
  .dots span { width: 5px; height: 5px; border-radius: 50%; background: var(--working); animation: hbounce 1.2s ease-in-out infinite; }
  .dots span:nth-child(2) { animation-delay: 0.15s; }
  .dots span:nth-child(3) { animation-delay: 0.3s; }
</style>
