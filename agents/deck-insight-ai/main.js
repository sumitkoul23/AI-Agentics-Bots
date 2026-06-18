#!/usr/bin/env node
/**
 * Deck Insight AI — NLP/AI presentation analyst.
 *
 * Reads a .key or .pptx deck, understands its content, and writes a summary in
 * the voice of a chosen persona (default: Head of Department). Local-first via
 * Ollama, with cloud fallbacks and an offline extractive mode so it always
 * produces output.
 *
 * Usage:
 *   node main.js <file.key|.pptx> [--persona hod|exec] [--out summary.md] [--json]
 *
 * Examples:
 *   node main.js deck.key
 *   node main.js deck.pptx --persona exec --out brief.md
 */
import { writeFileSync } from "node:fs";
import { extractPresentation } from "./lib/extract.js";
import { buildPrompt, listPersonas } from "./lib/persona.js";
import { summarise } from "./lib/llm.js";

function parseArgs(argv) {
  const args = { persona: "hod", out: null, json: false, file: null };
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i];
    if (a === "--persona") args.persona = argv[++i];
    else if (a === "--out") args.out = argv[++i];
    else if (a === "--json") args.json = true;
    else if (a === "--help" || a === "-h") args.help = true;
    else if (!a.startsWith("--")) args.file = a;
  }
  return args;
}

function help() {
  console.log(`Deck Insight AI — analyse & summarise a presentation

Usage:
  node main.js <file.key|.pptx> [options]

Options:
  --persona <key>   Summary voice. One of: ${listPersonas().join(", ")}
                    Default: hod (Head of Department)
  --out <path>      Write the summary to a file (Markdown)
  --json            Emit a JSON envelope (summary + extracted slides)
  --help            Show this help

Model selection is automatic: Ollama (local) → Groq → Gemini → Claude → OpenAI
→ offline extractive. Configure via env (OLLAMA_HOST, OLLAMA_MODEL,
GROQ_API_KEY, GEMINI_API_KEY, ANTHROPIC_API_KEY, OPENAI_API_KEY).`);
}

async function main() {
  const args = parseArgs(process.argv.slice(2));
  if (args.help || !args.file) {
    help();
    process.exit(args.file ? 0 : 1);
  }

  console.error(`▶ Reading ${args.file} ...`);
  const deck = extractPresentation(args.file);
  console.error(`  extracted ${deck.slideCount} slides (${deck.format})`);

  if (deck.slideCount === 0) {
    console.error("  ⚠ No text could be extracted from this file.");
  }

  const { system, user, label } = buildPrompt(args.persona, deck);
  console.error(`▶ Summarising as: ${label} ...`);
  const result = await summarise(system, user, deck);
  console.error(`  model: ${result.provider}`);

  const header = `<!-- Deck Insight AI · persona: ${label} · model: ${result.provider} · source: ${deck.source} -->\n\n`;
  const output = header + result.content + "\n";

  if (args.json) {
    process.stdout.write(
      JSON.stringify(
        {
          source: deck.source,
          format: deck.format,
          slideCount: deck.slideCount,
          persona: label,
          provider: result.provider,
          summary: result.content,
          slides: deck.slides,
        },
        null,
        2
      ) + "\n"
    );
  } else {
    process.stdout.write("\n" + output);
  }

  if (args.out) {
    writeFileSync(args.out, output);
    console.error(`✔ Summary written to ${args.out}`);
  }
}

main().catch((e) => {
  console.error("✖ " + e.message);
  process.exit(1);
});
