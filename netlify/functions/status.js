exports.handler = async (event) => {
  const hubUrl = event.headers['x-hub-url'] || 'http://localhost:8080';

  try {
    const res = await fetch(`${hubUrl}/status`, {
      signal: AbortSignal.timeout(5000),
    });
    const text = await res.text();
    return {
      statusCode: res.status,
      headers: { 'Content-Type': 'application/json', 'Access-Control-Allow-Origin': '*' },
      body: text,
    };
  } catch (_) {
    return {
      statusCode: 200,
      headers: { 'Content-Type': 'application/json', 'Access-Control-Allow-Origin': '*' },
      body: JSON.stringify({ status: 'offline', agents: 0 }),
    };
  }
};
