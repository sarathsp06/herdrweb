<script lang="ts">
	import { session } from '$lib/session/live';
	import { findPaneIn } from '$lib/session/derive';
	import { drafts, setDraft, history, pushHistory, config, showToast } from '$lib/ui/state';
	import { pokeScrollback } from '$lib/session/scrollback';
	import type { Call, Prompt } from '$lib/protocol';
	import PromptAnswers from './PromptAnswers.svelte';
	import { filterSlash, loadSlashCommands, SLASH_COMMANDS, type SlashCommand } from './slash';
	import { uploadImage } from './upload';
	let {
		paneId,
		blocked,
		agent = true,
		label = 'agent',
		prompt,
		blockedSince,
		controlMode = false,
		ontogglecontrol
	}: {
		paneId: string;
		blocked: boolean;
		agent?: boolean;
		label?: string;
		prompt?: Prompt;
		blockedSince?: number;
		controlMode?: boolean;
		ontogglecontrol?: () => void;
	} = $props();
	const s = session();
	const spaces = s.spaces;
	let ta: HTMLTextAreaElement | undefined = $state();
	// Each pane keeps its own draft: switching panes swaps the text in the box,
	// so nothing typed for one pane can be sent to another.
	const draft = $derived($drafts[paneId] ?? '');
	// Phones have no Shift+Enter, so on touch an agent composer's Enter makes a
	// newline and the send button sends. Terminals send on Enter everywhere —
	// a multi-line terminal line is the rare case.
	const coarse = typeof matchMedia === 'function' && matchMedia('(pointer: coarse)').matches;
	const enterSends = $derived(!agent || !coarse);
	// Safari-only attribute, absent from Svelte's element types — spread it in.
	// Autocorrect mangles paths, flags and code in both panes.
	const NO_AUTOCORRECT = { autocorrect: 'off' };

	// ---- Prompt history ------------------------------------------------------
	// ↑ in an empty box (or the ↺ key, for phones) walks back through this
	// pane's sent prompts, ↓ walks forward and finally back to an empty box.
	// Recall only fills the draft; sending stays deliberate. histIdx counts back
	// from the newest entry; -1 means not browsing. Any real typing stops
	// browsing, so ↑/↓ go back to moving the caret.
	let histIdx = $state(-1);
	const past = $derived($history[paneId] ?? []);
	// While an agent is blocked the answer panel replaces the key row; "All
	// keys" brings the row back. writeIn is set after picking a write-in answer
	// so the box says what to type.
	let keysOpen = $state(false);
	let writeIn = $state(false);
	const answering = $derived(agent && blocked);
	$effect(() => {
		paneId;
		histIdx = -1;
		keysOpen = false;
		writeIn = false;
	});
	$effect(() => {
		if (!blocked) writeIn = false;
	});
	function startWriteIn() {
		writeIn = true;
		queueMicrotask(() => ta?.focus());
	}
	function recall(step: 1 | -1): boolean {
		const next = histIdx + step;
		if (next >= past.length) return false;
		histIdx = Math.max(next, -1);
		setDraft(paneId, histIdx === -1 ? '' : past[past.length - 1 - histIdx]);
		queueMicrotask(() => {
			ta?.focus();
			autogrow();
		});
		return true;
	}
	// The ↺ key: step back one prompt; past the oldest, wrap to an empty box.
	function recallOlder() {
		if (recall(1)) return;
		histIdx = -1;
		setDraft(paneId, '');
	}

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
	const slashMatches = $derived(agent ? filterSlash(draft, commands) : []);
	// While browsing history a recalled "/cmd" must not hijack ↑/↓.
	const slashOpen = $derived(slashMatches.length > 0 && !slashDismissed && histIdx < 0);
	// Reset the highlight to the top whenever the match set changes (on typing);
	// arrow navigation mutates slashSel without touching slashMatches.
	$effect(() => {
		slashMatches;
		slashSel = 0;
	});
	function acceptSlash(cmd: string) {
		setDraft(paneId, cmd + ' ');
		slashDismissed = false;
		queueMicrotask(() => {
			ta?.focus();
			autogrow();
		});
	}
	// The "/" key button: start (or return to) a bare-slash draft and focus the
	// box, so the full command list opens without typing on a phone keyboard.
	function openSlash() {
		setDraft(paneId, '/');
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
		const d = draft;
		setDraft(paneId, d + (d && !d.endsWith(' ') ? ' ' : '') + text + ' ');
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
		ta.style.height = Math.min(ta.scrollHeight, window.innerHeight * 0.4) + 'px';
	}
	// Re-fit when the draft changes underneath the box (pane switch, restore).
	$effect(() => {
		draft;
		queueMicrotask(autogrow);
	});
	function onInput(e: Event) {
		histIdx = -1;
		setDraft(paneId, (e.currentTarget as HTMLTextAreaElement).value);
		autogrow();
	}
	// A rejection this soon after sending means the input never reached the
	// pane (socket down, bad target) — put the text back. agent.prompt also
	// rejects much later (its wait timing out, the socket dropping mid-wait),
	// after the prompt was long delivered; those must not resurrect a draft.
	const RESTORE_WINDOW_MS = 3000;
	function restoreOnFailure(id: string, t: string, sentAt: number) {
		if (Date.now() - sentAt > RESTORE_WINDOW_MS) return;
		// Never clobber something typed since.
		if (!($drafts[id] ?? '')) setDraft(id, t);
		showToast('send failed — draft restored');
	}
	async function send() {
		const t = draft.trim();
		if (!t) return;
		const id = paneId;
		const sentAt = Date.now();
		// Clear the composer as soon as input is submitted — the cleared box and
		// the terminal echoing the input are the feedback; no toast. For agents,
		// agent.prompt's wait resolves only when the agent next goes idle/blocked
		// (long-lived), so the UI must not block on it. Terminals are not agents:
		// type the literal text then press Enter via the pane.* API (agent.* is a
		// silent no-op on a plain pane) — in order, so Enter can't overtake it.
		setDraft(id, '');
		pushHistory(id, t);
		histIdx = -1;
		// A blocked agent refuses agent.prompt (Herdr answers agent_blocked
		// without typing anything), so an answer typed while it waits goes in as
		// raw text plus Enter, like a terminal.
		if (agent && !blocked) {
			s.request({ method: 'agent.prompt', params: { target: id, text: t, wait: { until: ['idle', 'done', 'blocked'], timeout_ms: 900000 } } }).catch(
				() => restoreOnFailure(id, t, sentAt)
			);
			setTimeout(() => pokeScrollback(id), 150);
			return;
		}
		try {
			await s.request({ method: 'pane.send_text', params: { pane_id: id, text: t } });
			await s.request({ method: 'pane.send_keys', params: { pane_id: id, keys: ['enter'] } });
			pokeScrollback(id);
		} catch {
			restoreOnFailure(id, t, sentAt);
		}
	}
	// The keys you need when an agent goes wrong come first, so they stay on
	// screen at phone width; the row scrolls (with a fade) for the rest.
	const URGENT: { k: string; glyph: string; label: string }[] = [
		{ k: 'esc', glyph: 'esc', label: 'esc' },
		{ k: 'ctrl+c', glyph: '⌃C', label: 'ctrl+c' },
		{ k: 'enter', glyph: '⏎', label: 'enter' }
	];
	const NAV: { k: string; glyph: string; label: string }[] = [
		{ k: 'up', glyph: '↑', label: 'up' },
		{ k: 'down', glyph: '↓', label: 'down' },
		{ k: 'left', glyph: '←', label: 'left' },
		{ k: 'right', glyph: '→', label: 'right' },
		{ k: 'tab', glyph: '⇥', label: 'tab' },
		{ k: 'shift+tab', glyph: '⇧⇥', label: 'shift+tab' },
		{ k: 'ctrl+d', glyph: '⌃D', label: 'ctrl+d' }
	];
	function sendKey(k: string) {
		const call: Call = agent
			? { method: 'agent.send_keys', params: { target: paneId, keys: [k] } }
			: { method: 'pane.send_keys', params: { pane_id: paneId, keys: [k] } };
		s.request(call).then(
			() => pokeScrollback(paneId),
			() => showToast(`${k} not sent`)
		);
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
		const plain = !e.shiftKey && !e.altKey && !e.metaKey && !e.ctrlKey;
		if (plain && e.key === 'ArrowUp' && (histIdx >= 0 || draft === '')) {
			if (recall(1)) e.preventDefault();
			return;
		}
		if (plain && e.key === 'ArrowDown' && histIdx >= 0) {
			e.preventDefault();
			recall(-1);
			return;
		}
		if (enterSends && e.key === 'Enter' && !e.shiftKey && !e.isComposing) {
			e.preventDefault();
			void send();
		}
	}
