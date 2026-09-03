/// <reference types="@sveltejs/kit" />
/// <reference lib="esnext" />
/// <reference lib="webworker" />

// Herdr Web service worker. Its only job is Web Push: show a notification when
// the bridge reports a blocked agent, and focus the pane when tapped. This is a
// live WebSocket app, so it deliberately does no offline asset caching.

const sw = self as unknown as ServiceWorkerGlobalScope;

sw.addEventListener('install', () => {
  sw.skipWaiting();
});

sw.addEventListener('activate', (event) => {
  event.waitUntil(sw.clients.claim());
});

interface PushPayload {
  title?: string;
  body?: string;
  url?: string;
}

sw.addEventListener('push', (event) => {
  let data: PushPayload = {};
  try {
    if (event.data) data = event.data.json() as PushPayload;
  } catch {
    if (event.data) data = { body: event.data.text() };
  }
  const title = data.title ?? 'Herdr';
  event.waitUntil(
    sw.registration.showNotification(title, {
      body: data.body ?? 'An agent needs you.',
      icon: '/icon-192.png',
      badge: '/icon-192.png',
      tag: data.url ?? 'herdr',
      renotify: true,
      data: { url: data.url ?? '/' }
    } as NotificationOptions)
  );
});

sw.addEventListener('notificationclick', (event) => {
  event.notification.close();
  const data = event.notification.data;
  // notification.data round-trips the { url } we set above; narrow before use.
  const target =
    data && typeof data === 'object' && 'url' in data && typeof data.url === 'string'
      ? data.url
      : '/';
  event.waitUntil(
    (async () => {
      // matchAll({ type: 'window' }) yields window clients; the lib types it as
      // the Client base, so narrow to WindowClient for focus()/navigate().
      const clients = (await sw.clients.matchAll({
        type: 'window',
        includeUncontrolled: true
      })) as readonly WindowClient[];
      for (const client of clients) {
        await client.focus();
        try {
          await client.navigate(target);
        } catch {
          // cross-origin or unsupported; ignore
        }
        return;
      }
      await sw.clients.openWindow(target);
    })()
  );
});
