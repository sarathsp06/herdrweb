<script lang="ts">
  import { onMount } from 'svelte';
  import { session } from '$lib/session/live';
  import { config, showToast } from '$lib/ui/state';
  import { enablePush, sendTestPush } from '$lib/push/register';
  import Toggle from '$lib/ui/Toggle.svelte';

  const s = session();
  type ThemeId = 'herdr-dark' | 'paper' | 'gruvbox' | 'solarized-light';
  const themes: { id: ThemeId; swatches: string[] }[] = [
    { id: 'herdr-dark', swatches: ['#0a0a0a', '#fafafa', '#f0a500'] },
    { id: 'paper', swatches: ['#ffffff', '#1a1a1b', '#0079d3'] },
    { id: 'gruvbox', swatches: ['#1d2021', '#ebdbb2', '#b8bb26'] },
    { id: 'solarized-light', swatches: ['#fdf6e3', '#586e75', '#b58900'] }
  ];
  const sizes: { label: string; scale: number }[] = [
    { label: 'S', scale: 0.9 },
    { label: 'M', scale: 1 },
    { label: 'L', scale: 1.15 },
    { label: 'XL', scale: 1.3 }
  ];
  const navCorners: { id: 'bottom-right' | 'bottom-left' | 'top'; label: string; glyph: string }[] = [
    { id: 'bottom-right', label: 'Bottom R', glyph: '◗' },
    { id: 'bottom-left', label: 'Bottom L', glyph: '◖' },
    { id: 'top', label: 'Top bar', glyph: '☰' }
  ];
  async function persist() {
    // In live mode, write the [web] table to config.toml (bridge reloads Herdr).
    try {
      await fetch('/api/config', { method: 'PUT', headers: { 'content-type': 'application/json' }, body: JSON.stringify($config) });
    } catch {
      // no bridge (fixtures/standalone); localStorage already holds the value
    }
  }
  async function reload() { await persist(); await s.request({ method: 'server.reload_config', params: {} }).catch(() => {}); showToast('config reloaded'); }
  function setConfig(patch: Partial<typeof $config>) { config.update((c) => ({ ...c, ...patch })); persist(); }
  function testToast(t: Awaited<ReturnType<typeof sendTestPush>>) {
    if (!t.ok) showToast('test failed — check bridge logs');
    else if (t.subs === 0) showToast('no device subscribed — enable push on this device');
    else if (t.sent > 0) showToast(`test sent to ${t.sent} device${t.sent > 1 ? 's' : ''}`);
    else showToast('push service rejected it — check bridge logs');
  }
  async function toggleNotify(v: boolean) {
    setConfig({ notify: v });
    if (!v) return;
    const r = await enablePush();
    if (!r.ok) {
      const why: Record<typeof r.reason, string> = {
        unsupported: 'push unsupported on this browser',
        insecure: 'push needs HTTPS (tailscale serve --https)',
        denied: 'notifications blocked in browser',
        nokey: 'bridge has no push key',
        error: 'push setup failed'
      };
      showToast(why[r.reason]);
      return;
    }
    // Enrolled — fire a test so success (or failure) is immediately visible.
    testToast(await sendTestPush());
  }
  async function testPush() {
    testToast(await sendTestPush());
  }
  function pickTheme(id: ThemeId) {
    setConfig({ theme: id });
  }

  let server = $state<Record<string, string>>({
    bridge: 'go · :7331',
    socket: '~/.config/herdr/herdr.sock',
    version: 'dev',
    protocol: 'ok'
  });

  onMount(() => {
    fetch('/api/health')
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (data) {
          server = {
            bridge: 'go · :7331',
            socket: data.socket || '~/.config/herdr/herdr.sock',
            version: data.version || 'dev',
            herdr: data.herdr || 'unknown',
            protocol: data.ok ? 'ok' : 'degraded'
          };
        }
      })
      .catch(() => {});
  });
</script>

<header class="hd"><h1 class="screen-title">Settings</h1></header>

