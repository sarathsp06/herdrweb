<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { session } from '$lib/session/live';
  import { width, BREAKPOINT } from '$lib/layout/responsive';
  import { config } from '$lib/ui/state';
  import Sidebar from '$lib/screens/Sidebar.svelte';
  import TabBar from '$lib/screens/TabBar.svelte';
  import Toast from '$lib/ui/Toast.svelte';
  import BottomSheet from '$lib/ui/BottomSheet.svelte';

  let { children } = $props();
  const s = session();
  const spaces = s.spaces;
  const connection = s.connection;

  const desktop = $derived($width >= BREAKPOINT);
  const path = $derived($page.url.pathname);
  // Full-screen pushes: chat and diff hide the tab bar.
  const fullscreen = $derived(path.startsWith('/pane/'));

  onMount(() => {
    document.documentElement.dataset.theme = $config.theme;
  });
  $effect(() => {
    document.documentElement.dataset.theme = $config.theme;
  });
</script>

<div class="shell" class:desktop>
  {#if desktop}
    <Sidebar spaces={$spaces} connection={$connection} />
  {/if}
  <main class="content" class:desktop>
    {@render children()}
  </main>
  {#if !desktop && !fullscreen}
    <TabBar {path} />
  {/if}
</div>
<Toast />
<BottomSheet />

<style>
  .shell { min-height: 100vh; display: flex; }
  .content { flex: 1; min-width: 0; }
  .shell.desktop .content { padding: 0 max(28px, calc(50% - 430px)); }
</style>
