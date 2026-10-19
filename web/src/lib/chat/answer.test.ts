import { describe, it, expect } from 'vitest';
import { get } from 'svelte/store';
import type { Prompt } from '$lib/protocol';
import { answered, clearAnswers, connLabel, inlineAnswers, keyHint, recordAnswer, waitedFor } from './answer';

const yesNo: Prompt = { question: 'Apply?', nav: 'keys', options: [{ label: 'Yes', keys: ['y'] }, { label: 'No', keys: ['n'] }] };

describe('waitedFor', () => {
  it('formats minutes and hours', () => {
    const now = 10 * 3_600_000;
    expect(waitedFor(undefined, now)).toBe('');
    expect(waitedFor(now - 20_000, now)).toBe('just now');
    expect(waitedFor(now - 4 * 60_000, now)).toBe('4m');
    expect(waitedFor(now - 65 * 60_000, now)).toBe('1h 5m');
  });
});

describe('connLabel', () => {
  it('uses plain words', () => {
    expect(connLabel('open')).toBe('live');
    expect(connLabel('closed')).toBe('offline');
  });
});

describe('inlineAnswers', () => {
  it('offers short single questions', () => {
    expect(inlineAnswers(yesNo)?.map((o) => o.label)).toEqual(['Yes', 'No']);
  });
  it('sends long, multi-question and write-in prompts to the pane', () => {
    expect(inlineAnswers(undefined)).toBeNull();
    expect(inlineAnswers({ ...yesNo, tabs: ['A', 'Submit'] })).toBeNull();
    expect(inlineAnswers({ ...yesNo, options: [...yesNo.options, { label: 'Other', keys: ['3'], needsText: true }] })).toBeNull();
    expect(inlineAnswers({ ...yesNo, nav: '' })).toBeNull();
  });
});

describe('keyHint', () => {
  it('shows digits and letters, hides arrow sequences', () => {
    expect(keyHint(yesNo, ['y'])).toBe('y');
    expect(keyHint({ ...yesNo, nav: 'arrows' }, ['down', 'enter'])).toBe('');
  });
});

describe('answered', () => {
  it('records one answer per question and clears per pane', () => {
    recordAnswer('w1:p1', 'Q1', 'A');
    recordAnswer('w1:p1', 'Q1', 'B');
    recordAnswer('w1:p1', 'Q2', 'C');
    expect(get(answered)['w1:p1']).toEqual([
      { question: 'Q1', answer: 'B' },
      { question: 'Q2', answer: 'C' }
    ]);
    clearAnswers('w1:p1');
    expect(get(answered)['w1:p1']).toBeUndefined();
  });
});