<section>
  <div class="section-label">Theme</div>
  <div class="themes">
    {#each themes as t}
      <button class="theme" class:sel={$config.theme === t.id} onclick={() => pickTheme(t.id)}>
        <span class="swatches">{#each t.swatches as sw}<span style="background: {sw}"></span>{/each}</span>
        <span class="mono name">{t.id}</span>
      </button>
    {/each}
  </div>
  <div class="cap mono">writes [theme] in config.toml → server.reload_config</div>
</section>

<section>
  <div class="section-label">Text size</div>
  <div class="sizes">
    {#each sizes as sz}
      <button class="size" class:sel={($config.fontScale ?? 1) === sz.scale} onclick={() => setConfig({ fontScale: sz.scale })}>
        <span class="aa" style="font-size: {13 + (sz.scale - 1) * 14}px">Aa</span>
        <span class="mono szlabel">{sz.label}</span>
      </button>
    {/each}
  </div>
  <div class="cap mono">scales the whole UI · writes font_scale in config.toml</div>
</section>

<section>
  <div class="section-label">Phone nav button</div>
  <div class="sizes">
    {#each navCorners as nc}
      <button class="size" class:sel={($config.navCorner ?? 'bottom-right') === nc.id} onclick={() => setConfig({ navCorner: nc.id })}>
        <span class="aa">{nc.glyph}</span>
        <span class="mono szlabel">{nc.label}</span>
      </button>
    {/each}
  </div>
  <div class="cap mono">where the ☰ toggle sits on phones (one-handed reach)</div>
</section>

<section>
  <div class="section-label">Behavior</div>
  <div class="rows">
    <div class="row"><div><div class="n">Push when blocked</div><div class="d">Notify when an agent needs you.</div></div><Toggle checked={$config.notify} onchange={toggleNotify} /></div>
    {#if $config.notify}<div class="row"><div><div class="n">Send test notification</div><div class="d">Verify push reaches this device.</div></div><button class="test" onclick={testPush}>Send test</button></div>{/if}
    <div class="row"><div><div class="n">Follow focused pane</div><div class="d">Open the pane Herdr focuses.</div></div><Toggle checked={$config.follow} onchange={(v) => setConfig({ follow: v })} /></div>
    <div class="row"><div><div class="n">Keep ANSI colors in raw</div><div class="d">Render terminal colors in raw mode.</div></div><Toggle checked={$config.ansi} onchange={(v) => setConfig({ ansi: v })} /></div>
    <div class="row"><div><div class="n">Developer captions</div><div class="d">Show socket-call captions in the UI.</div></div><Toggle checked={$config.devCaptions} onchange={(v) => setConfig({ devCaptions: v })} /></div>
  </div>
</section>

<section>
  <div class="section-label">Server</div>
  <div class="server">
    {#each Object.entries(server) as [k, v]}
      <div class="kv"><span class="k mono">{k}</span><span class="v mono">{v}</span></div>
    {/each}
  </div>
  <div class="cap mono">svelte → go bridge → herdr socket</div>
  <button class="reload" onclick={reload}>Reload config</button>
</section>

<style>
  .hd { padding: 16px 14px 4px; }
  section { padding: 14px; }
  .section-label { margin-bottom: 10px; }
  .themes { display: flex; gap: 10px; }
  .theme { flex: 1; background: var(--card); border: 1px solid var(--hairline); border-radius: var(--r-card); padding: 12px; display: flex; flex-direction: column; gap: 8px; }
  .theme.sel { border-color: var(--control-selected-2); background: var(--surface-tint); }
  .swatches { display: flex; gap: 4px; }
  .swatches span { width: 14px; height: 14px; border-radius: 4px; border: 1px solid var(--hairline); }
  .name { font-size: 12px; color: var(--text-2); }
  .sizes { display: flex; gap: 10px; }
  .size { flex: 1; background: var(--card); border: 1px solid var(--hairline); border-radius: var(--r-card); padding: 12px; display: flex; flex-direction: column; align-items: center; gap: 6px; color: var(--text-2); }
  .size.sel { border-color: var(--control-selected-2); background: var(--surface-tint); }
  .aa { line-height: 1; font-weight: 600; }
  .szlabel { font-size: 11px; color: var(--text-4); }
  .cap { font-size: 10.5px; color: var(--text-4); margin-top: 8px; }
  .rows { display: flex; flex-direction: column; background: var(--card); border: 1px solid var(--hairline); border-radius: var(--r-card); }
  .row { display: flex; align-items: center; gap: 12px; padding: 12px 14px; border-top: 1px solid var(--hairline); }
  .row:first-child { border-top: none; }
  .row > div:first-child { flex: 1; }
  .n { font-size: 13.5px; font-weight: 500; }
  .d { font-size: 11.5px; color: var(--text-4); }
  .server { background: var(--card); border: 1px solid var(--hairline); border-radius: var(--r-card); padding: 4px 14px; }
  .kv { display: flex; justify-content: space-between; padding: 9px 0; border-top: 1px solid var(--hairline); font-size: 12px; }
  .kv:first-child { border-top: none; }
  .k { color: var(--text-4); } .v { color: var(--text-2); }
  .reload { margin-top: 12px; min-height: 46px; width: 100%; border-radius: var(--r-btn); border: 1px solid var(--control); background: none; color: var(--text-2); font-weight: 600; }
  .test { padding: 7px 14px; border-radius: var(--r-btn); border: 1px solid var(--control); background: none; color: var(--text-2); font-weight: 600; font-size: 12.5px; white-space: nowrap; }
</style>
