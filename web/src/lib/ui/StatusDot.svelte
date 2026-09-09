<script lang="ts">
	import type { Rollup, Status } from '$lib/protocol';
	let { status }: { status: Status | Rollup } = $props();
	const color: Record<string, string> = {
		working: 'var(--working)', blocked: 'var(--blocked)', done: 'var(--done)',
		idle: 'var(--idle)', unknown: 'var(--unknown)', none: 'var(--unknown)'
	};
	const hollow = $derived(status === 'idle' || status === 'unknown' || status === 'none');
</script>

<span
	class="inline-block size-2 flex-none rounded-full {hollow
		? 'bg-transparent shadow-[inset_0_0_0_1.5px_var(--c)]'
		: 'bg-(--c)'}"
	style="--c: {color[status] ?? 'var(--unknown)'}"
	data-status={status}
	aria-hidden="true"
></span>
