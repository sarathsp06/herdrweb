<script lang="ts">
	import { Switch } from '$lib/components/ui/switch';
	let { checked = false, onchange }: { checked?: boolean; onchange?: (v: boolean) => void } = $props();
</script>

<Switch
	{checked}
	aria-label="toggle"
	class="data-checked:bg-done data-unchecked:bg-(--control)"
	onCheckedChange={(v) => onchange?.(v)}
/>
