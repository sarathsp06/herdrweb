<script lang="ts">
	import { Drawer as DrawerPrimitive } from "vaul-svelte";
	import { cn } from "$lib/utils.js";

	let {
		ref = $bindable(null),
		class: className,
		...restProps
	}: DrawerPrimitive.OverlayProps = $props();
</script>

<DrawerPrimitive.Overlay
	bind:ref
	data-slot="drawer-overlay"
	class={cn("bg-black/10 supports-backdrop-filter:backdrop-blur-xs fixed inset-0 z-50", className)}
	{...restProps}
/>
