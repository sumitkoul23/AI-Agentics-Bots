/**
 * LLM client with Ollama-first, cloud-fallback, and offline-last strategy.
 *
 * Mirrors the repo's design ethos (see FusionAI): local & free by default,
 * premium models when keys are present, and ALWAYS a useful result even with no
 * model available at all (deterministic extractive fallback).
 *
 * Order of preference:
 *   1. Ollama          (local, no key)        — OLLAMA_HOST, OLLAMA_MODEL
 *   2. Groq            (free tier)            — GROQ_API_KEY
 *   3. Google Gemini   (free tier)            — GEMINI_API_KEY
 *   4. Anthropic Claude (premium)             — ANTHROPIC_API_KEY
 *   5. OpenAI          (premium)              — OPENAI_API_KEY
 *   6. Offline extractive summariser          — always available
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
      messages: [
        { role: "system", content: system },
        { role: "user", content: user },
      ],
    }),
  });
  if (!res.ok) throw new Error(`Ollama ${res.status}`);
  const data = await res.json();
  return data.message?.content?.trim();
}

async function tryGroq(system, user) {
  const key = process.env.GROQ_API_KEY;
  if (!key) throw new Error("no GROQ_API_KEY");
  const res = await fetch("https://api.groq.com/openai/v1/chat/completions", {
    method: "POST",
    headers: { "content-type": "application/json", authorization: `Bearer ${key}` },
    body: JSON.stringify({
      model: process.env.GROQ_MODEL || "llama-3.3-70b-versatile",
      messages: [
        { role: "system", content: system },
        { role: "user", content: user },
      ],
    }),
  });
  if (!res.ok) throw new Error(`Groq ${res.status}`);
  const data = await res.json();
  return data.choices?.[0]?.message?.content?.trim();
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
  const data = await res.json();
  return data.candidates?.[0]?.content?.parts?.[0]?.text?.trim();
}

async function tryClaude(system, user) {
  const key = process.env.ANTHROPIC_API_KEY;
  if (!key) throw new Error("no ANTHROPIC_API_KEY");
  const res = await fetch("https://api.anthropic.com/v1/messages", {
    method: "POST",
    headers: {
      "content-type": "application/json",
      "x-api-key": key,
      "anthropic-version": "2023-06-01",
    },
    body: JSON.stringify({
      model: process.env.ANTHROPIC_MODEL || "claude-sonnet-4-6",
      max_tokens: 2048,
      system,
      messages: [{ role: "user", content: user }],
    }),
  });
  if (!res.ok) throw new Error(`Claude ${res.status}`);
  const data = await res.json();
  return data.content?.[0]?.text?.trim();
}

async function tryOpenAI(system, user) {
  const key = process.env.OPENAI_API_KEY;
  if (!key) throw new Error("no OPENAI_API_KEY");
  const res = await fetch("https://api.openai.com/v1/chat/completions", {
    method: "POST",
    headers: { "content-type": "application/json", authorization: `Bearer ${key}` },
    body: JSON.stringify({
      model: process.env.OPENAI_MODEL || "gpt-4o-mini",
      messages: [
        { role: "system", content: system },
        { role: "user", content: user },
      ],
    }),
  });
  if (!res.ok) throw new Error(`OpenAI ${res.status}`);
  const data = await res.json();
  return data.choices?.[0]?.message?.content?.trim();
}

const PROVIDERS = [
  ["ollama", tryOllama],
  ["groq", tryGroq],
  ["gemini", tryGemini],
  ["claude", tryClaude],
  ["openai", tryOpenAI],
];

/**
 * Run the persona prompt against the best available model.
 * @returns {Promise<{ provider: string, content: string }>}
 */
export async function summarise(system, user, deck) {
  const errors = [];
  for (const [name, fn] of PROVIDERS) {
    try {
      const content = await fn(system, user);
      if (content) return { provider: name, content };
    } catch (e) {
      errors.push(`${name}: ${e.message}`);
    }
  }
  // Everything unavailable — deterministic offline briefing.
  return { provider: "offline-extractive", content: offlineBriefing(deck), notes: errors };
}

// ── Offline extractive Head-of-Department briefing ───────────────────────
const CONFIG_RE = /\b([A-Z]{1,3}\d{2,}|\d{4,}|Z\d{2,}|[A-Z]\d{2})\b/;
const COMPLIANCE_RE = /\b(gst|tax|licen[sc]e|authoriz|reconcil|incoterm|payment|account group|register)\w*/i;

export function offlineBriefing(deck) {
  const flat = deck.slides.flat();
  const config = flat.filter((l) => CONFIG_RE.test(l));
  const compliance = flat.filter((l) => COMPLIANCE_RE.test(l));
  const title = flat.find((l) => /[A-Z]{4,}/.test(l)) || deck.source;

  const steps = deck.slides
    .map((s, i) => {
      const head = s.find((l) => l.length > 8) || s[0] || "(slide)";
      return `${i + 1}. ${head}`;
    })
    .join("\n");

  return `## Executive Summary
This is an automated extractive briefing of **${deck.source}** (${deck.slideCount} slides). No language model was reachable, so the points below are pulled directly from the deck's text. Topic appears to be: **${title}**.

## What This Process Does
The deck documents a step-by-step procedure across ${deck.slideCount} slides. Connect a model (Ollama/Groq/Gemini/Claude/OpenAI) for a fully written narrative.

## Key Steps / Structure
${steps}

## Critical Data & Configuration
${config.length ? config.map((c) => `- ${c}`).join("\n") : "- (no explicit codes/values detected)"}

## Risks, Compliance & Watch-outs
${compliance.length ? compliance.map((c) => `- ${c}`).join("\n") : "- (no compliance-flagged lines detected)"}

## Head of Department's Recommendations
- Verify every value in *Critical Data & Configuration* against the system of record before sign-off.
- Add field-level validation for compliance-sensitive entries.
- Turn this deck into a one-page SOP checklist for the team.
- Connect an LLM provider to this agent for richer, audience-tuned summaries.`;
}
