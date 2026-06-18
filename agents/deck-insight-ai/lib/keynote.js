/**
 * Apple Keynote (.key) text extractor.
 *
 * A .key file is a ZIP of IWA archives (Index/*.iwa). Each IWA is a stream of
 * chunks: a 4-byte header (0x00 + 3-byte little-endian length) followed by a
 * Snappy-compressed protobuf payload. We decompress every chunk and harvest the
 * human-readable strings from the protobuf (TSWP text runs). We avoid a full
 * protobuf/TSP decoder — for summarisation we only need the words, not layout.
 */
import { readZip } from "./zip.js";
import { snappyDecompress } from "./snappy.js";

/** Decompress an IWA file into a single buffer of concatenated chunk payloads. */
function decompressIwa(buf) {
  const parts = [];
  let i = 0;
  while (i + 4 <= buf.length) {
    const length = buf[i + 1] | (buf[i + 2] << 8) | (buf[i + 3] << 16);
    i += 4;
    const block = buf.subarray(i, i + length);
    i += length;
    try {
      parts.push(snappyDecompress(block));
    } catch {
      /* skip undecodable chunk */
    }
  }
  return Buffer.concat(parts);
}

// Protobuf/TSP internal markers and obvious non-content tokens to drop.
const NOISE = [
  /^TS[A-Z]/, // TSWP, TSD, TSK, TSP, TSCH...
  /^com\.apple/,
  /^\.TSP/,
  /Placeholder/i,
  /^Picture \d/i,
  /^Title \d/i,
  /^Subtitle \d/i,
  /^Content \d/i,
  /^Transition$/i,
  /^none$/i,
  /^decimal$/i,
];

function isContent(s) {
  const letters = (s.match(/[A-Za-z]/g) || []).length;
  if (letters < 4) return false;
  return !NOISE.some((re) => re.test(s));
}

/** Clean a raw harvested string: strip the protobuf length-prefix bytes Keynote
 *  leaves around a text run. These render as a stray symbol — or, annoyingly, a
 *  stray letter/digit — glued to the start/end of the real words. */
function clean(s) {
  let out = s.replace(/^[^A-Za-z0-9*]+/, "").replace(/\*+$/, "").trim();
  // A single junk char glued before a capitalised word: "PGeneral" -> "General",
  // "kSales" -> "Sales", "mIn" -> "In", "6Sales" -> "Sales".
  out = out.replace(/^[A-Za-z0-9](?=[A-Z][a-z])/, "");
  // A single junk capital glued after an all-caps word: "CREATIONJ" -> "CREATION".
  out = out.replace(/([A-Z]{3,})[A-Z]$/, "$1");
  return out.trim();
}

/**
 * Extract slide text blocks from a Keynote buffer.
 * @param {Buffer} buf
 * @returns {{ slides: string[][], text: string }}
 */
export function parseKeynote(buf) {
  const zip = readZip(buf);
  const slideFiles = [...zip.keys()]
    .filter((name) => /^Index\/Slide.*\.iwa$/.test(name))
    .sort();

  const slides = [];
  const seenGlobal = new Set();

  for (const name of slideFiles) {
    const raw = decompressIwa(zip.get(name));
    const strings = raw.toString("latin1").match(/[\x20-\x7e]{4,}/g) || [];
    const block = [];
    const seenLocal = new Set();
    for (const s0 of strings) {
      const s = clean(s0);
      if (!s || !isContent(s)) continue;
      if (seenLocal.has(s)) continue;
      seenLocal.add(s);
      block.push(s);
    }
    if (block.length) slides.push(block);
    block.forEach((b) => seenGlobal.add(b));
  }

  const text = [...seenGlobal].join("\n");
  return { slides, text };
}
