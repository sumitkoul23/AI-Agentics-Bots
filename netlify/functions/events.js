// Netlify Functions don't support long-lived SSE connections.
// This endpoint returns a single synthetic "connected" event so the
// frontend EventSource handshake succeeds, then the client reconnects
// periodically (standard SSE retry behaviour).
exports.handler = async (event) => {
  const hubUrl = event.headers['x-hub-url'] || 'http://localhost:8080';

  // Try to forward to the real hub SSE for a brief window
  try {
    const res = await fetch(`${hubUrl}/events`, {
      signal: AbortSignal.timeout(2000),
    });
    // If hub responds, return whatever it sent (may be empty on timeout)
    const text = await res.text();
    return {
      statusCode: 200,
      headers: {
        'Content-Type': 'text/event-stream',
        'Cache-Control': 'no-cache',
        'Access-Control-Allow-Origin': '*',
      },
      body: text || 'data: {"type":"ping"}\n\n',
    };
  } catch (_) {
    return {
      statusCode: 200,
      headers: {
        'Content-Type': 'text/event-stream',
        'Cache-Control': 'no-cache',
        'Access-Control-Allow-Origin': '*',
      },
      body: 'data: {"type":"ping"}\n\n',
    };
  }
};
