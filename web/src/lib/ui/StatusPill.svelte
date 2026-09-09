<script lang="ts">
	import type { Status } from '$lib/protocol';
	let { status }: { status: Status } = $props();
	const cls: Partial<Record<Status, string>> = {
		blocked: 'text-(--blocked-badge-text) border-(--blocked-border) bg-(--blocked-grad)',
		working: 'text-working border-[color-mix(in_oklch,var(--working)_40%,transparent)]',
		done: 'text-done border-[color-mix(in_oklch,var(--done)_40%,transparent)]'
	};
</script>

<span
	class="mono rounded-(--r-badge) border border-(--control) px-[7px] py-0.5 text-[10.5px] font-semibold tracking-[0.02em] whitespace-nowrap lowercase text-(--text-3) {cls[status] ?? ''}"
	data-status={status}
>{status}</span>
