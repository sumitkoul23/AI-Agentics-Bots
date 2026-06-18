/**
 * Self-awareness protocol.
 *
 * Not sentience — a disciplined introspection report: the agent states who it
 * is, what it knows (inventory), how confident it is and why, what it does NOT
 * know (gaps), and how it keeps learning. Grounding self-reports in measurable
 * state is what keeps an autonomous agent honest.
 */
import { computeConfidence, knowledgeCoverage } from "./confidence.js";
import { selfTest } from "./learn.js";

export function selfReport(knowledge, mem) {
  const test = selfTest(knowledge, mem);
  const conf = computeConfidence(knowledge, mem, test.passRate);
  const cov = knowledgeCoverage(knowledge, mem);

  const modules = (knowledge["sap-ecc-model"]?.modules || []).map((m) => m.code);
  const gaps = [];
  if (cov.parts.facts < 1) gaps.push("Org-specific configuration beyond the seeded Customer Master SOP.");
  if (cov.parts.tcodes < 1) gaps.push("Long-tail / industry-specific T-codes.");
  if ((mem.training?.runs || 0) < 30) gaps.push(`Training maturity still building (${mem.training?.runs || 0}/30 daily cycles).`);
  gaps.push("Live system data — I reason about the SAP model, I do not connect to a live SAP instance.");

  return { test, conf, cov, modules, gaps };
}

export function renderSelfReport(knowledge, mem) {
  const { test, conf, cov, modules, gaps } = selfReport(knowledge, mem);
  const id = mem.identity;
  const last = mem.training?.lastTrained || "never";

  return `╔═══════════════════════════════════════════════════════════╗
║  ${id.name} — Self-Awareness Protocol
╚═══════════════════════════════════════════════════════════╝

IDENTITY
  • Name           : ${id.name}
  • Specialisation : ${id.specialisation}
  • Born           : ${id.born}
  • Autonomy       : NLP/AI agent — learns daily, remembers persistently.

CONFIDENCE  →  ${conf.overall}%  (${conf.band})
  • Knowledge coverage : ${conf.components.knowledgeCoverage}%
  • Self-test pass rate: ${conf.components.selfTestPassRate}%  (${test.pass}/${test.total} questions)
  • Training maturity  : ${conf.components.trainingMaturity}%

KNOWLEDGE INVENTORY (what I have memorised)
  • SAP ECC modules    : ${cov.counts.modules}  [${modules.join(", ")}]
  • T-codes            : ${cov.counts.tcodes}
  • Tables             : ${cov.counts.tables}
  • Knowledge topics   : ${cov.counts.topics}
  • Learned facts      : ${cov.counts.facts}
  • Ingested documents : ${mem.ingestedDocuments?.length || 0}

LEARNING STATE
  • Training cycles run: ${mem.training?.runs || 0}
  • Last trained       : ${last}

KNOWN GAPS (what I do NOT claim to know)
${gaps.map((g) => `  • ${g}`).join("\n")}
`;
}
