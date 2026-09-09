<script lang="ts">
	import { session } from '$lib/session/live';
	import { findPaneIn } from '$lib/session/derive';
	import { draft, config, showToast } from '$lib/ui/state';
	import type { Call } from '$lib/protocol';
	import { filterSlash, loadSlashCommands, SLASH_COMMANDS, type SlashCommand } from './slash';
	import { uploadImage } from './upload';
	let { paneId, blocked, agent = true, controlMode = false, ontogglecontrol }:
		{ paneId: string; blocked: boolean; agent?: boolean; controlMode?: boolean; ontogglecontrol?: () => void } = $props();
	const s = session();
	const spaces = s.spaces;
	let ta: HTMLTextAreaElement | undefined = $state();

	// ---- Slash-command palette (agent panes only) ----------------------------
	// Typing "/foo" (or tapping the / key) surfaces a filtered command list.
	// Arrow keys move the highlight, Tab/Enter accept (filling the draft, never
	// auto-sending), Esc dismisses. The palette only edits the draft — send
	// stays a deliberate act. Commands are discovered host-side by the bridge
	// (user/project .claude/commands, skills) merged over the curated builtins;
	// the pane's space cwd scopes project commands.
	let slashSel = $state(0);
	let slashDismissed = $state(false);
	let commands: SlashCommand[] = $state(SLASH_COMMANDS);
	const cwd = $derived(findPaneIn($spaces, paneId)?.space.cwd);
	$effect(() => {
		if (!agent) return;
		void loadSlashCommands(cwd).then((c) => (commands = c));
	});
	const slashMatches = $derived(agent ? filterSlash($draft, commands) : []);
	const slashOpen = $derived(slashMatches.length > 0 && !slashDismissed);
	// Reset the highlight to the top whenever the match set changes (on typing);
	// arrow navigation mutates slashSel without touching slashMatches.
	$effect(() => {
		slashMatches;
		slashSel = 0;
	});
	function acceptSlash(cmd: string) {
		draft.set(cmd + ' ');
		slashDismissed = false;
		queueMicrotask(() => {
			ta?.focus();
			autogrow();
		});
	}
	// The "/" key button: start (or return to) a bare-slash draft and focus the
	// box, so the full command list opens without typing on a phone keyboard.
	function openSlash() {
		draft.set('/');
		slashDismissed = false;
		queueMicrotask(() => {
			ta?.focus();
			autogrow();
		});
	}

	// ---- Image attach (agent panes only) -------------------------------------
	// Coding agents read images by path, so an uploaded/pasted image is written
	// host-side by the bridge and its absolute path dropped into the draft — the
	// operator reviews and sends. Attach via the button or paste into the box.
	let fileInput: HTMLInputElement | undefined = $state();
	function insertIntoDraft(text: string) {
		draft.update((d) => d + (d && !d.endsWith(' ') ? ' ' : '') + text + ' ');
		queueMicrotask(() => {
			ta?.focus();
			autogrow();
		});
	}
	async function attachImage(file: Blob) {
		showToast('uploading image…');
		try {
			insertIntoDraft(await uploadImage(file));
			showToast('image attached');
		} catch {
			showToast('image upload failed');
		}
	}
	function onPickFile(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		if (file && file.type.startsWith('image/')) void attachImage(file);
		input.value = '';
	}
	function onPaste(e: ClipboardEvent) {
		if (!agent) return;
		const item = Array.from(e.clipboardData?.items ?? []).find((i) => i.type.startsWith('image/'));
		const file = item?.getAsFile();
		if (!file) return;
		e.preventDefault();
		void attachImage(file);
	}

	function autogrow() {
		slashDismissed = false;
		if (!ta) return;
		ta.style.height = 'auto';
		ta.style.height = Math.min(ta.scrollHeight, 96) + 'px';
	}
	function send(text?: string) {
		const t = (text ?? $draft).trim();
		if (!t) return;
		// Clear the composer as soon as input is submitted — the cleared box and
		// the terminal echoing the input are the feedback; no toast. For agents,
		// agent.prompt's wait resolves only when the agent next goes idle/blocked
		// (long-lived), so the UI must not block on it. Terminals are not agents:
		// type the literal text then press Enter via the pane.* API (agent.* is a
		// silent no-op on a plain pane).
		draft.set('');
		if (ta) ta.style.height = 'auto';
		if (agent) {
			void s
				.request({ method: 'agent.prompt', params: { target: paneId, text: t, wait: { until: ['idle', 'blocked'], timeout_ms: 900000 } } })
				.catch(() => {});
			return;
		}
		void s.request({ method: 'pane.send_text', params: { pane_id: paneId, text: t } }).catch(() => {});
		void s.request({ method: 'pane.send_keys', params: { pane_id: paneId, keys: ['enter'] } }).catch(() => {});
	}
	const NAV: { k: string; glyph: string; label: string }[] = [
		{ k: 'up', glyph: '↑', label: 'up' },
		{ k: 'down', glyph: '↓', label: 'down' },
		{ k: 'left', glyph: '←', label: 'left' },
		{ k: 'right', glyph: '→', label: 'right' },
		{ k: 'tab', glyph: '⇥', label: 'tab' },
		{ k: 'shift+tab', glyph: '⇧⇥', label: 'shift+tab' },
		{ k: 'enter', glyph: '⏎', label: 'enter' },
		{ k: 'esc', glyph: 'esc', label: 'esc' },
		{ k: 'ctrl+c', glyph: '⌃C', label: 'ctrl+c' },
		{ k: 'ctrl+d', glyph: '⌃D', label: 'ctrl+d' }
	];
	function sendKey(k: string) {
		const call: Call = agent
			? { method: 'agent.send_keys', params: { target: paneId, keys: [k] } }
			: { method: 'pane.send_keys', params: { pane_id: paneId, keys: [k] } };
		void s.request(call).catch(() => {});
	}
	function onKeydown(e: KeyboardEvent) {
		if (slashOpen) {
			if (e.key === 'ArrowDown') {
				e.preventDefault();
				slashSel = (slashSel + 1) % slashMatches.length;
				return;
			}
			if (e.key === 'ArrowUp') {
				e.preventDefault();
				slashSel = (slashSel - 1 + slashMatches.length) % slashMatches.length;
				return;
			}
			if (e.key === 'Tab' || (e.key === 'Enter' && !e.shiftKey && !e.isComposing)) {
				e.preventDefault();
				acceptSlash(slashMatches[slashSel].cmd);
				return;
			}
			if (e.key === 'Escape') {
				e.preventDefault();
				slashDismissed = true;
				return;
			}
		}
		if (e.key === 'Enter' && !e.shiftKey && !e.isComposing) {
			e.preventDefault();
			void send();
		}
	}
