// Slash-command palette data for agent composers.
//
// Coding agents (Claude Code and kin) accept "/command" input; on a phone the
// operator can't remember or comfortably type them. The composer surfaces a
// filtered picker while the draft is a bare slash token. The palette only fills
// the draft — the operator still reviews and sends via agent.prompt — so a
// command an agent doesn't recognize is harmless (it just gets sent as text).
//
// The builtin list below is the offline fallback; loadSlashCommands merges it
// with commands the bridge discovers on disk (user/project .claude/commands,
// skills, plugin cache) via GET /api/slash.

export interface SlashCommand {
  /** Command including its leading slash, e.g. "/clear". */
  cmd: string;
  /** One-line description shown beside the command. */
  desc: string;
  /** Where the command came from; absent on legacy data. */
  source?: 'user' | 'project' | 'skill' | 'builtin';
}

// Curated set of the common coding-agent slash commands, ordered by how often
// they matter from a phone (context/cost hygiene first, help last).
export const SLASH_COMMANDS: SlashCommand[] = [
  { cmd: '/clear', desc: 'clear conversation history', source: 'builtin' },
  { cmd: '/compact', desc: 'summarize and shrink context', source: 'builtin' },
  { cmd: '/context', desc: 'show context usage', source: 'builtin' },
  { cmd: '/cost', desc: 'show token cost', source: 'builtin' },
  { cmd: '/model', desc: 'switch model', source: 'builtin' },
  { cmd: '/review', desc: 'review changes', source: 'builtin' },
  { cmd: '/resume', desc: 'resume a session', source: 'builtin' },
  { cmd: '/status', desc: 'show status', source: 'builtin' },
  { cmd: '/init', desc: 'scaffold an agent guide', source: 'builtin' },
  { cmd: '/agents', desc: 'manage subagents', source: 'builtin' },
  { cmd: '/memory', desc: 'edit memory files', source: 'builtin' },
  { cmd: '/help', desc: 'list commands', source: 'builtin' }
];

// Module-level cache: discovery hits the filesystem server-side, so refetching
// on every keystroke-driven mount is wasteful. Keyed by cwd; 60s TTL.
const CACHE_TTL_MS = 60_000;
const cache = new Map<string, { at: number; commands: SlashCommand[] }>();

/**
 * loadSlashCommands fetches disk-discovered commands from the bridge and merges
 * them with the builtins (discovered first, deduped by cmd). Results are cached
 * per cwd for 60s; any fetch failure falls back to the builtins alone.
 */
export async function loadSlashCommands(cwd?: string): Promise<SlashCommand[]> {
  const key = cwd ?? '';
  const hit = cache.get(key);
  if (hit && Date.now() - hit.at < CACHE_TTL_MS) return hit.commands;
  try {
    const url = cwd ? `/api/slash?cwd=${encodeURIComponent(cwd)}` : '/api/slash';
    const res = await fetch(url);
    if (!res.ok) throw new Error(`slash fetch: ${res.status}`);
    const body = (await res.json()) as { commands?: SlashCommand[] };
    const merged = mergeSlashCommands(body.commands ?? [], SLASH_COMMANDS);
    cache.set(key, { at: Date.now(), commands: merged });
    return merged;
  } catch {
    return SLASH_COMMANDS;
  }
}

/**
 * mergeSlashCommands concatenates discovered before builtins, dropping later
 * duplicates by cmd. Exported for tests; pure.
 */
export function mergeSlashCommands(
  discovered: SlashCommand[],
  builtins: SlashCommand[]
): SlashCommand[] {
  const seen = new Set<string>();
  const out: SlashCommand[] = [];
  for (const c of [...discovered, ...builtins]) {
    if (seen.has(c.cmd)) continue;
    seen.add(c.cmd);
    out.push(c);
  }
  return out;
}

// A bare slash token: leading '/', then non-whitespace only. Once the operator
// types a space (arguments), the palette closes and the draft sends verbatim.
const SLASH_RE = /^\/(\S*)$/;

/**
 * filterSlash returns the commands from `commands` matching `text` when it is
 * a bare slash token (case-insensitive prefix); otherwise []. `/` alone lists
 * everything. Pure; defaults to the builtin set.
 */
export function filterSlash(text: string, commands: SlashCommand[] = SLASH_COMMANDS): SlashCommand[] {
  const m = SLASH_RE.exec(text.trimStart());
  if (!m) return [];
  const q = m[1].toLowerCase();
  return commands.filter((c) => c.cmd.slice(1).toLowerCase().startsWith(q));
}
