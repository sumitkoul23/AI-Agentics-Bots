// Priya service worker — offline shell + network-first for API
const CACHE = 'priya-v1'
const SHELL = ['/']

self.addEventListener('install', e => {
  e.waitUntil(caches.open(CACHE).then(c => c.addAll(SHELL)))
  self.skipWaiting()
})

self.addEventListener('activate', e => {
  e.waitUntil(
    caches.keys().then(keys =>
      Promise.all(keys.filter(k => k !== CACHE).map(k => caches.delete(k)))
    )
  )
  self.clients.claim()
})

self.addEventListener('fetch', e => {
  const url = new URL(e.request.url)
  // Always network for API endpoints
  const apiPaths = ['/chat', '/status', '/memory', '/agents', '/peers', '/mesh']
  if (apiPaths.some(p => url.pathname.startsWith(p))) return

  // Cache-first for app shell, fall back to network
  e.respondWith(
    caches.match(e.request).then(cached => {
      const network = fetch(e.request).then(resp => {
        if (resp.ok) {
          caches.open(CACHE).then(c => c.put(e.request, resp.clone()))
        }
        return resp
      }).catch(() => cached)
      return cached || network
    })
  )
})
