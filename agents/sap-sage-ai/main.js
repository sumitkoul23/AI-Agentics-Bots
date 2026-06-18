#!/usr/bin/env node
/**
 * SAP Sage AI — an autonomous NLP/AI agent specialising in SAP ECC.
 *
 * Commands:
 *   status                 Self-awareness protocol + confidence report
 *   ask "<question>"       Answer a SAP question accurately (with confidence)
 *   train                  Run one learning/training cycle (self-test + ingest)
 *   remember "<fact>"      Memorise a new fact permanently
 *   knowledge              Dump the knowledge inventory
 *
 * Model selection is automatic (Ollama → Groq → Gemini → Claude → offline).
 * The daily systemd timer runs `train` so the agent keeps learning every day.
 */
import { loadKnowledge, loadMemory, remember as rememberFact } from "./lib/memory.js";
import { renderSelfReport } from "./lib/awareness.js";
import { trainOnce } from "./lib/learn.js";
import { answer } from "./lib/answer.js";

function usage() {
  console.log(`SAP Sage AI — autonomous SAP ECC specialist

Usage:
  node main.js status
  node main.js ask "How do I view a customer in SAP?"
  node main.js train
  node main.js remember "Our plant for Mumbai DC is 2110"
  node main.js knowledge`);
}

async function main() {
  const [cmd, ...rest] = process.argv.slice(2);
  const arg = rest.join(" ").replace(/^["']|["']$/g, "");

  const knowledge = loadKnowledge();
  const mem = loadMemory();

  switch (cmd) {
    case "status":
    case "self":
    case "awareness": {
      process.stdout.write(renderSelfReport(knowledge, mem));
      break;
    }

    case "train": {
      console.error("▶ Running training cycle ...");
      const r = await trainOnce(knowledge, mem);
      console.log(`✔ Training cycle #${mem.training.runs} complete`);
      console.log(`  Self-test : ${r.test.pass}/${r.test.total} passed (${Math.round(r.test.passRate * 100)}%)`);
      console.log(`  Ingested  : ${r.ingested.length ? r.ingested.join(", ") : "nothing new in learn-inbox"}`);
      console.log(`  Confidence: ${r.confidence.overall}% (${r.confidence.band})`);
      break;
    }

    case "ask":
    case "q": {
      if (!arg) return usage();
      const res = await answer(knowledge, mem, arg);
      console.log(res.text);
      console.log(`\n— answer confidence: ${res.answerConfidence}% · model: ${res.provider}`);
      break;
    }

    case "remember": {
      if (!arg) return usage();
      rememberFact(mem, arg, "user");
      console.log(`✔ Memorised: "${arg}"  (total learned facts: ${mem.learnedFacts.length})`);
      break;
    }

    case "knowledge": {
      console.log(JSON.stringify({
        topics: Object.keys(knowledge),
        modules: (knowledge["sap-ecc-model"]?.modules || []).map((m) => m.code),
        tcodes: (knowledge["tcodes"]?.tcodes || []).length,
        tables: (knowledge["tables"]?.tables || []).length,
        learnedFacts: mem.learnedFacts.length,
      }, null, 2));
      break;
    }

    default:
      usage();
      process.exit(cmd ? 1 : 0);
  }
}

main().catch((e) => {
  console.error("✖ " + e.message);
  process.exit(1);
});
