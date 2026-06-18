/**
 * Presentation extractor — dispatches by file type.
 *
 * Supported: .key (Apple Keynote), .pptx (PowerPoint). Returns a normalised
 * structure the analyser and LLM layers consume.
 */
import { readFileSync } from "node:fs";
import { extname, basename } from "node:path";
import { parseKeynote } from "./keynote.js";
import { parsePptx } from "./pptx.js";

/**
 * @param {string} path
 * @returns {{ source: string, format: string, slides: string[][], text: string, slideCount: number }}
 */
export function extractPresentation(path) {
  const buf = readFileSync(path);
  const ext = extname(path).toLowerCase();

  let parsed;
  let format;
  if (ext === ".key") {
    parsed = parseKeynote(buf);
    format = "keynote";
  } else if (ext === ".pptx") {
    parsed = parsePptx(buf);
    format = "pptx";
  } else {
    throw new Error(`Unsupported file type "${ext}". Use .key or .pptx`);
  }

  return {
    source: basename(path),
    format,
    slides: parsed.slides,
    text: parsed.text,
    slideCount: parsed.slides.length,
  };
}