</script>

<div class="border-t border-hairline bg-background px-3.5 pt-2 pb-[calc(10px+env(safe-area-inset-bottom))]">
	<div class="flex gap-1.5 overflow-x-auto pb-2">
		{#if agent && ontogglecontrol}
			<button
				class="mono h-11 min-w-11 flex-none rounded-(--r-chip) border px-2.5 text-sm {controlMode
					? 'border-(--control-selected-2) bg-accent text-foreground'
					: 'border-border bg-card text-(--text-2) hover:bg-muted'}"
				aria-label="direct control"
				aria-pressed={controlMode}
				title="direct control — swipe the transcript to send arrow keys"
				onclick={ontogglecontrol}
			>{controlMode ? '◉' : '◎'}</button>
		{/if}
		{#if agent}
			<button
				class="mono h-11 min-w-11 flex-none rounded-(--r-chip) border border-border bg-card px-2.5 text-sm text-(--text-2) hover:bg-muted"
				aria-label="slash commands"
				title="slash commands"
				onclick={openSlash}
			>/</button>
		{/if}
		{#each NAV as n (n.k)}
			<button
				class="mono h-11 min-w-11 flex-none rounded-(--r-chip) border border-border bg-card px-2.5 text-sm text-(--text-2) hover:bg-muted"
				aria-label={n.label}
				title={n.label}
				onclick={() => sendKey(n.k)}
			>{n.glyph}</button>
		{/each}
	</div>
	{#if slashOpen}
		<ul class="mb-2 flex max-h-[180px] list-none flex-col gap-0.5 overflow-y-auto rounded-(--r-composer) border border-input bg-(--code-surface) p-1" role="listbox" aria-label="slash commands">
			{#each slashMatches as m, i (m.cmd)}
				<li>
					<button
						type="button"
						class="flex w-full items-baseline gap-2.5 rounded-(--r-chip) px-2.5 py-[7px] text-left {i === slashSel ? 'bg-muted' : ''}"
						role="option"
						aria-selected={i === slashSel}
						onclick={() => acceptSlash(m.cmd)}
					>
						<span class="mono flex-none text-[13px] text-foreground">{m.cmd}</span>
						<span class="truncate text-xs text-muted-foreground">{m.desc}</span>
					</button>
				</li>
			{/each}
		</ul>
	{/if}
	<div class="flex flex-col rounded-(--r-composer) border border-input bg-(--code-surface) px-2.5 py-2">
		<textarea
			bind:this={ta}
			bind:value={$draft}
			oninput={autogrow}
			onpaste={onPaste}
			onkeydown={onKeydown}
			placeholder={agent ? 'Message the agent' : 'Type into terminal'}
			rows="1"
			class="max-h-24 w-full resize-none border-none bg-transparent text-sm leading-[1.45] text-foreground outline-none"
		></textarea>
		<div class="mt-1 flex items-center gap-2">
			{#if agent}
				<button
					class="inline-flex size-11 flex-none items-center justify-center rounded-full border border-border bg-card p-0 text-(--text-2) hover:bg-muted"
					type="button"
					aria-label="attach image"
					title="attach image"
					onclick={() => fileInput?.click()}
				>
					<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="3" y="3" width="18" height="18" rx="2" /><circle cx="8.5" cy="8.5" r="1.5" /><path d="m21 15-5-5L5 21" /></svg>
				</button>
				<input bind:this={fileInput} class="hidden" type="file" accept="image/*" onchange={onPickFile} />
			{/if}
			<span class="mono text-[11px] text-muted-foreground">{blocked ? 'agent blocked — reply or use the keys above' : $config.devCaptions ? (agent ? 'agent.prompt' : 'pane.send_text') : ''}</span>
			<button
				class="ml-auto size-11 rounded-full border-none text-lg {$draft.trim().length > 0 ? 'bg-primary text-primary-foreground' : 'bg-(--control) text-(--text-on-light)'}"
				aria-label="send"
				onclick={() => send()}
			>↑</button>
		</div>
	</div>
</div>
