import { describe, it, expect } from 'vitest';
import { parseAnsiLines, segStyle } from './ansi';

const E = '\x1b';

describe('parseAnsiLines', () => {
  it('returns a single default segment for plain text', () => {
    expect(parseAnsiLines(['hello world'])).toEqual([[{ text: 'hello world', sgr: {} }]]);
  });

  it('splits on an SGR colour change and resolves the 16-colour palette', () => {
    const [line] = parseAnsiLines([`plain ${E}[31mred${E}[0m tail`]);
    expect(line).toEqual([
      { text: 'plain ', sgr: {} },
      { text: 'red', sgr: { fg: '#cd0000' } },
      { text: ' tail', sgr: {} }
    ]);
  });

  it('carries SGR state across line boundaries until reset', () => {
    const rows = parseAnsiLines([`${E}[32mgreen start`, 'still green', `done${E}[0m`]);
    expect(rows[0]).toEqual([{ text: 'green start', sgr: { fg: '#00cd00' } }]);
    expect(rows[1]).toEqual([{ text: 'still green', sgr: { fg: '#00cd00' } }]);
    expect(rows[2]).toEqual([{ text: 'done', sgr: { fg: '#00cd00' } }]);
  });

  it('accumulates bold + underline + bright fg', () => {
    const [line] = parseAnsiLines([`${E}[1;4;92mhi`]);
    expect(line).toEqual([{ text: 'hi', sgr: { bold: true, underline: true, fg: '#00ff00' } }]);
  });

  it('decodes 256-colour and truecolor foreground', () => {
    const [line] = parseAnsiLines([`${E}[38;5;196mx${E}[38;2;10;20;30my`]);
    expect(line[0].sgr.fg).toBe('rgb(255,0,0)');
    expect(line[1].sgr.fg).toBe('rgb(10,20,30)');
  });

  it('applies background codes and clears them with 49', () => {
    const [line] = parseAnsiLines([`${E}[44mb${E}[49mn`]);
    expect(line[0].sgr.bg).toBe('#0000ee');
    expect(line[1].sgr.bg).toBeUndefined();
  });

  it('empty CSI m is a full reset', () => {
    const [line] = parseAnsiLines([`${E}[1mbold${E}[mafter`]);
    expect(line[0].sgr).toEqual({ bold: true });
    expect(line[1].sgr).toEqual({});
  });

  it('strips non-SGR CSI (cursor move) without emitting glyphs', () => {
    const [line] = parseAnsiLines([`${E}[2Ka${E}[10;5Hb`]);
    expect(line).toEqual([{ text: 'ab', sgr: {} }]);
  });

  it('strips OSC sequences terminated by BEL', () => {
    const [line] = parseAnsiLines([`${E}]0;window title\x07visible`]);
    expect(line).toEqual([{ text: 'visible', sgr: {} }]);
  });

  it('drops two-char ESC sequences like charset select', () => {
    const [line] = parseAnsiLines([`${E}(Bplain`]);
    expect(line).toEqual([{ text: 'plain', sgr: {} }]);
  });
});

describe('segStyle', () => {
  it('emits colour, weight and decorations', () => {
    expect(segStyle({ fg: '#cd0000', bold: true, underline: true })).toBe(
      'color:#cd0000;font-weight:600;text-decoration:underline'
    );
  });

  it('swaps fg/bg for inverse with sensible defaults', () => {
    expect(segStyle({ inverse: true })).toBe('color:var(--bg);background:var(--text-1)');
    expect(segStyle({ fg: '#111', bg: '#eee', inverse: true })).toBe('color:#eee;background:#111');
  });

  it('is empty for the default style', () => {
    expect(segStyle({})).toBe('');
  });
});
