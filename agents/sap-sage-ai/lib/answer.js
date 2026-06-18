/**
 * Question answering.
 *
 * 1. Retrieve relevant knowledge from the brain (keyword scoring over the seed
 *    knowledge, learned facts, and the self-test Q&A).
 * 2. If a model is reachable, ground it with that context and answer in prose.
 * 3. Otherwise, answer accurately straight from the retrieved knowledge.
 *
 * Every answer carries a per-answer confidence derived from how strongly the
 * question matched what the agent actually knows.
 */
import { chat } from "./llm.js";

const STOP = new Set(["the", "a", "an", "is", "are", "to", "of", "in", "on", "for", "how", "what", "do", "i", "and", "with", "using", "use", "can", "sap", "you"]);

function tokens(s) {
  return (s.toLowerCase().match(/[a-z0-9/]+/g) || []).filter((t) => t.length > 1 && !STOP.has(t));
}

/** Flatten the brain into searchable snippets. */
function snippets(knowledge, mem) {
  const out = [];
  const push = (text, ref) => text && out.push({ text: String(text), ref });

  const m = knowledge["sap-ecc-model"];
  if (m) {
    push(m.summary, "ECC model");
    push(`Successor: ${m.successor}`, "ECC model");
    (m.enterpriseStructure || []).forEach((u) => push(`${u.unit} (${u.module}): ${u.desc}`, "Enterprise structure"));
    (m.modules || []).forEach((x) => push(`${x.code} — ${x.name}: ${x.covers}`, "Modules"));
    (m.endToEndProcesses || []).forEach((p) => push(`${p.name} (${p.module}): ${p.flow}`, "Processes"));
  }
  (knowledge["tcodes"]?.tcodes || []).forEach((t) => push(`${t.code} = ${t.text} [${t.module}]${t.scope ? " — " + t.scope : ""}${t.note ? " — " + t.note : ""}`, "T-codes"));
  (knowledge["tables"]?.tables || []).forEach((t) => push(`${t.name}: ${t.text} [${t.module}]`, "Tables"));

  const cm = knowledge["customer-master"];
  if (cm) {
    push(`Customer master: create ${cm.createTcode}, change ${cm.changeTcode}, display ${cm.displayTcode}.`, "Customer master");
    push(`How to view a customer: ${cm.howToView}`, "Customer master");
    Object.entries(cm.fixedValues || {}).forEach(([k, v]) => push(`${k} = ${v}`, "Customer master fixed values"));
    (cm.complianceNotes || []).forEach((n) => push(n, "Customer master compliance"));
  }
  (knowledge["qa-seed"]?.qa || []).forEach((q) => push(`${q.q} ${q.a}`, "Q&A"));
  (mem.learnedFacts || []).forEach((f) => push(f.fact, `learned (${f.source})`));
  return out;
}

export function retrieve(knowledge, mem, question, k = 6) {
  const qt = new Set(tokens(question));
  const scored = snippets(knowledge, mem)
    .map((s) => {
      const st = tokens(s.text);
      let score = 0;
      for (const t of st) if (qt.has(t)) score++;
      return { ...s, score };
    })
    .filter((s) => s.score > 0)
    .sort((a, b) => b.score - a.score);
  const top = scored.slice(0, k);
  const maxPossible = Math.max(1, qt.size);
  const topScore = top[0]?.score || 0;
  const answerConfidence = Math.min(100, Math.round((Math.min(topScore, maxPossible) / maxPossible) * 100));
  return { top, answerConfidence };
}

/** Produce an accurate, grounded answer to a SAP question. */
export async function answer(knowledge, mem, question) {
  const { top, answerConfidence } = retrieve(knowledge, mem, question);
  const context = top.map((t) => `- (${t.ref}) ${t.text}`).join("\n");

  const system =
    "You are SAP Sage AI, an autonomous assistant specialising in SAP ECC. " +
    "Answer accurately and concisely using ONLY the provided knowledge context. " +
    "If the context does not cover it, say so plainly. Prefer exact T-codes, tables and values.";
  const user = `Question: ${question}\n\nKnowledge context:\n${context || "(no relevant knowledge found)"}`;

  const llm = await chat(system, user);
  if (llm.provider && llm.content) {
    return { provider: llm.provider, answerConfidence, text: llm.content, context: top };
  }
  // Offline: answer straight from retrieved knowledge.
  const text = top.length
    ? `Based on what I know:\n${top.map((t) => `• ${t.text}`).join("\n")}`
    : "I don't yet have knowledge covering that. Teach me with `remember`, or run a training cycle.";
  return { provider: "offline-knowledge", answerConfidence, text, context: top };
}
