<script lang="ts">
	import type { Status } from '$lib/protocol';
	let { status }: { status: Status } = $props();
	// Each status owns its text and border colour outright: two border-colour
	// utilities on one element resolve by stylesheet order, not class order, so
	// the neutral defaults apply only when no status colour does. A gradient
	// variable needs the `image:` hint or Tailwind emits it as a colour.
	const cls: Partial<Record<Status, string>> = {
		blocked: 'text-(--blocked-badge-text) border-(--blocked-border) bg-(image:--blocked-grad)',
		working: 'text-working border-[color-mix(in_oklch,var(--working)_40%,transparent)]',
		done: 'text-done border-[color-mix(in_oklch,var(--done)_40%,transparent)]'
	};
</script>

<span
	class="mono rounded-(--r-badge) border px-[7px] py-0.5 text-[10.5px] font-semibold tracking-[0.02em] whitespace-nowrap lowercase {cls[status] ??
		'border-(--control) text-(--text-3)'}"
	data-status={status}
>{status}</span>
