/**
 * Learning & training.
 *
 * A training cycle (run daily by the systemd timer, or on demand):
 *   1. Self-test: answer the seed Q&A from retrieved knowledge and score it.
 *   2. Optionally ingest decks/notes dropped into ./learn-inbox (via the
 *      Deck Insight AI extractor if present) and remember new facts.
 *   3. Recompute confidence and append to the training history (so growth is
 *      visible day over day).
 *
 * This is honest "learning": the brain measurably broadens (more facts, more
 * coverage) and its self-test pass-rate is tracked — not a cosmetic counter.
 */
import { existsSync, readdirSync, renameSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join, extname } from "node:path";
import { retrieve } from "./answer.js";
import { computeConfidence } from "./confidence.js";
import { saveMemory, remember } from "./memory.js";

const __dirname = dirname(fileURLToPath(import.meta.url));
const INBOX = process.env.SAP_SAGE_INBOX || join(__dirname, "..", "learn-inbox");

/** Run the self-test: how many seed questions can we answer from our brain? */
export function selfTest(knowledge, mem) {
  const qa = knowledge["qa-seed"]?.qa || [];
  let pass = 0;
  const results = [];
  for (const item of qa) {
    const { top } = retrieve(knowledge, mem, item.q, 3);
    // Pass if the expected answer's keywords surface in the retrieved snippets.
    const hay = top.map((t) => t.text.toLowerCase()).join(" ");
    const kw = (item.keywords || []).map((k) => k.toLowerCase());
    const hit = kw.length ? kw.filter((k) => hay.includes(k)).length / kw.length : 0;
    const passed = hit >= 0.5;
    if (passed) pass++;
    results.push({ q: item.q, passed, hit: Math.round(hit * 100) });
  }
  return { total: qa.length, pass, passRate: qa.length ? pass / qa.length : 0, results };
}

/** Ingest any decks/notes in the learn-inbox, remembering one summary fact each. */
async function ingestInbox(mem) {
  if (!existsSync(INBOX)) return [];
  const ingested = [];
  let extractPresentation;
  try {
    ({ extractPresentation } = await import("../../deck-insight-ai/lib/extract.js"));
  } catch {
    /* Deck Insight AI not available — skip deck ingestion. */
  }
  for (const f of readdirSync(INBOX)) {
    const ext = extname(f).toLowerCase();
    const full = join(INBOX, f);
    if (extractPresentation && (ext === ".key" || ext === ".pptx")) {
      try {
        const deck = extractPresentation(full);
        const headline = deck.slides.flat().find((l) => l.length > 8) || f;
        remember(mem, `Ingested deck "${deck.source}" (${deck.slideCount} slides). Topic: ${headline}`, "deck");
        mem.ingestedDocuments.push({ source: deck.source, slides: deck.slideCount, ts: new Date().toISOString() });
        ingested.push(deck.source);
        renameSync(full, `${full}.learned`);
      } catch {
        /* ignore unreadable file */
      }
    }
  }
  return ingested;
}

/** Execute one full training cycle and persist results. */
export async function trainOnce(knowledge, mem) {
  const ingested = await ingestInbox(mem);
  const test = selfTest(knowledge, mem);
  const conf = computeConfidence(knowledge, mem, test.passRate);

  mem.training.runs += 1;
  mem.training.lastTrained = new Date().toISOString();
  mem.training.history.push({
    ts: mem.training.lastTrained,
    selfTestPassRate: Math.round(test.passRate * 1000) / 10,
    confidence: conf.overall,
    ingested,
  });
  // Keep history bounded.
  if (mem.training.history.length > 365) mem.training.history = mem.training.history.slice(-365);
  mem.confidenceHistory.push({ ts: mem.training.lastTrained, overall: conf.overall });
  if (mem.confidenceHistory.length > 365) mem.confidenceHistory = mem.confidenceHistory.slice(-365);
  saveMemory(mem);

  return { ingested, test, confidence: conf };
}

export { INBOX };
