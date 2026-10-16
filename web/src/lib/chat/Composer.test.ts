import '@testing-library/jest-dom/vitest';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';

const request = vi.fn(async (_call?: { method: string }) => ({ ok: true }));
vi.mock('$lib/session/live', async () => {
  // Dynamic import required: vi.mock factories are hoisted above top-level
  // imports, so a static `readable` binding is not initialized yet here.
  const { readable } = await import('svelte/store');
  return { session: () => ({ request, spaces: readable([]) }) };
});

import Composer from './Composer.svelte';
import { drafts, history } from '$lib/ui/state';
import { get } from 'svelte/store';

beforeEach(() => {
  request.mockClear();
  request.mockImplementation(async () => ({ ok: true }));
  drafts.set({});
  history.set({});
});
afterEach(() => vi.unstubAllGlobals());

describe('Composer', () => {
  it('agent nav keys send agent.send_keys as single-element arrays', async () => {
    render(Composer, { paneId: 'w1:p1', blocked: false, agent: true });
    await fireEvent.click(screen.getByLabelText('up'));
    expect(request).toHaveBeenCalledWith({ method: 'agent.send_keys', params: { target: 'w1:p1', keys: ['up'] } });
    await fireEvent.click(screen.getByLabelText('enter'));
    expect(request).toHaveBeenCalledWith({ method: 'agent.send_keys', params: { target: 'w1:p1', keys: ['enter'] } });
  });

  it('agent send button submits typed text via agent.prompt', async () => {
    render(Composer, { paneId: 'w1:p1', blocked: false, agent: true });
    const ta = screen.getByPlaceholderText('Message the agent');
    await fireEvent.input(ta, { target: { value: 'do the thing' } });
    await fireEvent.click(screen.getByLabelText('send'));
    expect(request).toHaveBeenCalledWith(
      expect.objectContaining({ method: 'agent.prompt', params: expect.objectContaining({ target: 'w1:p1', text: 'do the thing' }) })
    );
  });

  it('terminal send types literal text then Enter via pane.* (not agent.prompt)', async () => {
    render(Composer, { paneId: 'w1:p9', blocked: false, agent: false });
    const ta = screen.getByPlaceholderText('Type into terminal');
    await fireEvent.input(ta, { target: { value: 'ls -la' } });
    await fireEvent.click(screen.getByLabelText('send'));
    expect(request).toHaveBeenCalledWith({ method: 'pane.send_text', params: { pane_id: 'w1:p9', text: 'ls -la' } });
    expect(request).toHaveBeenCalledWith({ method: 'pane.send_keys', params: { pane_id: 'w1:p9', keys: ['enter'] } });
    expect(request).not.toHaveBeenCalledWith(expect.objectContaining({ method: 'agent.prompt' }));
  });

  it('terminal nav keys send pane.send_keys (not agent.send_keys)', async () => {
    render(Composer, { paneId: 'w1:p9', blocked: false, agent: false });
    await fireEvent.click(screen.getByLabelText('ctrl+c'));
    expect(request).toHaveBeenCalledWith({ method: 'pane.send_keys', params: { pane_id: 'w1:p9', keys: ['ctrl+c'] } });
  });

  it('tab nav key sends agent.send_keys tab', async () => {
    render(Composer, { paneId: 'w1:p1', blocked: false, agent: true });
    await fireEvent.click(screen.getByLabelText('tab'));
    expect(request).toHaveBeenCalledWith({ method: 'agent.send_keys', params: { target: 'w1:p1', keys: ['tab'] } });
  });

  it('typing a slash query surfaces matching commands', async () => {
    render(Composer, { paneId: 'w1:p1', blocked: false, agent: true });
    const ta = screen.getByPlaceholderText('Message the agent');
    await fireEvent.input(ta, { target: { value: '/cl' } });
    expect(screen.getByRole('option', { name: /\/clear/ })).toBeTruthy();
  });

  it('Tab accepts the selected slash command into the draft without sending', async () => {
    render(Composer, { paneId: 'w1:p1', blocked: false, agent: true });
    const ta = screen.getByPlaceholderText('Message the agent') as HTMLTextAreaElement;
    await fireEvent.input(ta, { target: { value: '/cl' } });
    await fireEvent.keyDown(ta, { key: 'Tab' });
    expect(ta.value).toBe('/clear ');
    expect(request).not.toHaveBeenCalled();
  });

  it('Enter accepts an open palette instead of sending', async () => {
    render(Composer, { paneId: 'w1:p1', blocked: false, agent: true });
    const ta = screen.getByPlaceholderText('Message the agent') as HTMLTextAreaElement;
    await fireEvent.input(ta, { target: { value: '/comp' } });
    await fireEvent.keyDown(ta, { key: 'Enter' });
    expect(ta.value).toBe('/compact ');
    expect(request).not.toHaveBeenCalled();
  });

  it('Enter sends once the palette is closed', async () => {
    render(Composer, { paneId: 'w1:p1', blocked: false, agent: true });
    const ta = screen.getByPlaceholderText('Message the agent');
    await fireEvent.input(ta, { target: { value: '/clear ' } });
    await fireEvent.keyDown(ta, { key: 'Enter' });
    expect(request).toHaveBeenCalledWith(
      expect.objectContaining({ method: 'agent.prompt', params: expect.objectContaining({ text: '/clear' }) })
    );
  });

  it('terminal panes get no slash palette', async () => {
    render(Composer, { paneId: 'w1:p9', blocked: false, agent: false });
    const ta = screen.getByPlaceholderText('Type into terminal');
    await fireEvent.input(ta, { target: { value: '/clear' } });
    expect(screen.queryByRole('option')).toBeNull();
  });

  it('attaching an image uploads it and inserts the path into the draft', async () => {
    const fetchMock = vi.fn(async () => ({ ok: true, json: async () => ({ path: '/tmp/herdr-uploads/x.png' }) }));
    vi.stubGlobal('fetch', fetchMock);
    render(Composer, { paneId: 'w1:p1', blocked: false, agent: true });
    const input = document.querySelector('input[type=file]') as HTMLInputElement;
    const file = new File([new Uint8Array([1, 2, 3])], 'shot.png', { type: 'image/png' });
    await fireEvent.change(input, { target: { files: [file] } });
    await waitFor(() => expect(get(drafts)['w1:p1']).toContain('/tmp/herdr-uploads/x.png'));
    expect(fetchMock).toHaveBeenCalledWith('/api/upload', expect.objectContaining({ method: 'POST' }));
  });

  it('terminal panes have no attach button', () => {
    render(Composer, { paneId: 'w1:p9', blocked: false, agent: false });
    expect(screen.queryByLabelText('attach image')).toBeNull();
  });

  it('keeps a separate draft per pane', async () => {
    const { rerender } = render(Composer, { paneId: 'w1:p1', blocked: false, agent: true });
    const ta = screen.getByPlaceholderText('Message the agent') as HTMLTextAreaElement;
    await fireEvent.input(ta, { target: { value: 'for pane one' } });
    await rerender({ paneId: 'w1:p2', blocked: false, agent: true });
    expect(ta.value).toBe('');
    await fireEvent.click(screen.getByLabelText('send'));
    expect(request).not.toHaveBeenCalledWith(expect.objectContaining({ method: 'agent.prompt' }));
    await rerender({ paneId: 'w1:p1', blocked: false, agent: true });
    expect(ta.value).toBe('for pane one');
  });

  it('terminal send presses Enter only after the text call resolves', async () => {
    let releaseText!: () => void;
    request.mockImplementation(async (call) => {
      if (call?.method === 'pane.send_text') await new Promise<void>((r) => (releaseText = r));
      return { ok: true };
    });
    render(Composer, { paneId: 'w1:p9', blocked: false, agent: false });
    const ta = screen.getByPlaceholderText('Type into terminal');
    await fireEvent.input(ta, { target: { value: 'ls' } });
    await fireEvent.click(screen.getByLabelText('send'));
    expect(request).not.toHaveBeenCalledWith(expect.objectContaining({ method: 'pane.send_keys' }));
    releaseText();
    await waitFor(() =>
      expect(request).toHaveBeenCalledWith({ method: 'pane.send_keys', params: { pane_id: 'w1:p9', keys: ['enter'] } })
    );
  });

  it('restores the draft when a send fails straight away', async () => {
    request.mockImplementation(async () => {
      throw new Error('socket not open');
    });
    render(Composer, { paneId: 'w1:p9', blocked: false, agent: false });
    const ta = screen.getByPlaceholderText('Type into terminal') as HTMLTextAreaElement;
    await fireEvent.input(ta, { target: { value: 'make test' } });
    await fireEvent.click(screen.getByLabelText('send'));
    await waitFor(() => expect(ta.value).toBe('make test'));
    expect(request).not.toHaveBeenCalledWith(expect.objectContaining({ method: 'pane.send_keys' }));
  });

  it('agent prompt rejection restores the draft too', async () => {
    request.mockImplementation(async () => {
      throw new Error('socket not open');
    });
    render(Composer, { paneId: 'w1:p1', blocked: false, agent: true });
    const ta = screen.getByPlaceholderText('Message the agent') as HTMLTextAreaElement;
    await fireEvent.input(ta, { target: { value: 'fix it' } });
    await fireEvent.click(screen.getByLabelText('send'));
    await waitFor(() => expect(ta.value).toBe('fix it'));
  });

  it('terminal box turns off autocapitalize/autocorrect/spellcheck', () => {
    render(Composer, { paneId: 'w1:p9', blocked: false, agent: false });
    const ta = screen.getByPlaceholderText('Type into terminal');
    expect(ta).toHaveAttribute('autocapitalize', 'off');
    expect(ta).toHaveAttribute('autocorrect', 'off');
    expect(ta).toHaveAttribute('spellcheck', 'false');
  });

  it('on touch, Enter in an agent composer is a newline, not a send', async () => {
    vi.stubGlobal('matchMedia', (q: string) => ({ matches: q === '(pointer: coarse)' }));
    render(Composer, { paneId: 'w1:p1', blocked: false, agent: true });
    const ta = screen.getByPlaceholderText('Message the agent');
    expect(ta).toHaveAttribute('enterkeyhint', 'enter');
    await fireEvent.input(ta, { target: { value: 'line one' } });
    await fireEvent.keyDown(ta, { key: 'Enter' });
    expect(request).not.toHaveBeenCalled();
  });

  it('↑ in an empty box recalls this pane\'s sent prompts, ↓ walks back', async () => {
    render(Composer, { paneId: 'w1:p1', blocked: false, agent: true });
    const ta = screen.getByPlaceholderText('Message the agent') as HTMLTextAreaElement;
    for (const t of ['first', 'second']) {
      await fireEvent.input(ta, { target: { value: t } });
      await fireEvent.click(screen.getByLabelText('send'));
    }
    request.mockClear();
    await fireEvent.keyDown(ta, { key: 'ArrowUp' });
    expect(ta.value).toBe('second');
    await fireEvent.keyDown(ta, { key: 'ArrowUp' });
    expect(ta.value).toBe('first');
    await fireEvent.keyDown(ta, { key: 'ArrowUp' });
    expect(ta.value).toBe('first');
    await fireEvent.keyDown(ta, { key: 'ArrowDown' });
    expect(ta.value).toBe('second');
    await fireEvent.keyDown(ta, { key: 'ArrowDown' });
    expect(ta.value).toBe('');
    expect(request).not.toHaveBeenCalled();
  });

  it('history is per pane and ↑ does nothing once you type', async () => {
    history.set({ 'w1:p2': ['other pane'] });
    render(Composer, { paneId: 'w1:p1', blocked: false, agent: true });
    const ta = screen.getByPlaceholderText('Message the agent') as HTMLTextAreaElement;
    await fireEvent.keyDown(ta, { key: 'ArrowUp' });
    expect(ta.value).toBe('');
    history.set({ 'w1:p1': ['old'] });
    await fireEvent.input(ta, { target: { value: 'typing' } });
    await fireEvent.keyDown(ta, { key: 'ArrowUp' });
    expect(ta.value).toBe('typing');
  });

  it('the ↺ key cycles older prompts then wraps to empty', async () => {
    history.set({ 'w1:p9': ['a', 'b'] });
    render(Composer, { paneId: 'w1:p9', blocked: false, agent: false });
    const ta = screen.getByPlaceholderText('Type into terminal') as HTMLTextAreaElement;
    const btn = screen.getByLabelText('previous prompt');
    await fireEvent.click(btn);
    expect(ta.value).toBe('b');
    await fireEvent.click(btn);
    expect(ta.value).toBe('a');
    await fireEvent.click(btn);
    expect(ta.value).toBe('');
    await fireEvent.click(btn);
    expect(ta.value).toBe('b');
  });

  it('a recalled slash command does not open the palette', async () => {
    history.set({ 'w1:p1': ['/clear', '/compact'] });
    render(Composer, { paneId: 'w1:p1', blocked: false, agent: true });
    const ta = screen.getByPlaceholderText('Message the agent') as HTMLTextAreaElement;
    await fireEvent.keyDown(ta, { key: 'ArrowUp' });
    expect(ta.value).toBe('/compact');
    expect(screen.queryByRole('option')).toBeNull();
    await fireEvent.keyDown(ta, { key: 'ArrowUp' });
    expect(ta.value).toBe('/clear');
  });

  it('blocked panes show quick replies that send single keys', async () => {
    render(Composer, { paneId: 'w1:p1', blocked: true, agent: true });
    await fireEvent.click(screen.getByLabelText('reply 2'));
    expect(request).toHaveBeenCalledWith({ method: 'agent.send_keys', params: { target: 'w1:p1', keys: ['2'] } });
    await fireEvent.click(screen.getByLabelText('reply y'));
    expect(request).toHaveBeenCalledWith({ method: 'agent.send_keys', params: { target: 'w1:p1', keys: ['y'] } });
  });

  it('quick replies are hidden when not blocked', () => {
    render(Composer, { paneId: 'w1:p1', blocked: false, agent: true });
    expect(screen.queryByLabelText('reply 1')).toBeNull();
  });
});
