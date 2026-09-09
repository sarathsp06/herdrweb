<script lang="ts">
  import '../app.css';
  import { page } from '$app/stores';
  import { afterNavigate } from '$app/navigation';
  import { session } from '$lib/session/live';
  import { width, BREAKPOINT } from '$lib/layout/responsive';
  import { config } from '$lib/ui/state';
  import Sidebar from '$lib/screens/Sidebar.svelte';
  import BottomNav from '$lib/screens/BottomNav.svelte';
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
  // Full-screen pushes: the pane (terminal + composer) and diff own the whole
  // height — the tab bar yields so the keyboard row and composer stay reachable.
  const fullscreen = $derived(path.startsWith('/pane/'));

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

<div class="flex h-[calc(100dvh/var(--font-scale,1))] overflow-hidden">
  {#if desktop}
    <Sidebar spaces={$spaces} connection={$connection} />
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
</div>
<Toast />
<BottomSheet />
