/**
 * LLM client — Ollama-first, cloud fallback, offline last.
 * Same strategy as the rest of the swarm: local & free by default, premium when
 * keys are set, and always a result (offline) so the agent is never blocked.
 */
const OLLAMA_HOST = process.env.OLLAMA_HOST || "http://localhost:11434";
const OLLAMA_MODEL = process.env.OLLAMA_MODEL || "llama3.2";

async function tryOllama(system, user) {
  const res = await fetch(`${OLLAMA_HOST}/api/chat`, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify({
      model: OLLAMA_MODEL,
      stream: false,
      messages: [{ role: "system", content: system }, { role: "user", content: user }],
    }),
  });
  if (!res.ok) throw new Error(`Ollama ${res.status}`);
  return (await res.json()).message?.content?.trim();
}

async function tryGroq(system, user) {
  const key = process.env.GROQ_API_KEY;
  if (!key) throw new Error("no GROQ_API_KEY");
  const res = await fetch("https://api.groq.com/openai/v1/chat/completions", {
    method: "POST",
    headers: { "content-type": "application/json", authorization: `Bearer ${key}` },
    body: JSON.stringify({
      model: process.env.GROQ_MODEL || "llama-3.3-70b-versatile",
      messages: [{ role: "system", content: system }, { role: "user", content: user }],
    }),
  });
  if (!res.ok) throw new Error(`Groq ${res.status}`);
  return (await res.json()).choices?.[0]?.message?.content?.trim();
}

async function tryGemini(system, user) {
  const key = process.env.GEMINI_API_KEY;
  if (!key) throw new Error("no GEMINI_API_KEY");
  const model = process.env.GEMINI_MODEL || "gemini-2.0-flash";
  const res = await fetch(
    `https://generativelanguage.googleapis.com/v1beta/models/${model}:generateContent?key=${key}`,
    {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        systemInstruction: { parts: [{ text: system }] },
        contents: [{ role: "user", parts: [{ text: user }] }],
      }),
    }
  );
  if (!res.ok) throw new Error(`Gemini ${res.status}`);
  return (await res.json()).candidates?.[0]?.content?.parts?.[0]?.text?.trim();
}

async function tryClaude(system, user) {
  const key = process.env.ANTHROPIC_API_KEY;
  if (!key) throw new Error("no ANTHROPIC_API_KEY");
  const res = await fetch("https://api.anthropic.com/v1/messages", {
    method: "POST",
    headers: { "content-type": "application/json", "x-api-key": key, "anthropic-version": "2023-06-01" },
    body: JSON.stringify({
      model: process.env.ANTHROPIC_MODEL || "claude-sonnet-4-6",
      max_tokens: 1500,
      system,
      messages: [{ role: "user", content: user }],
    }),
  });
  if (!res.ok) throw new Error(`Claude ${res.status}`);
  return (await res.json()).content?.[0]?.text?.trim();
}

const PROVIDERS = [
  ["ollama", tryOllama],
  ["groq", tryGroq],
  ["gemini", tryGemini],
  ["claude", tryClaude],
];

/** Returns { provider, content } or { provider: null } if no model reachable. */
export async function chat(system, user) {
  for (const [name, fn] of PROVIDERS) {
    try {
      const content = await fn(system, user);
      if (content) return { provider: name, content };
    } catch {
      /* try next */
    }
  }
  return { provider: null, content: null };
}
