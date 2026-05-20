self.addEventListener('install', () => self.skipWaiting());
self.addEventListener('activate', e => e.waitUntil(self.clients.claim()));
self.addEventListener('fetch', e => {
  if (e.request.method !== 'GET') return;
  if (e.request.url.includes('/chat') || e.request.url.includes('/models')) return;
  e.respondWith(fetch(e.request).catch(() => caches.match(e.request)));
});