</script>

<div class="border-t border-hairline bg-background px-3.5 pt-2 pb-[calc(10px+env(safe-area-inset-bottom))]">
	{#if answering}
		<PromptAnswers
			{paneId}
			{label}
			{prompt}
			{blockedSince}
			{keysOpen}
			ontogglekeys={() => (keysOpen = !keysOpen)}
			onwritein={startWriteIn}
		/>
	{/if}
	{#if !answering || keysOpen}
		<div class="keyrow flex gap-1.5 overflow-x-auto pb-2" role="group" aria-label="keys">
			{#each URGENT as n (n.k)}
				<button
					class="mono h-11 min-w-11 flex-none rounded-(--r-chip) border border-(--control-selected-2) bg-card px-2.5 text-sm text-foreground hover:bg-muted"
					aria-label={n.label}
					title={n.label}
					onclick={() => sendKey(n.k)}
				>{n.glyph}</button>
			{/each}
			<span class="my-2 w-px flex-none bg-border" aria-hidden="true"></span>
			{#each NAV as n (n.k)}
				<button
					class="mono h-11 min-w-11 flex-none rounded-(--r-chip) border border-border bg-card px-2.5 text-sm text-(--text-2) hover:bg-muted"
					aria-label={n.label}
					title={n.label}
					onclick={() => sendKey(n.k)}
				>{n.glyph}</button>
			{/each}
			{#if agent}
				<button
					class="mono h-11 min-w-11 flex-none rounded-(--r-chip) border border-border bg-card px-2.5 text-sm text-(--text-2) hover:bg-muted"
					aria-label="slash commands"
					title="slash commands"
					onclick={openSlash}
				>/</button>
			{/if}
			{#if past.length > 0}
				<button
					class="mono h-11 min-w-11 flex-none rounded-(--r-chip) border border-border bg-card px-2.5 text-sm text-(--text-2) hover:bg-muted"
					aria-label="previous prompt"
					title="previous prompt (↑)"
					onclick={recallOlder}
				>↺</button>
			{/if}
			{#if agent && ontogglecontrol}
				<button
					class="h-11 flex-none rounded-(--r-chip) border px-3 text-[13px] font-semibold {controlMode
						? 'border-(--worktree-text) bg-(--worktree-bg) text-foreground'
						: 'border-border bg-card text-(--text-2) hover:bg-muted'}"
					aria-label="direct control: swipe keys"
					aria-pressed={controlMode}
					title="swipe the terminal to send arrow keys"
					onclick={ontogglecontrol}
				>Swipe keys</button>
			{/if}
		</div>
	{/if}
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
			value={draft}
			oninput={onInput}
			onpaste={onPaste}
			onkeydown={onKeydown}
			placeholder={!agent ? 'Type into terminal' : writeIn ? 'Type your answer' : answering ? 'Or type a different answer' : 'Message the agent'}
			rows="1"
			autocomplete="off"
			{...NO_AUTOCORRECT}
			autocapitalize={agent ? 'sentences' : 'off'}
			spellcheck={agent}
			enterkeyhint={enterSends ? 'send' : 'enter'}
			class="max-h-[40dvh] w-full resize-none border-none bg-transparent text-sm leading-[1.45] text-foreground outline-none"
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
			<span class="mono text-[11px] text-muted-foreground">{$config.devCaptions ? (agent && !blocked ? 'agent.prompt' : 'pane.send_text') : ''}</span>
			<button
				class="ml-auto size-11 rounded-full border-none text-lg {draft.trim().length > 0 ? 'bg-primary text-primary-foreground' : 'bg-(--control) text-(--text-on-light)'}"
				aria-label="send"
				onclick={() => send()}
			>↑</button>
		</div>
	</div>
</div>

<style>
	/* Show that the key row scrolls: fade its right edge. */
	.keyrow {
		-webkit-mask-image: linear-gradient(90deg, #000 85%, transparent);
		mask-image: linear-gradient(90deg, #000 85%, transparent);
	}
</style>
