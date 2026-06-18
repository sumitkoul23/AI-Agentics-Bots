/**
 * PowerPoint (.pptx) text extractor.
 *
 * A .pptx is an Open XML ZIP. Slide text lives in ppt/slides/slideN.xml inside
 * <a:t> runs. We read those parts in slide order and pull the text runs. No XML
 * library needed — the <a:t> elements are simple enough to match directly.
 */
import { readZip } from "./zip.js";

function decodeEntities(s) {
  return s
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&quot;/g, '"')
    .replace(/&apos;/g, "'");
}

/**
 * @param {Buffer} buf
 * @returns {{ slides: string[][], text: string }}
 */
export function parsePptx(buf) {
  const zip = readZip(buf);
  const slideNames = [...zip.keys()]
    .filter((n) => /^ppt\/slides\/slide\d+\.xml$/.test(n))
    .sort((a, b) => {
      const na = parseInt(a.match(/slide(\d+)\.xml$/)[1], 10);
      const nb = parseInt(b.match(/slide(\d+)\.xml$/)[1], 10);
      return na - nb;
    });

  const slides = [];
  for (const name of slideNames) {
    const xml = zip.get(name).toString("utf8");
    const runs = [...xml.matchAll(/<a:t>([\s\S]*?)<\/a:t>/g)]
      .map((m) => decodeEntities(m[1]).trim())
      .filter(Boolean);
    if (runs.length) slides.push(runs);
  }

  const text = slides.map((s) => s.join("\n")).join("\n");
  return { slides, text };
}
