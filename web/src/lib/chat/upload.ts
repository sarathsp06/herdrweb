// Image upload for agent composers.
//
// There is no image-input RPC to proxy: coding agents read images by file path.
// So the bridge writes the uploaded bytes to a host-local file (it shares the
// agent's host) and returns the absolute path, which the composer drops into
// the draft for the operator to send via agent.prompt.

/** uploadImage POSTs an image blob to the bridge and returns the host path it
 *  was written to. Throws on a non-2xx response or a malformed reply. */
export async function uploadImage(blob: Blob): Promise<string> {
  const res = await fetch('/api/upload', {
    method: 'POST',
    headers: { 'content-type': blob.type || 'application/octet-stream' },
    body: blob
  });
  if (!res.ok) throw new Error(`upload failed (${res.status})`);
  const body: unknown = await res.json();
  if (body && typeof body === 'object' && 'path' in body && typeof body.path === 'string' && body.path) {
    return body.path;
  }
  throw new Error('upload: malformed response');
}
