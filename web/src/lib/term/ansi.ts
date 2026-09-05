// Minimal ANSI parser for the pane transcript. Herdr's `pane.read` with
// `format: 'ansi'` returns recent output with SGR (colour/style) escape
// sequences preserved. We render those as styled spans and strip every other
// escape (cursor moves, OSC, etc.) so control bytes never leak as glyphs.
//
// Scope is deliberately narrow: this is a monitoring transcript, not a terminal
// emulator. There is no cursor addressing, no alternate screen, no line reflow —
// just faithful colours/styles for the output Herdr already flattened to text.

/** Resolved style for a run of text. Colours are concrete CSS colour strings. */
export interface Sgr {
  fg?: string;
  bg?: string;
  bold?: boolean;
  dim?: boolean;
  italic?: boolean;
  underline?: boolean;
  strike?: boolean;
  inverse?: boolean;
}

export interface AnsiSegment {
  text: string;
  sgr: Sgr;
}

// Standard 16-colour ANSI palette (xterm defaults). Fixed values keep the
// parser self-contained; extended (256/truecolor) colours are computed exactly.
const PALETTE_16 = [
  '#000000', '#cd0000', '#00cd00', '#cdcd00', '#0000ee', '#cd00cd', '#00cdcd', '#e5e5e5',
  '#7f7f7f', '#ff0000', '#00ff00', '#ffff00', '#5c5cff', '#ff00ff', '#00ffff', '#ffffff'
] as const;

/** Resolve an xterm 256-colour index to a CSS hex string. */
function xterm256(n: number): string {
  if (n < 16) return PALETTE_16[n];
  if (n < 232) {
    const i = n - 16;
    const steps = [0, 95, 135, 175, 215, 255];
    const r = steps[Math.floor(i / 36) % 6];
    const g = steps[Math.floor(i / 6) % 6];
    const b = steps[i % 6];
    return `rgb(${r},${g},${b})`;
  }
  const v = 8 + (n - 232) * 10;
  return `rgb(${v},${v},${v})`;
}

/** Apply one decoded SGR parameter list to a style, returning the next style. */
function applySgr(cur: Sgr, params: number[]): Sgr {
  const s: Sgr = { ...cur };
  for (let i = 0; i < params.length; i++) {
    const p = params[i];
    switch (p) {
      case 0:
        // Full reset.
        for (const k of Object.keys(s)) delete (s as Record<string, unknown>)[k];
        break;
      case 1: s.bold = true; break;
      case 2: s.dim = true; break;
      case 3: s.italic = true; break;
      case 4: s.underline = true; break;
      case 7: s.inverse = true; break;
      case 9: s.strike = true; break;
      case 22: s.bold = false; s.dim = false; break;
      case 23: s.italic = false; break;
      case 24: s.underline = false; break;
      case 27: s.inverse = false; break;
      case 29: s.strike = false; break;
      case 38:
      case 48: {
        // Extended colour: 38;5;n / 38;2;r;g;b (and 48;* for background).
        const target = p === 38 ? 'fg' : 'bg';
        const mode = params[i + 1];
        if (mode === 5 && i + 2 < params.length) {
          s[target] = xterm256(params[i + 2]);
          i += 2;
        } else if (mode === 2 && i + 4 < params.length) {
          s[target] = `rgb(${params[i + 2]},${params[i + 3]},${params[i + 4]})`;
          i += 4;
        } else {
          i += 1;
        }
        break;
      }
      case 39: delete s.fg; break;
      case 49: delete s.bg; break;
      default:
        if (p >= 30 && p <= 37) s.fg = PALETTE_16[p - 30];
        else if (p >= 40 && p <= 47) s.bg = PALETTE_16[p - 40];
        else if (p >= 90 && p <= 97) s.fg = PALETTE_16[p - 90 + 8];
        else if (p >= 100 && p <= 107) s.bg = PALETTE_16[p - 100 + 8];
        // Unknown parameters are ignored.
        break;
    }
  }
  return s;
}

/**
 * Parse a batch of already-line-split text (SGR escapes intact) into per-line
 * segment arrays. SGR state carries across line boundaries, matching terminal
 * behaviour where a colour set on one line persists until reset. Non-SGR escape
 * sequences (CSI without final `m`, OSC, single-char ESC) are stripped.
 */
export function parseAnsiLines(lines: string[]): AnsiSegment[][] {
  let sgr: Sgr = {};
  const out: AnsiSegment[][] = [];
  for (const line of lines) {
    const segs: AnsiSegment[] = [];
    let text = '';
    const flush = () => {
      if (text) {
        segs.push({ text, sgr });
        text = '';
      }
    };
    let i = 0;
    while (i < line.length) {
      const ch = line[i];
      if (ch !== '\x1b') {
        text += ch;
        i++;
        continue;
      }
      // ESC sequence.
      const next = line[i + 1];
      if (next === '[') {
        // CSI: ESC [ params intermediate final
        let j = i + 2;
        while (j < line.length && /[0-9;:?]/.test(line[j])) j++;
        const body = line.slice(i + 2, j);
        const final = line[j];
        if (final === 'm') {
          flush();
          const params = body === '' ? [0] : body.split(';').map((p) => parseInt(p, 10) || 0);
          sgr = applySgr(sgr, params);
        }
        // Any CSI (SGR or otherwise) is consumed and not rendered.
        i = j + 1;
      } else if (next === ']') {
        // OSC: ESC ] ... (terminated by BEL or ESC \).
        let j = i + 2;
        while (j < line.length && line[j] !== '\x07' && !(line[j] === '\x1b' && line[j + 1] === '\\')) j++;
        i = line[j] === '\x1b' ? j + 2 : j + 1;
      } else {
        // nF escape (e.g. ESC ( B): ESC, zero+ intermediates (0x20-0x2F),
        // then one final byte (0x30-0x7E). Consume the whole thing.
        let j = i + 1;
        while (j < line.length && line[j] >= '\x20' && line[j] <= '\x2f') j++;
        i = j < line.length ? j + 1 : j;
      }
    }
    flush();
    out.push(segs);
  }
  return out;
}

/** Build an inline CSS style string for a segment, honouring inverse video. */
export function segStyle(sgr: Sgr): string {
  let fg = sgr.fg;
  let bg = sgr.bg;
  if (sgr.inverse) {
    // Swap; fall back to the pane's default fg/bg so inverse is always visible.
    const nf = bg ?? 'var(--bg)';
    const nb = fg ?? 'var(--text-1)';
    fg = nf;
    bg = nb;
  }
  const parts: string[] = [];
  if (fg) parts.push(`color:${fg}`);
  if (bg) parts.push(`background:${bg}`);
  if (sgr.bold) parts.push('font-weight:600');
  if (sgr.dim) parts.push('opacity:0.7');
  if (sgr.italic) parts.push('font-style:italic');
  const deco: string[] = [];
  if (sgr.underline) deco.push('underline');
  if (sgr.strike) deco.push('line-through');
  if (deco.length) parts.push(`text-decoration:${deco.join(' ')}`);
  return parts.join(';');
}
