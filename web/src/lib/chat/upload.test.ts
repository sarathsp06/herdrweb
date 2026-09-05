import { describe, it, expect, vi, afterEach } from 'vitest';
import { uploadImage } from './upload';

afterEach(() => vi.unstubAllGlobals());

describe('uploadImage', () => {
  it('POSTs the blob to /api/upload and returns the path', async () => {
    const fetchMock = vi.fn(async () => ({ ok: true, json: async () => ({ path: '/tmp/herdr-uploads/x.png' }) }));
    vi.stubGlobal('fetch', fetchMock);

    const path = await uploadImage(new Blob([new Uint8Array([1, 2, 3])], { type: 'image/png' }));

    expect(path).toBe('/tmp/herdr-uploads/x.png');
    const [url, opts] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
    expect(url).toBe('/api/upload');
    expect(opts.method).toBe('POST');
    expect((opts.headers as Record<string, string>)['content-type']).toBe('image/png');
  });

  it('throws on a non-ok response', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => ({ ok: false, status: 415 })));
    await expect(uploadImage(new Blob([], { type: 'image/png' }))).rejects.toThrow(/415/);
  });

  it('throws on a malformed response body', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => ({ ok: true, json: async () => ({}) })));
    await expect(uploadImage(new Blob([], { type: 'image/png' }))).rejects.toThrow(/malformed/);
  });
});
