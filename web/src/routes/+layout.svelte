<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { session } from '$lib/session/live';
  import { width, BREAKPOINT } from '$lib/layout/responsive';
  import { config, navOpen } from '$lib/ui/state';
  import Sidebar from '$lib/screens/Sidebar.svelte';
  import Breadcrumbs from '$lib/screens/Breadcrumbs.svelte';
  import Toast from '$lib/ui/Toast.svelte';
  import BottomSheet from '$lib/ui/BottomSheet.svelte';

  let { children } = $props();
  const s = session();
  const spaces = s.spaces;
  const connection = s.connection;

  const desktop = $derived($width >= BREAKPOINT);
  const path = $derived($page.url.pathname);
  // Full-screen pushes: chat and diff hide the tab bar and use the full width.
  const fullscreen = $derived(path.startsWith('/pane/'));

  // Initialise nav visibility per breakpoint once: open on desktop, closed
  // (drawer) on phones.
  let inited = false;
  $effect(() => {
    if (inited) return;
    navOpen.set(desktop);
    inited = true;
  });

  const closeOnMobile = () => { if (!desktop) navOpen.set(false); };

  onMount(() => {
    document.documentElement.dataset.theme = $config.theme;
  });
  $effect(() => {
    document.documentElement.dataset.theme = $config.theme;
  });
  // Webapp-level text size: scale the whole UI via document zoom.
  $effect(() => {
    document.documentElement.style.setProperty('zoom', String($config.fontScale ?? 1));
  });
</script>

<div class="shell" class:desktop>
  {#if desktop}
    {#if $navOpen}
      <Sidebar spaces={$spaces} connection={$connection} onselect={closeOnMobile} onclose={() => navOpen.set(false)} />
    {/if}
  {:else if $navOpen}
    <button class="backdrop" aria-label="close navigation" onclick={() => navOpen.set(false)}></button>
    <div class="drawer">
      <Sidebar spaces={$spaces} connection={$connection} embedded onselect={closeOnMobile} onclose={() => navOpen.set(false)} />
    </div>
  {/if}

  <div class="mainwrap">
    <Breadcrumbs />
    <main class="content" class:desktop class:full={fullscreen}>
      {@render children()}
    </main>
  </div>
</div>
<Toast />
<BottomSheet />

<style>
  .shell { height: 100vh; display: flex; overflow: hidden; }
  .mainwrap { flex: 1; min-width: 0; display: flex; flex-direction: column; height: 100vh; }
  .content { flex: 1; min-width: 0; min-height: 0; overflow-y: auto; }
  .shell.desktop .content { padding: 0 max(28px, calc(50% - 560px)); }
  .shell.desktop .content.full { padding: 0; }

  .backdrop { position: fixed; inset: 0; z-index: 45; border: none; background: rgba(0, 0, 0, 0.5); }
  .drawer { position: fixed; top: 0; left: 0; bottom: 0; z-index: 46; width: min(88vw, 340px); background: var(--sidebar-bg); border-right: 1px solid var(--hairline); overflow-y: auto; box-shadow: 0 0 40px rgba(0, 0, 0, 0.45); }
</style>
