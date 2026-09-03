// Client-side Web Push enrolment. Called from a user gesture (the Settings
// "Push when blocked" toggle) so the permission prompt is allowed. Registers
// with the bridge's VAPID key and posts the subscription back for delivery.

export type PushResult =
  | { ok: true }
  | { ok: false; reason: 'unsupported' | 'insecure' | 'denied' | 'nokey' | 'error'; detail?: string };
function urlBase64ToUint8Array(base64: string): Uint8Array<ArrayBuffer> {
  const padding = '='.repeat((4 - (base64.length % 4)) % 4);
  const normalized = (base64 + padding).replace(/-/g, '+').replace(/_/g, '/');
  const raw = atob(normalized);
  const buf = new ArrayBuffer(raw.length);
  const out = new Uint8Array(buf);
  for (let i = 0; i < raw.length; i++) out[i] = raw.charCodeAt(i);
  return out;
}

async function fetchVapidKey(): Promise<string> {
  const res = await fetch('/api/push/key');
  if (!res.ok) return '';
  const body: unknown = await res.json();
  if (body && typeof body === 'object' && 'key' in body && typeof body.key === 'string') {
    return body.key;
  }
  return '';
}

/** Enrol this browser for Web Push. Safe to call repeatedly (idempotent). */
export async function enablePush(): Promise<PushResult> {
  const supported =
    typeof navigator !== 'undefined' &&
    'serviceWorker' in navigator &&
    typeof window !== 'undefined' &&
    'PushManager' in window &&
    'Notification' in window;
  if (!supported) return { ok: false, reason: 'unsupported' };
  // Service workers and Push require a secure context (HTTPS or localhost).
  // Over Tailscale that means `tailscale serve --https`.
  if (!window.isSecureContext) return { ok: false, reason: 'insecure' };

  try {
    const permission = await Notification.requestPermission();
    if (permission !== 'granted') return { ok: false, reason: 'denied' };

    const key = await fetchVapidKey();
    if (!key) return { ok: false, reason: 'nokey' };

    const reg = await navigator.serviceWorker.ready;
    const existing = await reg.pushManager.getSubscription();
    const sub =
      existing ??
      (await reg.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(key)
      }));

    const res = await fetch('/api/push/subscribe', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify(sub.toJSON())
    });
    if (!res.ok) return { ok: false, reason: 'error', detail: `subscribe ${res.status}` };
    return { ok: true };
  } catch (e) {
    return { ok: false, reason: 'error', detail: e instanceof Error ? e.message : String(e) };
  }
}
