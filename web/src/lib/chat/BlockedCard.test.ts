import '@testing-library/jest-dom/vitest';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';

const request = vi.fn(async () => ({ ok: true }));
vi.mock('$lib/session/live', () => ({ session: () => ({ request }) }));

import BlockedCard from './BlockedCard.svelte';
import Composer from './Composer.svelte';

beforeEach(() => request.mockClear());

describe('BlockedCard', () => {
  it('maps answer buttons to agent.send_keys y/a/n/esc', async () => {
    render(BlockedCard, { paneId: 'w1:p2', question: 'Apply patch?', terminal: '? (y/n)', age: '4m' });
    await fireEvent.click(screen.getByText('No'));
    expect(request).toHaveBeenCalledWith({ method: 'agent.send_keys', params: { target: 'w1:p2', keys: 'n' } });
    await fireEvent.click(screen.getByText('Yes', { exact: true }));
    expect(request).toHaveBeenCalledWith({ method: 'agent.send_keys', params: { target: 'w1:p2', keys: 'y' } });
    await fireEvent.click(screen.getByText("Yes, don't ask"));
    expect(request).toHaveBeenCalledWith({ method: 'agent.send_keys', params: { target: 'w1:p2', keys: 'a' } });
    await fireEvent.click(screen.getByText('esc'));
    expect(request).toHaveBeenCalledWith({ method: 'agent.send_keys', params: { target: 'w1:p2', keys: 'esc' } });
  });
});

describe('Composer', () => {
  it('shows blocked quick-send chips when blocked', () => {
    render(Composer, { paneId: 'w1:p2', blocked: true });
    for (const c of ['y', 'n', 'esc', 'ctrl+c', '/status']) expect(screen.getByText(c)).toBeInTheDocument();
    expect(screen.queryByText('run tests')).toBeNull();
  });

  it('shows normal chips when not blocked and sends free text via agent.prompt', async () => {
    render(Composer, { paneId: 'w1:p1', blocked: false });
    for (const c of ['continue', 'run tests', 'git diff', 'esc', 'ctrl+c']) expect(screen.getByText(c)).toBeInTheDocument();
    await fireEvent.click(screen.getByText('run tests'));
    expect(request).toHaveBeenCalledWith(
      expect.objectContaining({ method: 'agent.prompt', params: expect.objectContaining({ target: 'w1:p1', text: 'run tests' }) })
    );
  });
});
