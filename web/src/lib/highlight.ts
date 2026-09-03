import type { Highlighter } from 'shiki';

// A Shiki theme mapped onto the README's syntax palette (hex approximations of
// the documented oklch values), so highlighting is real (not hand-rolled) yet
// matches the spec's colours. The surface is applied via CSS, not the theme.
const herdrTheme = {
  name: 'herdr',
  type: 'dark',
  colors: { 'editor.background': '#0d0d0f', 'editor.foreground': '#d4d4d8' },
  settings: [
    { settings: { foreground: '#d4d4d8' } },
    { scope: ['keyword', 'storage', 'storage.type', 'keyword.control', 'keyword.operator.new'], settings: { foreground: '#c191f0' } },
    { scope: ['string', 'string.quoted', 'punctuation.definition.string'], settings: { foreground: '#dcc27a' } },
    { scope: ['constant.numeric', 'constant.language'], settings: { foreground: '#7db6f0' } },
    { scope: ['entity.name.function', 'support.function', 'meta.function-call'], settings: { foreground: '#5fc7c7' } },
    { scope: ['entity.name.type', 'entity.name.class', 'support.class', 'entity.name.tag'], settings: { foreground: '#6fd0a8' } },
    { scope: ['comment', 'punctuation.definition.comment'], settings: { foreground: '#5c5c63', fontStyle: 'italic' } },
    { scope: ['punctuation', 'meta.brace'], settings: { foreground: '#8a8a92' } }
  ]
} as const;

let hlPromise: Promise<Highlighter> | null = null;
const LANGS = ['ts', 'js', 'svelte', 'sql', 'json', 'bash'];

async function get(): Promise<Highlighter> {
  if (!hlPromise) {
    const { createHighlighter } = await import('shiki');
    hlPromise = createHighlighter({ themes: [herdrTheme as any], langs: LANGS });
  }
  return hlPromise;
}

function normLang(lang: string): string {
  const l = lang.toLowerCase();
  if (l === 'typescript') return 'ts';
  if (l === 'javascript') return 'js';
  return LANGS.includes(l) ? l : 'ts';
}

/** Highlight code to HTML (token spans only, no <pre>/<code> wrappers). */
export async function highlight(code: string, lang: string): Promise<string> {
  const hl = await get();
  const html = hl.codeToHtml(code, { lang: normLang(lang), theme: 'herdr' });
  // strip the outer <pre class..><code>..</code></pre>, keep the lines
  const m = html.match(/<code[^>]*>([\s\S]*)<\/code>/);
  return m ? m[1] : escapeHtml(code);
}

export function escapeHtml(s: string): string {
  return s.replace(/[&<>]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' })[c] as string);
}
