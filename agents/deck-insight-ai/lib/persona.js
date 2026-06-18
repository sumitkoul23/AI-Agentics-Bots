/**
 * Persona prompt builders.
 *
 * The flagship persona is "Head of Department" — it reads the deck the way a
 * department head would brief their leadership: what this is, what matters, what
 * the risks are, and what to do next. Additional personas are easy to add.
 */

const PERSONAS = {
  hod: {
    label: "Head of Department",
    system:
      "You are an experienced Head of Department reviewing an internal " +
      "presentation. You think in terms of process ownership, operational " +
      "risk, compliance, and team enablement. You are concise, decisive, and " +
      "speak to leadership and to your team — never restating slides verbatim.",
    instructions: `Read the presentation content below and produce a Head-of-Department briefing in Markdown with EXACTLY these sections:

## Executive Summary
2-3 sentences: what this deck is, who it's for, and why it matters.

## What This Process Does
A short paragraph explaining the end-to-end process or topic in plain business language.

## Key Steps / Structure
A numbered list of the main steps or sections, in order, each one line.

## Critical Data & Configuration
A bullet list of every specific value, code, field, or setting mentioned (e.g. account groups, codes, fixed values). These are the things that must be entered correctly.

## Risks, Compliance & Watch-outs
Bullets on where mistakes are costly, compliance-sensitive fields, and anything ambiguous.

## Head of Department's Recommendations
3-5 action-oriented bullets: training, controls, validations, or follow-ups you would put in place.

Be specific and ground every point in the content provided. Do not invent values that are not present.`,
  },
  exec: {
    label: "Executive (C-suite)",
    system:
      "You are briefing a time-poor executive. You lead with outcomes and " +
      "decisions, strip jargon, and keep it to the essentials.",
    instructions:
      "Summarise the presentation for a C-suite reader in under 200 words: " +
      "the bottom line, 3 key points as bullets, and 1 recommended decision.",
  },
};

export function listPersonas() {
  return Object.entries(PERSONAS).map(([k, v]) => `${k} (${v.label})`);
}

/**
 * Build the messages for a chat-style LLM.
 * @param {string} personaKey
 * @param {{ source: string, slides: string[][] }} deck
 */
export function buildPrompt(personaKey, deck) {
  const persona = PERSONAS[personaKey] || PERSONAS.hod;
  const body = deck.slides
    .map((s, i) => `--- Slide ${i + 1} ---\n${s.join("\n")}`)
    .join("\n\n");

  const user = `Presentation file: ${deck.source}\nSlides: ${deck.slides.length}\n\n${persona.instructions}\n\n=== PRESENTATION CONTENT ===\n${body}`;

  return { system: persona.system, user, label: persona.label };
}

export { PERSONAS };
