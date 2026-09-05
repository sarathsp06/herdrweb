import { describe, it, expect } from 'vitest';
import { filterSlash, SLASH_COMMANDS } from './slash';

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
});
