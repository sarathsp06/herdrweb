<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { afterNavigate } from '$app/navigation';
  import { session } from '$lib/session/live';
  import { width, BREAKPOINT } from '$lib/layout/responsive';
  import { config, navOpen } from '$lib/ui/state';
  import Sidebar from '$lib/screens/Sidebar.svelte';
  import Breadcrumbs from '$lib/screens/Breadcrumbs.svelte';
  import Toast from '$lib/ui/Toast.svelte';
  import BottomSheet from '$lib/ui/BottomSheet.svelte';

  let { children } = $props();
  let contentEl: HTMLElement | undefined = $state();

  // The outer window/body scroll is locked (dvh app shell), so SvelteKit's
  // default scroll-reset-on-navigate never runs — `.content` below is the real
  // scroll container for every route without its own inner scroller. Reset it
  // ourselves on every completed client-side navigation regardless of source
  // (link, programmatic goto, back/forward); a route with its own inner
  // scrollback (e.g. the pane view) owns its own pinning and is unaffected.
  afterNavigate(() => {
    if (contentEl) contentEl.scrollTop = 0;
  });
  const s = session();
  const spaces = s.spaces;
  const connection = s.connection;

  const desktop = $derived($width >= BREAKPOINT);
  const path = $derived($page.url.pathname);
  // Full-screen pushes: chat and diff hide the tab bar and use the full width.
  const fullscreen = $derived(path.startsWith('/pane/'));
  const navCorner = $derived($config.navCorner ?? 'bottom-right');
  const fab = $derived(!desktop && navCorner !== 'top');
  // Raise the FAB above the pane composer (chat route only; diff has a short footer).
  const fabRaised = $derived(fullscreen && !path.endsWith('/diff'));

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
  // Webapp-level text size: scale the whole UI via document zoom. `zoom` also
  // scales viewport-unit heights (100dvh/100vh), so full-height shells must
  // divide by --font-scale to stay pinned to the real viewport (else the pane
  // composer is pushed below the clipped bottom edge).
  $effect(() => {
    const scale = $config.fontScale ?? 1;
    document.documentElement.style.setProperty('zoom', String(scale));
    document.documentElement.style.setProperty('--font-scale', String(scale));
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
    <Breadcrumbs showNav={!fab} />
    <main class="content" class:desktop class:full={fullscreen} bind:this={contentEl}>
      {@render children()}
    </main>
  </div>
</div>
{#if fab && !$navOpen}
  <button
    class="navfab {navCorner}"
    class:raised={fabRaised}
    aria-label="toggle navigation"
    title="navigation"
    onclick={() => navOpen.set(true)}
  >☰</button>
{/if}
<Toast />
<BottomSheet />

<style>
  .shell { height: calc(100dvh / var(--font-scale, 1)); display: flex; overflow: hidden; }
  .mainwrap { flex: 1; min-width: 0; display: flex; flex-direction: column; height: 100%; min-height: 0; }
  .content { flex: 1; min-width: 0; min-height: 0; overflow-y: auto; }
  .shell.desktop .content { padding: 0 max(28px, calc(50% - 560px)); }
  .shell.desktop .content.full { padding: 0; }

  .backdrop { position: fixed; inset: 0; z-index: 45; border: none; background: rgba(0, 0, 0, 0.5); }
  .drawer { position: fixed; top: 0; left: 0; bottom: 0; z-index: 46; width: min(88vw, 340px); background: var(--sidebar-bg); border-right: 1px solid var(--hairline); overflow-y: auto; box-shadow: 0 0 40px rgba(0, 0, 0, 0.45); }
  .navfab { position: fixed; z-index: 50; width: 48px; height: 48px; border-radius: 50%; border: 1px solid var(--control); background: var(--card); color: var(--text-1); font-size: 19px; line-height: 1; box-shadow: 0 6px 20px rgba(0, 0, 0, 0.4); bottom: calc(16px + env(safe-area-inset-bottom)); }
  .navfab:active { background: var(--surface-tint); }
  .navfab.bottom-right { right: 16px; }
  .navfab.bottom-left { left: 16px; }
  /* Clear the pane composer so the FAB doesn't cover the send button. */
  .navfab.raised { bottom: calc(132px + env(safe-area-inset-bottom)); }
</style>
