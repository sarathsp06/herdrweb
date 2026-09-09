<script lang="ts">
	import { toast } from '$lib/ui/state';
</script>

{#if $toast}
	<div
		class="fixed inset-x-3.5 bottom-[92px] z-60 mx-auto flex max-w-[520px] items-center gap-2 rounded-(--r-btn) border border-hairline bg-(--toast-bg) px-3.5 py-[11px] text-xs shadow-[0_12px_30px_-8px_rgba(0,0,0,0.8)]"
		role="status"
	>
		<span class="size-[7px] flex-none rounded-full bg-done"></span>
		<span class="mono">{$toast}</span>
	</div>
{/if}
