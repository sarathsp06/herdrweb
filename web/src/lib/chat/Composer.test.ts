import '@testing-library/jest-dom/vitest';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';

const request = vi.fn(async () => ({ ok: true }));
vi.mock('$lib/session/live', () => ({ session: () => ({ request }) }));

import Composer from './Composer.svelte';
import { draft } from '$lib/ui/state';
import { get } from 'svelte/store';

beforeEach(() => {
  request.mockClear();
  draft.set('');
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
    await waitFor(() => expect(get(draft)).toContain('/tmp/herdr-uploads/x.png'));
    expect(fetchMock).toHaveBeenCalledWith('/api/upload', expect.objectContaining({ method: 'POST' }));
  });

  it('terminal panes have no attach button', () => {
    render(Composer, { paneId: 'w1:p9', blocked: false, agent: false });
    expect(screen.queryByLabelText('attach image')).toBeNull();
  });
});
