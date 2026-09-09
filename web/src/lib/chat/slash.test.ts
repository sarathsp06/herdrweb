import { describe, it, expect } from 'vitest';
import { filterSlash, mergeSlashCommands, SLASH_COMMANDS, type SlashCommand } from './slash';

describe('filterSlash', () => {
  it('lists everything for a bare slash', () => {
    expect(filterSlash('/')).toEqual(SLASH_COMMANDS);
  });

  it('prefix-matches case-insensitively', () => {
    expect(filterSlash('/CL').map((c) => c.cmd)).toEqual(['/clear']);
    expect(filterSlash('/co').map((c) => c.cmd)).toEqual(['/compact', '/context', '/cost']);
  });

  it('ignores leading whitespace', () => {
    expect(filterSlash('  /re').map((c) => c.cmd)).toEqual(['/review', '/resume']);
  });

  it('returns nothing once the token has a space or no leading slash', () => {
    expect(filterSlash('/clear ')).toEqual([]);
    expect(filterSlash('/model opus')).toEqual([]);
    expect(filterSlash('hello')).toEqual([]);
    expect(filterSlash('')).toEqual([]);
  });

  it('returns nothing when nothing matches', () => {
    expect(filterSlash('/zzz')).toEqual([]);
  });

  it('filters an explicitly supplied command list', () => {
    const commands: SlashCommand[] = [
      { cmd: '/deploy', desc: 'ship it', source: 'user' },
      { cmd: '/debug', desc: 'debug', source: 'project' }
    ];
    expect(filterSlash('/de', commands).map((c) => c.cmd)).toEqual(['/deploy', '/debug']);
    expect(filterSlash('/clear', commands)).toEqual([]);
  });
});

describe('mergeSlashCommands', () => {
  it('puts discovered commands first and dedupes by cmd (discovered wins)', () => {
    const discovered: SlashCommand[] = [
      { cmd: '/deploy', desc: 'ship it', source: 'user' },
      { cmd: '/clear', desc: 'user override of a builtin', source: 'user' }
    ];
    const merged = mergeSlashCommands(discovered, SLASH_COMMANDS);
    expect(merged.slice(0, 2)).toEqual(discovered);
    expect(merged.filter((c) => c.cmd === '/clear')).toEqual([discovered[1]]);
    expect(merged.length).toBe(SLASH_COMMANDS.length + 1);
  });
});
