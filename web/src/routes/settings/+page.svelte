<script lang="ts">
  import { onMount } from 'svelte';
  import { session } from '$lib/session/live';
  import { config, showToast } from '$lib/ui/state';
  import { enablePush, sendTestPush } from '$lib/push/register';
  import Toggle from '$lib/ui/Toggle.svelte';
  import { Button } from '$lib/components/ui/button';

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
  // Enrol this browser and surface why it failed. Idempotent (reuses an existing
  // subscription), so both the toggle and the test button can call it.
  async function ensureEnrolled(): Promise<boolean> {
    const r = await enablePush();
    if (r.ok) return true;
    const why: Record<typeof r.reason, string> = {
      unsupported: 'push unsupported on this browser',
      insecure: 'push needs HTTPS (tailscale serve --https)',
      denied: 'notifications blocked in browser',
      nokey: 'bridge has no push key',
      error: 'push setup failed'
    };
    showToast(why[r.reason]);
    return false;
  }
  async function toggleNotify(v: boolean) {
    setConfig({ notify: v });
    if (!v) return;
    // Enrol, then fire a test so success (or failure) is immediately visible.
    if (await ensureEnrolled()) testToast(await sendTestPush());
  }
  async function testPush() {
    // The button is shown whenever `notify` is on — including the default-on
    // first load where this device was never actually subscribed. Enrol first
    // (this click is the required user gesture) so the test has a device to reach.
    if (await ensureEnrolled()) testToast(await sendTestPush());
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

<header class="px-3.5 pt-4 pb-1"><h1 class="screen-title">Settings</h1></header>

<section class="p-3.5">
  <div class="section-label mb-2.5">Theme</div>
  <div class="themes flex gap-2.5">
    {#each themes as t}
      <button
        class="flex flex-1 flex-col gap-2 rounded-(--r-card) border p-3 {$config.theme === t.id
          ? 'border-ring bg-muted'
          : 'border-(--hairline) bg-card'}"
        onclick={() => pickTheme(t.id)}
      >
        <span class="flex gap-1"
          >{#each t.swatches as sw}<span
              class="h-3.5 w-3.5 rounded border border-(--hairline)"
              style="background: {sw}"
            ></span>{/each}</span
        >
        <span class="mono text-xs text-(--text-2)">{t.id}</span>
      </button>
    {/each}
  </div>
  <div class="mono mt-2 text-[10.5px] text-muted-foreground">writes [theme] in config.toml → server.reload_config</div>
</section>

<section class="p-3.5">
  <div class="section-label mb-2.5">Text size</div>
  <div class="flex gap-2.5">
    {#each sizes as sz}
      <button
        class="flex flex-1 flex-col items-center gap-1.5 rounded-(--r-card) border p-3 text-(--text-2) {($config.fontScale ?? 1) === sz.scale
          ? 'border-ring bg-muted'
          : 'border-(--hairline) bg-card'}"
        onclick={() => setConfig({ fontScale: sz.scale })}
      >
        <span class="leading-none font-semibold" style="font-size: {13 + (sz.scale - 1) * 14}px">Aa</span>
        <span class="mono text-[11px] text-muted-foreground">{sz.label}</span>
      </button>
    {/each}
  </div>
  <div class="mono mt-2 text-[10.5px] text-muted-foreground">scales the whole UI · writes font_scale in config.toml</div>
</section>


<section class="p-3.5">
  <div class="section-label mb-2.5">Behavior</div>
  <div class="flex flex-col rounded-(--r-card) border border-(--hairline) bg-card">
    <div class="row flex items-center gap-3 px-3.5 py-3">
      <div class="flex-1">
        <div class="text-[13.5px] font-medium">Push when blocked</div>
        <div class="text-[11.5px] text-muted-foreground">Notify when an agent needs you.</div>
      </div>
      <Toggle checked={$config.notify} onchange={toggleNotify} />
    </div>
    {#if $config.notify}
      <div class="row flex items-center gap-3 border-t border-(--hairline) px-3.5 py-3">
        <div class="flex-1">
          <div class="text-[13.5px] font-medium">Send test notification</div>
          <div class="text-[11.5px] text-muted-foreground">Verify push reaches this device.</div>
        </div>
        <Button variant="outline" class="rounded-lg text-[12.5px] font-semibold whitespace-nowrap text-(--text-2)" onclick={testPush}
          >Send test</Button
        >
      </div>
    {/if}
    <div class="row flex items-center gap-3 border-t border-(--hairline) px-3.5 py-3">
      <div class="flex-1">
        <div class="text-[13.5px] font-medium">Follow focused pane</div>
        <div class="text-[11.5px] text-muted-foreground">Open the pane Herdr focuses.</div>
      </div>
      <Toggle checked={$config.follow} onchange={(v) => setConfig({ follow: v })} />
    </div>
    <div class="row flex items-center gap-3 border-t border-(--hairline) px-3.5 py-3">
      <div class="flex-1">
        <div class="text-[13.5px] font-medium">Keep ANSI colors in raw</div>
        <div class="text-[11.5px] text-muted-foreground">Render terminal colors in raw mode.</div>
      </div>
      <Toggle checked={$config.ansi} onchange={(v) => setConfig({ ansi: v })} />
    </div>
    <div class="row flex items-center gap-3 border-t border-(--hairline) px-3.5 py-3">
      <div class="flex-1">
        <div class="text-[13.5px] font-medium">Developer captions</div>
        <div class="text-[11.5px] text-muted-foreground">Show socket-call captions in the UI.</div>
      </div>
      <Toggle checked={$config.devCaptions} onchange={(v) => setConfig({ devCaptions: v })} />
    </div>
  </div>
</section>

<section class="p-3.5">
  <div class="section-label mb-2.5">Server</div>
  <div class="rounded-(--r-card) border border-(--hairline) bg-card px-3.5 py-1">
    {#each Object.entries(server) as [k, v], i}
      <div class="flex justify-between py-[9px] text-xs {i > 0 ? 'border-t border-(--hairline)' : ''}">
        <span class="mono text-muted-foreground">{k}</span><span class="mono text-(--text-2)">{v}</span>
      </div>
    {/each}
  </div>
  <div class="mono mt-2 text-[10.5px] text-muted-foreground">svelte → go bridge → herdr socket</div>
  <Button variant="outline" class="mt-3 min-h-[46px] w-full rounded-lg font-semibold text-(--text-2)" onclick={reload}
    >Reload config</Button
  >
</section>
