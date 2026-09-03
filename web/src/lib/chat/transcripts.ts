import type { Message } from '$lib/protocol';
import { FIXTURE_TRANSCRIPTS } from '$lib/transport/fixture';

/** Transcript for a pane. Fixture data backs the mocked dataset; live panes
 *  render via raw scrollback (pane.read) until a transcript feed exists. */
export function getTranscript(paneId: string): Message[] {
  return FIXTURE_TRANSCRIPTS[paneId] ?? [];
}
