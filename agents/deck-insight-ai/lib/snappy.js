/**
 * Minimal pure-JS Snappy block decompressor.
 *
 * Apple Keynote (.key) stores its slide model as IWA files, whose chunk
 * payloads are Snappy-compressed protobuf. Node has no built-in Snappy, so we
 * implement just the decoder (no external dependency). Reference: the Snappy
 * format description — https://github.com/google/snappy/blob/main/format_description.txt
 */

/**
 * Decompress a single raw Snappy block.
 * @param {Buffer|Uint8Array} input
 * @returns {Buffer}
 */
export function snappyDecompress(input) {
  const src = Buffer.from(input);
  let pos = 0;

  // Preamble: uncompressed length as a varint.
  let outLen = 0;
  let shift = 0;
  while (pos < src.length) {
    const b = src[pos++];
    outLen |= (b & 0x7f) << shift;
    if ((b & 0x80) === 0) break;
    shift += 7;
  }

  const out = Buffer.allocUnsafe(outLen);
  let outPos = 0;

  while (pos < src.length) {
    const tag = src[pos++];
    const type = tag & 0x03;

    if (type === 0) {
      // Literal.
      let len = tag >> 2;
      if (len < 60) {
        len += 1;
      } else {
        const extraBytes = len - 59; // 1..4 bytes follow holding (length-1).
        let value = 0;
        for (let i = 0; i < extraBytes; i++) value |= src[pos++] << (8 * i);
        len = value + 1;
      }
      src.copy(out, outPos, pos, pos + len);
      pos += len;
      outPos += len;
    } else {
      // Copy.
      let length;
      let offset;
      if (type === 1) {
        length = 4 + ((tag >> 2) & 0x07);
        offset = ((tag >> 5) << 8) | src[pos++];
      } else if (type === 2) {
        length = 1 + (tag >> 2);
        offset = src[pos] | (src[pos + 1] << 8);
        pos += 2;
      } else {
        length = 1 + (tag >> 2);
        offset = src[pos] | (src[pos + 1] << 8) | (src[pos + 2] << 16) | (src[pos + 3] << 24);
        pos += 4;
      }
      let from = outPos - offset;
      for (let i = 0; i < length; i++) out[outPos++] = out[from++];
    }
  }

  return out.subarray(0, outPos);
}
