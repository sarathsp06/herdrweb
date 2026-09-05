// Slash-command palette data for agent composers.
//
// Coding agents (Claude Code and kin) accept "/command" input; on a phone the
// operator can't remember or comfortably type them. The composer surfaces a
// filtered picker while the draft is a bare slash token. The palette only fills
// the draft — the operator still reviews and sends via agent.prompt — so a
// command an agent doesn't recognize is harmless (it just gets sent as text).

export interface SlashCommand {
  /** Command including its leading slash, e.g. "/clear". */
  cmd: string;
  /** One-line description shown beside the command. */
  desc: string;
}

// Curated set of the common coding-agent slash commands, ordered by how often
// they matter from a phone (context/cost hygiene first, help last).
export const SLASH_COMMANDS: SlashCommand[] = [
  { cmd: '/clear', desc: 'clear conversation history' },
  { cmd: '/compact', desc: 'summarize and shrink context' },
  { cmd: '/context', desc: 'show context usage' },
  { cmd: '/cost', desc: 'show token cost' },
  { cmd: '/model', desc: 'switch model' },
  { cmd: '/review', desc: 'review changes' },
  { cmd: '/resume', desc: 'resume a session' },
  { cmd: '/status', desc: 'show status' },
  { cmd: '/init', desc: 'scaffold an agent guide' },
  { cmd: '/agents', desc: 'manage subagents' },
  { cmd: '/memory', desc: 'edit memory files' },
  { cmd: '/help', desc: 'list commands' }
];

// A bare slash token: leading '/', then non-whitespace only. Once the operator
// types a space (arguments), the palette closes and the draft sends verbatim.
const SLASH_RE = /^\/(\S*)$/;

/**
 * filterSlash returns the commands matching `text` when it is a bare slash
 * token (case-insensitive prefix); otherwise []. `/` alone lists everything.
 */
export function filterSlash(text: string): SlashCommand[] {
  const m = SLASH_RE.exec(text.trimStart());
  if (!m) return [];
  const q = m[1].toLowerCase();
  return SLASH_COMMANDS.filter((c) => c.cmd.slice(1).toLowerCase().startsWith(q));
}
