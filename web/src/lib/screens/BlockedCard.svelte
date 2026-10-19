<script lang="ts">
	import type { AgentRef } from '$lib/session/derive';
	import { session } from '$lib/session/live';
	import { showToast } from '$lib/ui/state';
	import { inlineAnswers, now, sendAnswer, waitedFor } from '$lib/chat/answer';

	// A blocked agent in the inbox: what it asks, and — for a short single
	// question — the answers right on the card, so answering never needs the
	// pane. compact (the desktop sidebar) shows the question only.
	let { agent, onopen, compact = false }: { agent: AgentRef; onopen: () => void; compact?: boolean } = $props();
	const s = session();
	const prompt = $derived(agent.pane.prompt);
	const answers = $derived(compact ? null : inlineAnswers(prompt));
	let sent = $state('');
	// A new question (or the agent moving on) clears the "sent" note.
	$effect(() => {
		prompt?.question;
		sent = '';
	});

	function answer(label: string, keys: string[]) {
		sendAnswer((c) => s.request(c), agent.pane.id, keys).then(
			() => (sent = label),
			() => showToast(`${label} not sent`)
		);
	}
</script>

<article
	class="flex flex-col gap-2.5 rounded-(--r-card) border border-(--blocked-border) bg-(image:--blocked-grad) {compact ? 'p-2.5' : 'p-3'}"
>
	<button class="flex flex-col gap-1.5 text-left" onclick={onopen} aria-label="{agent.pane.label} {agent.pane.id}">
		<span class="flex items-baseline gap-2">
			<span class="mono text-[13.5px] font-semibold text-foreground">{agent.pane.label}</span>
			<span class="mono text-[11px] text-muted-foreground">{agent.pane.id}</span>
			<span class="ml-auto text-[11px] text-muted-foreground">{waitedFor(agent.pane.blockedSince, $now)}</span>
		</span>
		<span class="text-pretty text-foreground {compact ? 'line-clamp-2 text-[12.5px] leading-snug' : 'line-clamp-3 text-[14px] leading-[1.45]'}"
			>{prompt?.question || 'Waiting for you'}</span
		>
		<span class="mono truncate text-[11px] text-muted-foreground"
			>{agent.space.label}{agent.space.branch ? ' · ' + agent.space.branch : ''}{prompt?.tabs?.length
				? ` · ${prompt.tabs.length - 1} questions`
				: ''}</span
		>
	</button>
	{#if sent}
		<div class="mono flex min-h-11 items-center gap-2 text-[12px] text-(--text-2)" role="status">
			<span class="text-done">✓</span>Sent {sent}
		</div>
	{:else if answers}
		<div class="grid gap-2" style="grid-template-columns: repeat({answers.length + 1}, minmax(0, 1fr))">
			{#each answers as o, i (o.label)}
				<button
					class="h-11 truncate rounded-(--r-btn) px-2 text-[13.5px] font-semibold {i === 0
						? 'bg-primary text-primary-foreground'
						: 'border border-border bg-card text-(--text-2)'}"
					title={o.label}
					onclick={() => answer(o.label, o.keys)}>{o.label}</button
				>
			{/each}
			<button class="h-11 rounded-(--r-btn) border border-border bg-card text-[13.5px] font-semibold text-(--text-2)" onclick={onopen}>Open</button>
		</div>
	{/if}
</article>
