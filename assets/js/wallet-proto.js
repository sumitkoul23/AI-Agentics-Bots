/**
 * SKYMETRIC Wallet — minimal Cosmos protobuf encoder
 *
 * Hand-rolled wire-format encoder for the subset of Cosmos SDK messages we need:
 *   - cosmos.base.v1beta1.Coin
 *   - cosmos.bank.v1beta1.MsgSend
 *   - cosmos.tx.v1beta1.TxBody, AuthInfo, SignerInfo, ModeInfo, Fee, SignDoc, TxRaw
 *   - cosmos.crypto.secp256k1.PubKey
 *   - google.protobuf.Any
 *
 * Wire format: tag = (field_number << 3) | wire_type
 *   wire types:  0 = varint, 1 = 64-bit, 2 = length-delimited, 5 = 32-bit
 *
 * Reference: https://protobuf.dev/programming-guides/encoding/
 */

/* ── Varint + tag helpers ──────────────────────────────────────────── */
function varint(n) {
  const out = [];
  let v = BigInt(n);
  while (v > 0x7fn) { out.push(Number(v & 0x7fn) | 0x80); v >>= 7n; }
  out.push(Number(v));
  return new Uint8Array(out);
}
function tag(field, wireType) { return varint((field << 3) | wireType); }

function lenDelim(field, bytes) {
  if (!bytes || bytes.length === 0) return new Uint8Array(0);
  return concat(tag(field, 2), varint(bytes.length), bytes);
}
function strField(field, str) {
  if (!str) return new Uint8Array(0);
  return lenDelim(field, new TextEncoder().encode(str));
}
function varintField(field, n) {
  if (n === undefined || n === null || n === 0 || n === 0n) return new Uint8Array(0);
  return concat(tag(field, 0), varint(n));
}
function concat(...arrs) {
  let total = 0;
  for (const a of arrs) total += a.length;
  const out = new Uint8Array(total);
  let off = 0;
  for (const a of arrs) { out.set(a, off); off += a.length; }
  return out;
}

/* ── Messages ──────────────────────────────────────────────────────── */

// cosmos.base.v1beta1.Coin { string denom = 1; string amount = 2; }
function encodeCoin({ denom, amount }) {
  return concat(strField(1, denom), strField(2, String(amount)));
}

// cosmos.bank.v1beta1.MsgSend { string from = 1; string to = 2; repeated Coin amount = 3; }
function encodeMsgSend({ fromAddress, toAddress, amount }) {
  const coins = amount.map(c => lenDelim(3, encodeCoin(c)));
  return concat(strField(1, fromAddress), strField(2, toAddress), ...coins);
}

// google.protobuf.Any { string type_url = 1; bytes value = 2; }
function encodeAny(typeUrl, value) {
  return concat(strField(1, typeUrl), lenDelim(2, value));
}

// cosmos.crypto.secp256k1.PubKey { bytes key = 1; }
function encodeSecp256k1PubKey(keyBytes) {
  return lenDelim(1, keyBytes);
}

// cosmos.tx.v1beta1.TxBody { repeated Any messages = 1; string memo = 2; ... }
function encodeTxBody({ messages, memo = "" }) {
  const msgAnys = messages.map(m => lenDelim(1, encodeAny(m.typeUrl, m.value)));
  return concat(...msgAnys, strField(2, memo));
}

// cosmos.tx.v1beta1.ModeInfo.Single { int32 mode = 1; }
//   SIGN_MODE_DIRECT = 1
function encodeModeInfoSingle(mode = 1) {
  return lenDelim(1, varintField(1, mode));
}
// cosmos.tx.v1beta1.ModeInfo { oneof sum { Single single = 1; ... } }
function encodeModeInfo(mode = 1) {
  return lenDelim(1, encodeModeInfoSingle(mode));
}

// cosmos.tx.v1beta1.SignerInfo { Any public_key = 1; ModeInfo mode_info = 2; uint64 sequence = 3; }
function encodeSignerInfo({ pubkey, sequence }) {
  const pubAny = encodeAny("/cosmos.crypto.secp256k1.PubKey", encodeSecp256k1PubKey(pubkey));
  return concat(
    lenDelim(1, pubAny),
    lenDelim(2, encodeModeInfo(1)),
    varintField(3, sequence),
  );
}

// cosmos.tx.v1beta1.Fee { repeated Coin amount = 1; uint64 gas_limit = 2; string payer = 3; string granter = 4; }
function encodeFee({ amount = [], gasLimit }) {
  const coins = amount.map(c => lenDelim(1, encodeCoin(c)));
  return concat(...coins, varintField(2, gasLimit));
}

// cosmos.tx.v1beta1.AuthInfo { repeated SignerInfo signer_infos = 1; Fee fee = 2; ... }
function encodeAuthInfo({ signerInfo, fee }) {
  return concat(
    lenDelim(1, encodeSignerInfo(signerInfo)),
    lenDelim(2, encodeFee(fee)),
  );
}

// cosmos.tx.v1beta1.SignDoc { bytes body_bytes = 1; bytes auth_info_bytes = 2; string chain_id = 3; uint64 account_number = 4; }
function encodeSignDoc({ bodyBytes, authInfoBytes, chainId, accountNumber }) {
  return concat(
    lenDelim(1, bodyBytes),
    lenDelim(2, authInfoBytes),
    strField(3, chainId),
    varintField(4, accountNumber),
  );
}

// cosmos.tx.v1beta1.TxRaw { bytes body_bytes = 1; bytes auth_info_bytes = 2; repeated bytes signatures = 3; }
function encodeTxRaw({ bodyBytes, authInfoBytes, signatures }) {
  const sigs = signatures.map(s => lenDelim(3, s));
  return concat(
    lenDelim(1, bodyBytes),
    lenDelim(2, authInfoBytes),
    ...sigs,
  );
}

/* ── Public API ────────────────────────────────────────────────────── */
export const Proto = {
  encodeCoin,
  encodeMsgSend,
  encodeAny,
  encodeTxBody,
  encodeSignerInfo,
  encodeFee,
  encodeAuthInfo,
  encodeSignDoc,
  encodeTxRaw,
  encodeSecp256k1PubKey,
};

/** Helper: build a full MsgSend Any { typeUrl, value } */
export function buildMsgSendAny({ fromAddress, toAddress, amount }) {
  return {
    typeUrl: "/cosmos.bank.v1beta1.MsgSend",
    value:   encodeMsgSend({ fromAddress, toAddress, amount }),
  };
}

/** Convert Uint8Array → base64 (browser-safe) */
export function toBase64(bytes) {
  let s = "";
  for (const b of bytes) s += String.fromCharCode(b);
  return btoa(s);
}
