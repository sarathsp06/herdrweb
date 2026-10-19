<script lang="ts">
	import { session } from '$lib/session/live';
	import { pokeScrollback } from '$lib/session/scrollback';
	import { showToast } from '$lib/ui/state';
	import type { Prompt, PromptOption } from '$lib/protocol';
	import { answered, keyHint, now, recordAnswer, sendAnswer, waitedFor } from './answer';

	// The answer panel a blocked agent pane shows in place of the key row: the
	// question the bridge read off the screen and one button per answer. Every
	// button sends real keys, so it only offers what was parsed; when nothing
	// was, it says so and falls back to Enter / esc / ⌃C.
	let {
		paneId,
		label,
		prompt,
		blockedSince,
		keysOpen,
		ontogglekeys,
		onwritein
	}: {
		paneId: string;
		label: string;
		prompt?: Prompt;
		blockedSince?: number;
		keysOpen: boolean;
		ontogglekeys: () => void;
		onwritein: () => void;
	} = $props();

	const s = session();
	const options = $derived(prompt?.nav ? prompt.options : []);
	const yesNo = $derived(prompt?.nav === 'keys' && options.length === 2);
	// Multi-question forms show one question at a time; the answers picked so
	// far are remembered client-side to draw progress and the review step.
	const done = $derived($answered[paneId] ?? []);
	const questions = $derived(prompt?.tabs?.filter((t) => t.toLowerCase() !== 'submit') ?? []);
	const reviewing = $derived(!options.length && done.length > 0);

	function send(keys: string[], what: string) {
		sendAnswer((c) => s.request(c), paneId, keys).then(
			() => {
				showToast(`sent ${what} to ${label}`);
				pokeScrollback(paneId);
			},
			() => showToast(`${what} not sent`)
		);
	}
	function pick(o: PromptOption) {
		if (prompt?.tabs?.length) recordAnswer(paneId, prompt.question, o.label);
		send(o.keys, o.label);
		if (o.needsText) onwritein();
	}
</script>

<section
	class="mb-2 flex max-h-[62dvh] flex-col gap-2.5 overflow-y-auto rounded-(--r-card) border border-(--blocked-border) bg-(image:--blocked-grad) p-3"
	aria-label="answer {label}"
