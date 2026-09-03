import '@testing-library/jest-dom/vitest';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';

const request = vi.fn(async () => ({ ok: true }));
vi.mock('$lib/session/live', () => ({ session: () => ({ request }) }));

import Composer from './Composer.svelte';

beforeEach(() => request.mockClear());

describe('Composer', () => {
  it('nav keys send agent.send_keys as single-element arrays', async () => {
    render(Composer, { paneId: 'w1:p1', blocked: false });
    await fireEvent.click(screen.getByLabelText('up'));
    expect(request).toHaveBeenCalledWith({ method: 'agent.send_keys', params: { target: 'w1:p1', keys: ['up'] } });
    await fireEvent.click(screen.getByLabelText('enter'));
    expect(request).toHaveBeenCalledWith({ method: 'agent.send_keys', params: { target: 'w1:p1', keys: ['enter'] } });
  });

  it('send button submits typed text via agent.prompt', async () => {
    render(Composer, { paneId: 'w1:p1', blocked: false });
    const ta = screen.getByPlaceholderText('Message the agent');
    await fireEvent.input(ta, { target: { value: 'do the thing' } });
    await fireEvent.click(screen.getByLabelText('send'));
    expect(request).toHaveBeenCalledWith(
      expect.objectContaining({ method: 'agent.prompt', params: expect.objectContaining({ target: 'w1:p1', text: 'do the thing' }) })
    );
  });
});
