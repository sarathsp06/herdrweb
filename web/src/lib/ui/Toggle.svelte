<script lang="ts">
  let { checked = false, onchange }: { checked?: boolean; onchange?: (v: boolean) => void } = $props();
</script>
<button class="track" class:on={checked} role="switch" aria-label="toggle" aria-checked={checked} onclick={() => onchange?.(!checked)}>
  <span class="knob"></span>
</button>
<style>
  .track {
    width: 38px; height: 22px; border-radius: 999px; background: var(--control); position: relative;
    border: none; padding: 0; flex: none; transition: background 0.16s;
  }
  .track.on { background: var(--done); }
  .knob {
    position: absolute; top: 2px; left: 2px; width: 18px; height: 18px; border-radius: 50%;
    background: var(--text-1); transition: left 0.16s;
  }
  .track.on .knob { left: 18px; background: var(--text-on-light); }
</style>
