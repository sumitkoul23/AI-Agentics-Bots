/**
 * SKYMETRIC Wallet — real cryptography module
 *
 * Loads audited primitives via pinned ESM imports:
 *   @noble/secp256k1  → ECDSA signing on the secp256k1 curve (Cosmos default)
 *   @noble/hashes      → SHA-256, SHA-512, RIPEMD-160, HMAC, PBKDF2
 *
 * This module REPLACES the entropy-only "address" used by the v0 wallet with:
 *   1. BIP39 mnemonic → seed (PBKDF2-HMAC-SHA512, 2048 rounds)
 *   2. BIP32 HD derivation along Cosmos path m/44'/118'/0'/0/0
 *   3. secp256k1 pubkey (compressed, 33 bytes)
 *   4. Cosmos address = bech32("agentic", RIPEMD-160(SHA-256(pubkey)))
 *   5. ECDSA signing for SignDoc bytes
 *
 * Protobuf encoding for MsgSend / TxBody / AuthInfo lives in wallet-proto.js.
 * Broadcast lives in wallet-rpc.js.
 *
 * All three modules are loaded from /assets/js/ alongside the existing wallet.js.
 */

// Pinned versions. Bump deliberately.
const NOBLE_SECP_URL    = "https://esm.sh/@noble/secp256k1@2.1.0";
const NOBLE_HASHES_BASE = "https://esm.sh/@noble/hashes@1.4.0";

let _secp, _sha256, _sha512, _ripemd160, _hmac, _pbkdf2;

async function _loadNoble() {
  if (_secp) return;
  const [secp, sha2, sha512mod, ripemd, hmacMod, pbkdf2Mod] = await Promise.all([
    import(NOBLE_SECP_URL),
    import(NOBLE_HASHES_BASE + "/sha256"),
    import(NOBLE_HASHES_BASE + "/sha512"),
    import(NOBLE_HASHES_BASE + "/ripemd160"),
    import(NOBLE_HASHES_BASE + "/hmac"),
    import(NOBLE_HASHES_BASE + "/pbkdf2"),
  ]);
  _secp      = secp;
  _sha256    = sha2.sha256;
  _sha512    = sha512mod.sha512;
  _ripemd160 = ripemd.ripemd160;
  _hmac      = hmacMod.hmac;
  _pbkdf2    = pbkdf2Mod.pbkdf2;
}

/* ── BIP39 mnemonic → seed ─────────────────────────────────────────── */
export async function mnemonicToSeed(mnemonic, passphrase = "") {
  await _loadNoble();
  const norm = mnemonic.normalize("NFKD");
  const salt = ("mnemonic" + passphrase).normalize("NFKD");
  return _pbkdf2(_sha512, new TextEncoder().encode(norm), new TextEncoder().encode(salt), { c: 2048, dkLen: 64 });
}

/* ── BIP32 HD derivation ───────────────────────────────────────────── */
const HARDENED = 0x80000000;

function _u32be(n) {
  const b = new Uint8Array(4);
  new DataView(b.buffer).setUint32(0, n >>> 0, false);
  return b;
}

async function _ckdPriv(parent, index) {
  const data = new Uint8Array(37);
  if (index >= HARDENED) {
    data[0] = 0x00;
    data.set(parent.key, 1);
  } else {
    // For non-hardened derivation we need the parent pubkey
    const pub = _secp.getPublicKey(parent.key, true);
    data.set(pub, 0);
  }
  data.set(_u32be(index), 33);
  const I = _hmac(_sha512, parent.chainCode, data);
  const IL = I.slice(0, 32), IR = I.slice(32);
  // child = (parent + IL) mod n
  const parentBI = BigInt("0x" + _hex(parent.key));
  const ILBI     = BigInt("0x" + _hex(IL));
  const N        = _secp.CURVE.n;
  const childBI  = ((parentBI + ILBI) % N);
  if (childBI === 0n) throw new Error("Invalid BIP32 child key");
  const child = _bytesFromBI(childBI, 32);
  return { key: child, chainCode: IR };
}

