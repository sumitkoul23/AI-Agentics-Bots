exports.handler = async (event) => {
  const hubUrl = event.headers['x-hub-url'] || 'http://localhost:8080';

  if (event.httpMethod !== 'POST') {
    return { statusCode: 405, body: 'Method Not Allowed' };
  }

  try {
    const res = await fetch(`${hubUrl}/chat`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: event.body,
      signal: AbortSignal.timeout(30000),
    });
    const text = await res.text();
    return {
      statusCode: res.status,
      headers: { 'Content-Type': 'application/json', 'Access-Control-Allow-Origin': '*' },
      body: text,
    };
  } catch (err) {
    if (err.name === 'TimeoutError' || err.code === 'ECONNREFUSED' || err.cause?.code === 'ECONNREFUSED') {
      return {
        statusCode: 200,
        headers: { 'Content-Type': 'application/json', 'Access-Control-Allow-Origin': '*' },
        body: JSON.stringify({
          reply: "👋 I'm Bodhi — your on-device AI swarm. To connect me, run the Bodhi hub on your device and enter its URL in Settings. Download at: https://github.com/sumitkoul23/ai-agentics-bots/releases",
          agent: "bodhi",
        }),
      };
    }
    return {
      statusCode: 502,
      headers: { 'Access-Control-Allow-Origin': '*' },
      body: JSON.stringify({ error: err.message }),
    };
  }
};
