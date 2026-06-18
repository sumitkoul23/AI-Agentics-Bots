/**
 * Persistent memory for SAP Sage AI.
 *
 * "Memorise everything": the agent keeps a durable JSON brain on disk. On every
 * run it loads the seed knowledge (knowledge/*.json) plus anything it has learned
 * since (facts the user taught it, decks it ingested, training history).
 */
import { readFileSync, writeFileSync, existsSync, readdirSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

const __dirname = dirname(fileURLToPath(import.meta.url));
const ROOT = join(__dirname, "..");
const KNOWLEDGE_DIR = join(ROOT, "knowledge");
const MEMORY_FILE = process.env.SAP_SAGE_MEMORY || join(ROOT, "memory.json");

const EMPTY = {
  identity: {
    name: "SAP Sage AI",
    specialisation: "SAP ECC (ERP Central Component)",
    born: null,
    version: 1,
  },
  learnedFacts: [], // [{ fact, source, ts }]
  ingestedDocuments: [], // [{ source, topics, ts }]
  training: { runs: 0, lastTrained: null, history: [] }, // history: [{ ts, selfTestPassRate, confidence }]
  confidenceHistory: [], // [{ ts, overall }]
};

/** Load all seed knowledge files into one map keyed by file name. */
export function loadKnowledge() {
  const out = {};
  for (const f of readdirSync(KNOWLEDGE_DIR)) {
    if (!f.endsWith(".json")) continue;
    out[f.replace(/\.json$/, "")] = JSON.parse(readFileSync(join(KNOWLEDGE_DIR, f), "utf8"));
  }
  return out;
}

/** Load the durable memory (creating it on first run). */
export function loadMemory() {
  if (!existsSync(MEMORY_FILE)) {
    const fresh = structuredClone(EMPTY);
    fresh.identity.born = new Date().toISOString();
    writeFileSync(MEMORY_FILE, JSON.stringify(fresh, null, 2));
    return fresh;
  }
  const mem = JSON.parse(readFileSync(MEMORY_FILE, "utf8"));
  // Forward-compatible defaults.
  return { ...structuredClone(EMPTY), ...mem, identity: { ...EMPTY.identity, ...mem.identity } };
}

export function saveMemory(mem) {
  writeFileSync(MEMORY_FILE, JSON.stringify(mem, null, 2));
}

/** Teach the agent a new fact (it is remembered forever). */
export function remember(mem, fact, source = "user") {
  mem.learnedFacts.push({ fact, source, ts: new Date().toISOString() });
  saveMemory(mem);
  return mem;
}

export { MEMORY_FILE };