function _hex(b) { return Array.from(b).map(x => x.toString(16).padStart(2, "0")).join(""); }
function _bytesFromBI(bi, length) {
  const hex = bi.toString(16).padStart(length * 2, "0");
  const out = new Uint8Array(length);
  for (let i = 0; i < length; i++) out[i] = parseInt(hex.slice(i*2, i*2+2), 16);
  return out;
}

export async function derivePath(seed, path = "m/44'/118'/0'/0/0") {
  await _loadNoble();
  const I = _hmac(_sha512, new TextEncoder().encode("Bitcoin seed"), seed);
  let node = { key: I.slice(0, 32), chainCode: I.slice(32) };
  const segments = path.split("/").slice(1); // skip "m"
  for (const seg of segments) {
    const hardened = seg.endsWith("'");
    const idx = parseInt(hardened ? seg.slice(0, -1) : seg, 10);
    if (Number.isNaN(idx)) throw new Error(`Invalid path segment: ${seg}`);
    node = await _ckdPriv(node, hardened ? idx + HARDENED : idx);
  }
  return node.key;
}

/* ── Public key + address ──────────────────────────────────────────── */
export async function pubkeyFromPriv(priv) {
  await _loadNoble();
  return _secp.getPublicKey(priv, true);  // 33-byte compressed
}

export async function addressFromPubkey(pubkey, hrp = "agentic") {
  await _loadNoble();
  const sha   = _sha256(pubkey);
  const rip   = _ripemd160(sha);
  return _bech32Encode(hrp, rip);
}

/* ── Full account from mnemonic ────────────────────────────────────── */
export async function accountFromMnemonic(mnemonic, hrp = "agentic", path = "m/44'/118'/0'/0/0") {
  const seed   = await mnemonicToSeed(mnemonic);
  const priv   = await derivePath(seed, path);
  const pubkey = await pubkeyFromPriv(priv);
  const address = await addressFromPubkey(pubkey, hrp);
  return { priv, pubkey, address };
}

/* ── Signing ────────────────────────────────────────────────────────── */
export async function signBytes(priv, msgBytes) {
  await _loadNoble();
  const hash = _sha256(msgBytes);
  const sig = await _secp.signAsync(hash, priv, { lowS: true });
  // Cosmos expects 64-byte r||s
  const compact = sig.toCompactRawBytes ? sig.toCompactRawBytes() : new Uint8Array([...sig.r ? [] : []]);
  if (compact && compact.length === 64) return compact;
  // Fallback for older API
  const r = _bytesFromBI(sig.r, 32);
  const s = _bytesFromBI(sig.s, 32);
  const out = new Uint8Array(64);
  out.set(r, 0); out.set(s, 32);
  return out;
}

/* ── Bech32 (mirrors wallet.js but kept module-local for portability) ── */
const _BC = "qpzry9x8gf2tvdw0s3jn54khce6mua7l";
const _BG = [0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3];
function _poly(v) { let c=1; for(const p of v){const t=c>>25; c=((c&0x1ffffff)<<5)^p; for(let i=0;i<5;i++) if((t>>i)&1) c^=_BG[i];} return c; }
function _expand(h) { const r=[]; for(let i=0;i<h.length;i++) r.push(h.charCodeAt(i)>>5); r.push(0); for(let i=0;i<h.length;i++) r.push(h.charCodeAt(i)&31); return r; }
function _chk(h,d) { const p=_poly(_expand(h).concat(d,[0,0,0,0,0,0]))^1; return Array.from({length:6},(_,i)=>(p>>(5*(5-i)))&31); }
function _conv(data,from,to) { let a=0,b=0; const o=[],m=(1<<to)-1; for(const v of data){a=(a<<from)|v;b+=from;while(b>=to){b-=to;o.push((a>>b)&m);}} if(b>0) o.push((a<<(to-b))&m); return o; }
function _bech32Encode(hrp, bytes) {
  const d = _conv(Array.from(bytes), 8, 5);
  const c = d.concat(_chk(hrp, d));
  return hrp + "1" + c.map(x => _BC[x]).join("");
}
export { _bech32Encode as bech32Encode };

/* ── Health check (for the UI to surface "real crypto loaded") ─────── */
export async function cryptoReady() {
  try { await _loadNoble(); return true; } catch (e) { console.error("crypto load failed", e); return false; }
}