>
	<div class="flex items-center gap-2">
		<span class="size-[7px] flex-none rounded-[2px] bg-blocked"></span>
		<span class="text-[12.5px] font-semibold text-(--blocked-badge-text)"
			>{reviewing ? `Review your answers for ${label}` : `${label} is asking`}</span
		>
		<span class="mono ml-auto text-[11px] text-muted-foreground">{waitedFor(blockedSince, $now)}</span>
	</div>

	{#if questions.length > 1}
		<ol class="m-0 flex list-none gap-1.5 overflow-x-auto p-0" aria-label="questions">
			{#each questions as q, i (q)}
				<li
					class="mono flex h-7 flex-none items-center gap-1.5 rounded-(--r-chip) border px-2 text-[11.5px] {i === done.length
						? 'border-foreground bg-foreground text-background'
						: 'border-border text-(--text-2)'}"
					aria-current={i === done.length ? 'step' : undefined}
				>
					<span class={i < done.length ? 'text-done' : ''}>{i < done.length ? '✓' : i + 1}</span>{q}
				</li>
			{/each}
			<li class="mono flex h-7 flex-none items-center rounded-(--r-chip) border border-dashed border-border px-2 text-[11.5px] text-(--text-3)">
				Submit
			</li>
		</ol>
	{/if}

	{#if reviewing}
		<ul class="m-0 flex list-none flex-col rounded-(--r-chip) border border-border bg-card p-0">
			{#each done as a, i (a.question)}
				<li class="flex flex-col gap-0.5 px-3 py-2 {i ? 'border-t border-(--hairline)' : ''}">
					<span class="text-[11.5px] text-muted-foreground">{a.question}</span>
					<span class="text-[13.5px] font-medium text-foreground">{a.answer}</span>
				</li>
			{/each}
		</ul>
		<div class="flex gap-2">
			<button class="h-11 rounded-(--r-btn) border border-border bg-card px-4 text-sm font-semibold text-(--text-2)" onclick={() => send(['left'], 'back')}
				>Back</button
			>
			<button class="h-11 flex-1 rounded-(--r-btn) bg-primary text-sm font-semibold text-primary-foreground" onclick={() => send(['enter'], 'submit')}
				>Submit {done.length} {done.length === 1 ? 'answer' : 'answers'}</button
			>
		</div>
	{:else if options.length}
		<p class="m-0 text-[14.5px] leading-[1.45] text-pretty text-foreground">{prompt?.question}</p>
		{#if yesNo}
			<div class="grid grid-cols-2 gap-2">
				<button class="h-12 rounded-(--r-btn) bg-primary text-[15px] font-semibold text-primary-foreground" onclick={() => pick(options[0])}
					>{options[0].label} <span class="mono text-[11px] opacity-60">{options[0].keys.join(' ')}</span></button
				>
				<button
					class="h-12 rounded-(--r-btn) border border-(--blocked-border) text-[15px] font-semibold text-(--blocked-badge-text)"
					onclick={() => pick(options[1])}
					>{options[1].label} <span class="mono text-[11px] opacity-60">{options[1].keys.join(' ')}</span></button
				>
			</div>
		{:else}
			<div class="flex flex-col gap-1.5" role="group" aria-label="answers">
				{#each options as o (o.label)}
					<button
						class="flex min-h-12 w-full items-start gap-3 rounded-(--r-btn) border border-border bg-card px-3 py-2.5 text-left hover:bg-muted"
						onclick={() => pick(o)}
					>
						<span class="flex min-w-0 flex-1 flex-col gap-0.5">
							<span class="flex flex-wrap items-baseline gap-2">
								<span class="text-[14px] font-semibold text-foreground">{o.label}</span>
								{#if o.recommended}<span class="mono rounded-(--r-badge) bg-[color-mix(in_oklch,var(--done)_16%,transparent)] px-1.5 py-px text-[10.5px] text-done"
										>recommended</span
									>{/if}
							</span>
							{#if o.desc}<span class="text-[12px] leading-[1.4] text-muted-foreground">{o.desc}</span>{/if}
							{#if o.needsText}<span class="text-[12px] text-muted-foreground">Then type your answer below.</span>{/if}
						</span>
						{#if keyHint(prompt!, o.keys)}<span class="mono flex-none pt-0.5 text-[11px] text-muted-foreground">{keyHint(prompt!, o.keys)}</span>{/if}
					</button>
				{/each}
			</div>
		{/if}
	{:else}
		<p class="m-0 text-[13px] leading-[1.45] text-(--text-2)">
			Couldn't read the question. Check the terminal above and answer with the keys below.
		</p>
		<button class="h-11 rounded-(--r-btn) bg-primary text-sm font-semibold text-primary-foreground" onclick={() => send(['enter'], 'enter')}>Enter ⏎</button>
	{/if}

	<div class="flex gap-1.5">
		<button class="mono h-11 min-w-11 rounded-(--r-chip) border border-border bg-card px-2.5 text-sm text-(--text-2)" aria-label="esc" onclick={() => send(['esc'], 'esc')}>esc</button>
		<button class="mono h-11 min-w-11 rounded-(--r-chip) border border-border bg-card px-2.5 text-sm text-(--text-2)" aria-label="ctrl+c" onclick={() => send(['ctrl+c'], '⌃C')}>⌃C</button>
		<button
			class="ml-auto h-11 rounded-(--r-chip) border border-border bg-card px-3 text-[13px] font-semibold text-(--text-2)"
			aria-expanded={keysOpen}
			onclick={ontogglekeys}>{keysOpen ? 'Hide keys' : 'All keys'}</button
		>
	</div>
</section>
