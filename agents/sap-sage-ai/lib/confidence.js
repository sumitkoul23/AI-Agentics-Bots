/**
 * Confidence model.
 *
 * Confidence is NOT a vibe — it is computed deterministically from measurable
 * coverage so it can be trusted and tracked over time:
 *
 *   confidence = 0.45 * knowledgeCoverage   (how much SAP ECC ground we cover)
 *              + 0.40 * selfTestPassRate     (how well we answer our own Q&A)
 *              + 0.15 * trainingMaturity     (how many daily training cycles run)
 *
 * Each component is 0..1; the result is reported as a percentage with a band.
 */

// Reference targets for "full" coverage of an ECC generalist. Reaching these
// gives full marks on the coverage component; growing past them is capped at 1.
const TARGETS = { modules: 11, tcodes: 25, tables: 16, topics: 6, facts: 40 };

export function knowledgeCoverage(knowledge, mem) {
  const modules = knowledge["sap-ecc-model"]?.modules?.length || 0;
  const tcodes = knowledge["tcodes"]?.tcodes?.length || 0;
  const tables = knowledge["tables"]?.tables?.length || 0;
  const topics = Object.keys(knowledge).length;
  const facts = mem.learnedFacts.length;

  const parts = {
    modules: Math.min(1, modules / TARGETS.modules),
    tcodes: Math.min(1, tcodes / TARGETS.tcodes),
    tables: Math.min(1, tables / TARGETS.tables),
    topics: Math.min(1, topics / TARGETS.topics),
    facts: Math.min(1, facts / TARGETS.facts),
  };
  const score = (parts.modules + parts.tcodes + parts.tables + parts.topics + parts.facts) / 5;
  return { score, parts, counts: { modules, tcodes, tables, topics, facts } };
}

export function trainingMaturity(mem) {
  // Saturates around 30 daily cycles (~one month of learning).
  return Math.min(1, (mem.training?.runs || 0) / 30);
}

export function band(pct) {
  if (pct >= 85) return "Expert";
  if (pct >= 70) return "Proficient";
  if (pct >= 50) return "Competent";
  if (pct >= 30) return "Developing";
  return "Novice";
}

/**
 * @param {object} knowledge
 * @param {object} mem
 * @param {number} selfTestPassRate 0..1 (default 0 if not yet tested)
 */
export function computeConfidence(knowledge, mem, selfTestPassRate = 0) {
  const cov = knowledgeCoverage(knowledge, mem);
  const maturity = trainingMaturity(mem);
  const overall = 0.45 * cov.score + 0.4 * selfTestPassRate + 0.15 * maturity;
  const pct = Math.round(overall * 1000) / 10;
  return {
    overall: pct,
    band: band(pct),
    components: {
      knowledgeCoverage: Math.round(cov.score * 1000) / 10,
      selfTestPassRate: Math.round(selfTestPassRate * 1000) / 10,
      trainingMaturity: Math.round(maturity * 1000) / 10,
    },
    coverage: cov,
  };
}
