<script lang="ts">
	import * as Drawer from '$lib/components/ui/drawer';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { sheet, closeSheet, type SheetAction } from '$lib/ui/state';

	let label = $state('');
	let cwd = $state('');
	$effect(() => {
		if ($sheet) { label = ''; cwd = ''; }
	});
	const actions = $derived($sheet && 'actions' in $sheet ? ($sheet.actions as SheetAction[]) : null);
	function confirm() {
		if ($sheet && !('actions' in $sheet)) $sheet.onConfirm(label, cwd);
		closeSheet();
	}
	function pick(a: SheetAction) {
		closeSheet();
		a.onSelect();
	}
</script>

<Drawer.Root open={$sheet !== null} onOpenChange={(o) => { if (!o) closeSheet(); }}>
	{#if $sheet}
		<Drawer.Content class="mx-auto max-w-[560px] border-hairline bg-raised px-4.5 pb-[calc(18px+env(safe-area-inset-bottom))]">
			<Drawer.Header class="p-0 text-left">
				<Drawer.Title class="mb-1.5 text-base font-semibold text-foreground">{$sheet.title}</Drawer.Title>
			</Drawer.Header>
			{#if actions}
				<ul class="my-1.5 mb-3.5 flex max-h-[50dvh] list-none flex-col gap-0.5 overflow-y-auto p-0" role="menu">
					{#each actions as a (a.label)}
						<li>
							<button
								class="flex min-h-12 w-full items-center gap-2.5 rounded-(--r-chip) px-3 text-left text-sm text-foreground hover:bg-muted data-[active=true]:bg-accent data-[danger=true]:text-(--blocked-badge-text)"
								data-active={a.active || undefined}
								data-danger={a.destructive || undefined}
								role="menuitem"
								onclick={() => pick(a)}
							>
								{#if a.glyph}<span class="mono w-5 flex-none text-center {a.destructive ? 'text-inherit' : 'text-(--text-3)'}" aria-hidden="true">{a.glyph}</span>{/if}
								<span class="min-w-0 flex-1 truncate">{a.label}</span>
								{#if a.hint}<span class="mono flex-none text-[11px] text-muted-foreground">{a.hint}</span>{/if}
							</button>
						</li>
					{/each}
				</ul>
				<div class="flex gap-2.5">
					<Button variant="outline" class="min-h-[46px] flex-1 rounded-(--r-btn) text-sm font-semibold" onclick={closeSheet}>Cancel</Button>
				</div>
			{:else if !('actions' in $sheet)}
				<p class="prose mb-3.5 text-[13px] text-(--text-3)">{$sheet.body}</p>
				{#if $sheet.hasInput}
					<label class="mb-3 block">
						<span class="mb-1.5 block text-[11px] text-(--text-3b)">{$sheet.inputLabel ?? 'Label'}</span>
						<Input class="mono bg-(--code-surface) text-sm" bind:value={label} placeholder="name" />
					</label>
				{/if}
				{#if $sheet.hasCwd}
					<label class="mb-3 block">
						<span class="mb-1.5 block text-[11px] text-(--text-3b)">Working directory</span>
						<Input class="mono bg-(--code-surface) text-sm" bind:value={cwd} placeholder="~/code/project" />
					</label>
				{/if}
				<div class="mono mb-4 rounded-(--r-chip) border border-hairline bg-(--code-surface) px-2.5 py-2 text-[11px] text-muted-foreground">{$sheet.call}</div>
				<div class="flex gap-2.5">
					<Button variant="outline" class="min-h-[46px] flex-1 rounded-(--r-btn) text-sm font-semibold" onclick={closeSheet}>Cancel</Button>
					<Button
						variant={$sheet.destructive ? 'destructive' : 'default'}
						class="min-h-[46px] flex-1 rounded-(--r-btn) text-sm font-semibold"
						onclick={confirm}
					>{$sheet.cta}</Button>
				</div>
			{/if}
		</Drawer.Content>
	{/if}
</Drawer.Root>
