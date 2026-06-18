/**
 * Minimal ZIP reader using only Node's built-in zlib.
 *
 * Both .pptx (PowerPoint) and .key (Keynote) are ZIP containers. We parse the
 * central directory and inflate entries (deflate via zlib.inflateRawSync, or
 * "stored" verbatim). Enough to read presentation parts without any dependency.
 */
import zlib from "node:zlib";

const EOCD_SIG = 0x06054b50; // End Of Central Directory
const CDH_SIG = 0x02014b50; // Central Directory Header

/**
 * Read all entries of a ZIP buffer into a name -> Buffer map.
 * @param {Buffer} buf
 * @returns {Map<string, Buffer>}
 */
export function readZip(buf) {
  // Locate the End Of Central Directory record (scan backwards).
  let eocd = -1;
  for (let i = buf.length - 22; i >= 0; i--) {
    if (buf.readUInt32LE(i) === EOCD_SIG) {
      eocd = i;
      break;
    }
  }
  if (eocd < 0) throw new Error("Not a ZIP file (no EOCD record)");

  const count = buf.readUInt16LE(eocd + 10);
  let ptr = buf.readUInt32LE(eocd + 16); // offset of central directory

  const entries = new Map();
  for (let n = 0; n < count; n++) {
    if (buf.readUInt32LE(ptr) !== CDH_SIG) break;
    const method = buf.readUInt16LE(ptr + 10);
    const compSize = buf.readUInt32LE(ptr + 20);
    const nameLen = buf.readUInt16LE(ptr + 28);
    const extraLen = buf.readUInt16LE(ptr + 30);
    const commentLen = buf.readUInt16LE(ptr + 32);
    const localOff = buf.readUInt32LE(ptr + 42);
    const name = buf.toString("utf8", ptr + 46, ptr + 46 + nameLen);

    // Read the local header to find where the data actually starts.
    const lhNameLen = buf.readUInt16LE(localOff + 26);
    const lhExtraLen = buf.readUInt16LE(localOff + 28);
    const dataStart = localOff + 30 + lhNameLen + lhExtraLen;
    const raw = buf.subarray(dataStart, dataStart + compSize);

    let content;
    if (method === 0) content = Buffer.from(raw); // stored
    else if (method === 8) content = zlib.inflateRawSync(raw); // deflate
    else content = Buffer.alloc(0); // unsupported method — skip body

    entries.set(name, content);
    ptr += 46 + nameLen + extraLen + commentLen;
  }
  return entries;
}
