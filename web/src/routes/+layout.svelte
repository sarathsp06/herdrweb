<script lang="ts">
  import '../app.css';
  import { page } from '$app/stores';
  import { afterNavigate } from '$app/navigation';
  import { session } from '$lib/session/live';
  import { width, BREAKPOINT } from '$lib/layout/responsive';
  import { config, sidebarCollapsed, sidebarPeek } from '$lib/ui/state';
  import Sidebar from '$lib/screens/Sidebar.svelte';
  import BottomNav from '$lib/screens/BottomNav.svelte';
  import Toast from '$lib/ui/Toast.svelte';
  import BottomSheet from '$lib/ui/BottomSheet.svelte';

  let { children } = $props();
  let contentEl: HTMLElement | undefined = $state();

  // The outer window/body scroll is locked (full-screen app shell), so SvelteKit's
  // default scroll-reset-on-navigate never runs — `.content` below is the real
  // scroll container for every route without its own inner scroller. Reset it
  // ourselves on every completed client-side navigation regardless of source
  // (link, programmatic goto, back/forward); a route with its own inner
  // scrollback (e.g. the pane view) owns its own pinning and is unaffected.
  afterNavigate(() => {
    if (contentEl) contentEl.scrollTop = 0;
    sidebarPeek.set(false);
  });
  const s = session();
  const spaces = s.spaces;
  const connection = s.connection;

  const desktop = $derived($width >= BREAKPOINT);
  const path = $derived($page.url.pathname);
  // Full-screen pushes: the pane (terminal + composer) and diff own the whole
  // height — the tab bar yields so the keyboard row and composer stay reachable.
  const fullscreen = $derived(path.startsWith('/pane/'));
  // Collapse only ever applies to the pane: other routes are narrow menus that
  // gain nothing from the width, and would be stranded without navigation.
  const collapsed = $derived(desktop && fullscreen && $sidebarCollapsed);

  // Keep the OS/browser chrome colour in sync with the active theme - must
  // match each theme's `--app-bg` in lib/tokens.css and app.html's pre-paint
  // copy of this map. `$effect` already runs once on mount, so no separate
  // onMount is needed.
  const THEME_COLOR: Record<string, string> = {
    'herdr-dark': '#0a0a0a',
    gruvbox: '#1d2021',
    'solarized-light': '#fdf6e3',
    paper: '#ffffff'
  };
  $effect(() => {
    const isLight = $config.theme === 'solarized-light' || $config.theme === 'paper';
    document.documentElement.dataset.theme = $config.theme;
    document.querySelector('meta[name="theme-color"]')?.setAttribute('content', THEME_COLOR[$config.theme] ?? '#0a0a0a');
    document
      .querySelector('meta[name="apple-mobile-web-app-status-bar-style"]')
      ?.setAttribute('content', isLight ? 'default' : 'black-translucent');
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

<!-- --screen-h: see app.css. pt keeps headers below a translucent iOS status bar. -->
<div
  class="relative flex h-[calc(var(--screen-h)/var(--font-scale,1))] overflow-hidden pt-[calc(env(safe-area-inset-top)/var(--font-scale,1))]"
>
  {#if desktop && !collapsed}
    <Sidebar spaces={$spaces} connection={$connection} control={fullscreen ? 'hide' : undefined} />
  {/if}

  <div class="flex h-full min-h-0 min-w-0 flex-1 flex-col">
    <main
      class="content min-h-0 min-w-0 flex-1 overflow-y-auto {desktop && !fullscreen
        ? 'px-[max(28px,calc(50%-560px))]'
        : ''}"
      bind:this={contentEl}
    >
      {@render children()}
    </main>
    {#if !desktop && !fullscreen}
      <BottomNav />
    {/if}
  </div>

  <!-- Absolute inside the shell, not fixed: an installed iOS app lays fixed
       elements out against a viewport shorter than the screen (see app.css). -->
  {#if collapsed && $sidebarPeek}
    <button class="absolute inset-0 z-40 bg-black/50" aria-label="close sidebar" onclick={() => sidebarPeek.set(false)}
    ></button>
    <div class="absolute bottom-0 left-0 z-50 top-[calc(env(safe-area-inset-top)/var(--font-scale,1))] shadow-2xl">
      <Sidebar spaces={$spaces} connection={$connection} control="pin" />
    </div>
  {/if}
</div>
<Toast />
<BottomSheet />
