import type { Highlighter } from 'shiki';

// The highlighter emits token colours as CSS variables (never baked hex), so
// syntax adapts to whatever theme is active — dark, solarized-light, or paper.
// Each --syn-* token is themed in tokens.css; the surface is applied via CSS.
const herdrTheme = {
  name: 'herdr',
  type: 'dark',
  colors: { 'editor.background': 'var(--code-surface)', 'editor.foreground': 'var(--syn-default)' },
  settings: [
    { settings: { foreground: 'var(--syn-default)' } },
    { scope: ['keyword', 'storage', 'storage.type', 'keyword.control', 'keyword.operator.new'], settings: { foreground: 'var(--syn-keyword)' } },
    { scope: ['string', 'string.quoted', 'punctuation.definition.string'], settings: { foreground: 'var(--syn-string)' } },
    { scope: ['constant.numeric', 'constant.language'], settings: { foreground: 'var(--syn-number)' } },
    { scope: ['entity.name.function', 'support.function', 'meta.function-call'], settings: { foreground: 'var(--syn-func)' } },
    { scope: ['entity.name.type', 'entity.name.class', 'support.class', 'entity.name.tag'], settings: { foreground: 'var(--syn-comp)' } },
    { scope: ['comment', 'punctuation.definition.comment'], settings: { foreground: 'var(--syn-comment)', fontStyle: 'italic' } },
    { scope: ['punctuation', 'meta.brace'], settings: { foreground: 'var(--syn-punct)' } }
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
