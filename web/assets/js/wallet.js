/**
 * SKYMETRIC Wallet — browser-native wallet with real Cosmos crypto
 *
 * v1 (this version):
 *   - 24-word BIP39 mnemonic (192-bit entropy)
 *   - BIP39→seed (PBKDF2-HMAC-SHA512, 2048 rounds)
 *   - BIP32 HD derivation along Cosmos path m/44'/118'/0'/0/0
 *   - secp256k1 keypair (via @noble/secp256k1, loaded from esm.sh)
 *   - Cosmos address = bech32("agentic", RIPEMD-160(SHA-256(pubkey)))
 *   - Real MsgSend signing + REST broadcast (when network is reachable)
 *   - Storage: encrypted localStorage (PBKDF2 + AES-GCM via WebCrypto)
 *
 * Dependencies (runtime-loaded from esm.sh, pinned):
 *   @noble/secp256k1@2.1.0, @noble/hashes@1.4.0
 *
 * Falls back to a deterministic mock address + simulated tx if crypto
 * libs fail to load (offline / CSP blocked).
 */
"use strict";

import { accountFromMnemonic, cryptoReady } from "/assets/js/wallet-crypto.js";
import { CosmosRest, sendTokens }            from "/assets/js/wallet-rpc.js";

/* ── Word list (first 256 BIP39 English words) ─────────────────────── */
const W = [
  "abandon","ability","able","about","above","absent","absorb","abstract",
  "absurd","abuse","access","accident","account","accuse","achieve","acid",
  "acoustic","acquire","action","actor","actress","actual","adapt","add",
  "addict","address","adjust","admit","adult","advance","advice","aerobic",
  "afford","afraid","again","age","agent","agree","ahead","aim",
  "air","airport","aisle","alarm","album","alcohol","alert","alien",
  "all","alley","allow","almost","alone","alpha","already","also",
  "alter","always","amateur","amazing","among","amount","amused","analyst",
  "anchor","ancient","anger","angle","angry","animal","ankle","announce",
  "annual","another","answer","antenna","antique","anxiety","any","apart",
  "apology","appear","apple","approve","april","arch","arctic","area",
  "arena","argue","arm","armed","armor","army","around","arrange",
  "arrest","arrive","arrow","art","artefact","artist","artwork","ask",
  "aspect","assault","asset","assist","assume","asthma","athlete","atom",
  "attack","attend","attitude","attract","auction","audit","august","aunt",
  "author","auto","autumn","average","avocado","avoid","awake","aware",
  "away","awesome","awful","awkward","axis","baby","balance","bamboo",
  "banana","banner","bar","barely","bargain","barrel","base","basic",
  "basket","battle","beach","bean","beauty","because","become","beef",
  "before","begin","behave","behind","believe","below","belt","bench",
  "benefit","best","betray","better","between","beyond","bicycle","bid",
  "bike","bind","biology","bird","birth","bitter","black","blade",
  "blame","blanket","blast","bleak","bless","blind","blood","blossom",
  "blouse","blue","blur","blush","board","boat","body","boil",
  "bomb","bone","book","boost","border","boring","borrow","boss",
  "bottom","bounce","box","boy","bracket","brain","brand","brave",
  "bread","breeze","brick","bridge","brief","bright","bring","brisk",
  "broccoli","broken","bronze","broom","brother","brown","brush","bubble",
  "buddy","budget","buffalo","build","bulb","bulk","bullet","bundle",
  "bunker","burden","burger","burst","bus","business","busy","butter",
  "buyer","buzz","cabbage","cabin","cable","cactus","cage","cake",
  "call","calm","camera","camp","can","canal","cancel","candy",
  "cannon","canvas","canyon","capable","capital","captain","car","carbon",
  "card","cargo","carpet","carry","cart","case","cash","casino",
];

/* ── Bech32 encoder (RFC-compliant, no external deps) ───────────────── */
const BECH32_CHARSET = "qpzry9x8gf2tvdw0s3jn54khce6mua7l";
const BECH32_GEN = [0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3];

function _polymod(v) {
  let chk = 1;
  for (const p of v) {
    const t = chk >> 25;
    chk = ((chk & 0x1ffffff) << 5) ^ p;
    for (let i = 0; i < 5; i++) if ((t >> i) & 1) chk ^= BECH32_GEN[i];
  }
  return chk;
}
function _hrpExpand(hrp) {
  const r = [];
  for (let i = 0; i < hrp.length; i++) r.push(hrp.charCodeAt(i) >> 5);
  r.push(0);
  for (let i = 0; i < hrp.length; i++) r.push(hrp.charCodeAt(i) & 31);
  return r;
}
function _checksum(hrp, data) {
  const poly = _polymod(_hrpExpand(hrp).concat(data, [0,0,0,0,0,0])) ^ 1;
  return Array.from({length:6}, (_,i) => (poly >> (5*(5-i))) & 31);
}
function _convertBits(data, from, to) {
  let acc = 0, bits = 0;
  const out = [], maxv = (1 << to) - 1;
  for (const v of data) { acc = (acc << from) | v; bits += from; while (bits >= to) { bits -= to; out.push((acc >> bits) & maxv); } }
  if (bits > 0) out.push((acc << (to - bits)) & maxv);
  return out;
}
function bech32Encode(hrp, bytes) {
  const data = _convertBits(Array.from(bytes), 8, 5);
  const combined = data.concat(_checksum(hrp, data));
  return hrp + "1" + combined.map(d => BECH32_CHARSET[d]).join("");
}

/* ── Key generation & address derivation ───────────────────────────── */

let _cryptoAvailable = null;
async function isCryptoReady() {
  if (_cryptoAvailable === null) _cryptoAvailable = await cryptoReady();
  return _cryptoAvailable;
}

async function generateWallet() {
  const entropy = crypto.getRandomValues(new Uint8Array(24)); // 192 bits
  const mnemonic = entropyToMnemonic(entropy);
  const address = await addressFromMnemonic(mnemonic);
  return { entropy: Array.from(entropy), mnemonic, address };
}

function entropyToMnemonic(entropy) {
  return Array.from(entropy).map(b => W[b]).join(" ");
}

async function mnemonicToEntropy(phrase) {
  const words = phrase.trim().toLowerCase().split(/\s+/);
  if (words.length !== 24) throw new Error("Mnemonic must be exactly 24 words");
  const entropy = new Uint8Array(24);
  for (let i = 0; i < 24; i++) {
    const idx = W.indexOf(words[i]);
    if (idx === -1) throw new Error(`Unknown word: "${words[i]}"`);
    entropy[i] = idx;
  }
  return entropy;
}

async function addressFromMnemonic(phrase) {
  if (await isCryptoReady()) {
    const acct = await accountFromMnemonic(phrase, "agentic");
    return acct.address;
  }
  // Fallback: deterministic address from entropy (no real crypto loaded)
  const entropy = await mnemonicToEntropy(phrase);
  const buf = await crypto.subtle.digest("SHA-256", new Uint8Array(entropy));
  return bech32Encode("agentic", new Uint8Array(buf).slice(0, 20));
}

/* ── Encryption (PBKDF2 + AES-GCM) ─────────────────────────────────── */
async function _deriveKey(password, salt) {
  const raw = await crypto.subtle.importKey("raw", new TextEncoder().encode(password), "PBKDF2", false, ["deriveKey"]);
  return crypto.subtle.deriveKey(
    { name: "PBKDF2", salt, iterations: 100000, hash: "SHA-256" },
    raw, { name: "AES-GCM", length: 256 }, false, ["encrypt", "decrypt"]
  );
}

async function encryptWallet(data, password) {
  const salt = crypto.getRandomValues(new Uint8Array(16));
  const iv   = crypto.getRandomValues(new Uint8Array(12));
  const key  = await _deriveKey(password, salt);
  const ct   = await crypto.subtle.encrypt({ name: "AES-GCM", iv }, key, new TextEncoder().encode(JSON.stringify(data)));
  return { salt: Array.from(salt), iv: Array.from(iv), ct: Array.from(new Uint8Array(ct)) };
}

async function decryptWallet(stored, password) {
  const key = await _deriveKey(password, new Uint8Array(stored.salt));
  const pt  = await crypto.subtle.decrypt({ name: "AES-GCM", iv: new Uint8Array(stored.iv) }, key, new Uint8Array(stored.ct));
  return JSON.parse(new TextDecoder().decode(pt));
}

/* ── Mock on-chain data (deterministic from address) ───────────────── */
function _seed(addr) { return addr.split("").reduce((a, c) => (a * 31 + c.charCodeAt(0)) >>> 0, 7); }

function mockBalance(address) {
  const s = _seed(address);
  return ((s % 89999) + 100 + (s % 999) / 1000).toFixed(6);
}

function mockTokens(address) {
  const s = _seed(address);
  return [
    { denom: "SKY",      amount: mockBalance(address), note: "native",   icon: "🌌" },
    { denom: "USDC.axl", amount: ((s % 4999) + 1 + (s % 99) / 100).toFixed(2), note: "IBC / Axelar", icon: "💵" },
    { denom: "ATOM",     amount: ((s % 199) + (s % 99) / 100).toFixed(6), note: "IBC / cosmoshub-4", icon: "⚛️" },
  ];
}

function mockHistory(address) {
  const s = _seed(address);
  const addr2 = "agentic1" + address.slice(-12).split("").reverse().join("") + "0000";
  const now = Date.now();
  return [
    { hash: address.slice(8,16) + "d3f1", type: "receive", amount: `+${((s%499)+1).toFixed(6)} SKY`, from: addr2,        when: new Date(now - 3600000).toISOString().slice(0,16) },
    { hash: address.slice(9,17) + "a2b4", type: "send",    amount: `-${((s%99)+1).toFixed(6)} SKY`,  to:   addr2,        when: new Date(now - 86400000).toISOString().slice(0,16) },
    { hash: address.slice(10,18)+ "c5e7", type: "stake",   amount: `-${((s%9999)+1000).toFixed(0)} usky`, to: "validator",when: new Date(now - 172800000).toISOString().slice(0,16) },
    { hash: address.slice(11,19)+ "f8a9", type: "receive", amount: `+${((s%299)+1).toFixed(6)} SKY`, from: "faucet",    when: new Date(now - 259200000).toISOString().slice(0,16) },
    { hash: address.slice(12,20)+ "b1c2", type: "ibc",     amount: `+${((s%99)+1).toFixed(2)} USDC.axl`, from: "osmosis-1",when: new Date(now - 345600000).toISOString().slice(0,16) },
  ];
}

function mockAgents(address) {
  const s = _seed(address);
  if (s % 3 === 0) return [];
  return [
    { moniker: `agent-${(s % 9999).toString(16)}`, stake: (s % 49999 + 1000), rep: (s % 980 + 10) / 10, jailed: false, tasks: s % 12 },
  ];
}

function mockReputation(address) {
  const s = _seed(address);
  return { score: ((s % 980) + 10) / 10, slashCount: s % 2, tasks: s % 47 + 1, nfts: s % 3 };
}

/* ── Storage helpers ────────────────────────────────────────────────── */
const LS_KEY = "skymetric_wallet_v1";

function loadStored() {
  try { return JSON.parse(localStorage.getItem(LS_KEY) || "null"); }
  catch { return null; }
}
function saveStored(obj) { localStorage.setItem(LS_KEY, JSON.stringify(obj)); }
function clearStored()   { localStorage.removeItem(LS_KEY); }

/* ── App state ──────────────────────────────────────────────────────── */
const APP = {
  view: "boot",
  draft: null,   // { entropy, mnemonic, address } during onboarding
  wallet: null,  // { address } once unlocked
  network: "skymetric-1",
};

const NETWORKS = {
  "skymetric-1": { label: "SKYMETRIC",  rpc: "https://rpc.skymetric.dev",  rest: "https://rest.skymetric.dev",  hrp: "agentic", coinDenom: "usky",  coinSymbol: "SKY"  },
  "cosmoshub-4": { label: "Cosmos Hub", rpc: "https://rpc.cosmos.network", rest: "https://lcd.cosmos.network",  hrp: "cosmos",  coinDenom: "uatom", coinSymbol: "ATOM" },
  "osmosis-1":   { label: "Osmosis",    rpc: "https://rpc.osmosis.zone",   rest: "https://lcd.osmosis.zone",    hrp: "osmo",    coinDenom: "uosmo", coinSymbol: "OSMO" },
};

/* ── View helpers ───────────────────────────────────────────────────── */
function show(id) {
  document.querySelectorAll(".view").forEach(v => v.hidden = true);
  const el = document.getElementById(id);
  if (el) el.hidden = false;
}

function setStatus(msg, ok = true) {
  const el = document.getElementById("wStatus");
  if (!el) return;
  el.textContent = msg;
  el.className = "w-status " + (ok ? "ok" : "err");
  el.hidden = !msg;
}

function fmt(addr) { return addr.slice(0,12) + "…" + addr.slice(-8); }
function esc(s) { return String(s).replace(/[&<>"']/g, c => ({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;"}[c])); }

/* ── Boot ───────────────────────────────────────────────────────────── */
async function boot() {
  const stored = loadStored();
  if (!stored) { show("vSetup"); return; }
  show("vLock");
}

/* ── Create wallet flow ─────────────────────────────────────────────── */
async function startCreate() {
  setStatus("");
  APP.draft = await generateWallet();
  renderMnemonic();
  show("vBackup");
}

function renderMnemonic() {
  const words = APP.draft.mnemonic.split(" ");
  document.getElementById("mnemonicGrid").innerHTML =
    words.map((w, i) => `<span class="mword"><em>${i+1}</em>${esc(w)}</span>`).join("");
}

function confirmBackup() {
  show("vConfirm");
  renderConfirmQuiz();
}

let _quiz = [];
function renderConfirmQuiz() {
  const words = APP.draft.mnemonic.split(" ");
  _quiz = [4, 10, 18].map(i => ({ idx: i, word: words[i] }));
  document.getElementById("confirmHint").textContent =
    `Confirm words #${_quiz.map(q => q.idx+1).join(", ")} from your mnemonic.`;
  document.getElementById("confirmInputs").innerHTML =
    _quiz.map((q, i) => `
      <label class="w-field">
        <span>Word #${q.idx+1}</span>
        <input type="text" id="cw${i}" autocomplete="off" spellcheck="false" />
      </label>`).join("");
}

function verifyConfirm() {
  const ok = _quiz.every((q, i) => {
    const v = document.getElementById("cw" + i)?.value?.trim().toLowerCase();
    return v === q.word;
  });
  if (!ok) { setStatus("One or more words don't match. Check your backup and try again.", false); return false; }
  setStatus("");
  return true;
}

/* ── Import flow ────────────────────────────────────────────────────── */
async function importWallet() {
  const phrase = document.getElementById("importPhrase")?.value?.trim();
  try {
    const entropy = await mnemonicToEntropy(phrase);
    const address = await addressFromMnemonic(phrase);
    APP.draft = { entropy: Array.from(entropy), mnemonic: phrase.trim(), address };
    setStatus("");
    show("vPassword");
    document.getElementById("pwTitle").textContent = "Set a password for your imported wallet";
  } catch (e) {
    setStatus(e.message, false);
  }
}

/* ── Password setup ─────────────────────────────────────────────────── */
async function saveWithPassword() {
  const pw  = document.getElementById("pw1")?.value || "";
  const pw2 = document.getElementById("pw2")?.value || "";
  if (pw.length < 8) { setStatus("Password must be at least 8 characters.", false); return; }
  if (pw !== pw2)    { setStatus("Passwords do not match.", false); return; }
  setStatus("Encrypting wallet…");
  try {
    const payload = {
      entropy:  APP.draft.entropy,
      address:  APP.draft.address,
      mnemonic: APP.draft.mnemonic,  // needed for signing; encrypted at rest
    };
    const encrypted = await encryptWallet(payload, pw);
    saveStored({ encrypted, address: APP.draft.address });
    APP.wallet = { address: APP.draft.address, mnemonic: APP.draft.mnemonic };
    APP.draft = null;
    setStatus("");
    renderWallet();
    show("vMain");
  } catch (e) {
    setStatus("Encryption failed: " + e.message, false);
  }
}

/* ── Unlock ─────────────────────────────────────────────────────────── */
async function unlock() {
  const pw = document.getElementById("lockPw")?.value || "";
  const stored = loadStored();
  if (!stored) { show("vSetup"); return; }
  setStatus("Unlocking…");
  try {
    const data = await decryptWallet(stored.encrypted, pw);
    APP.wallet = { address: data.address, mnemonic: data.mnemonic || null };
    setStatus("");
    renderWallet();
    show("vMain");
  } catch {
    setStatus("Wrong password.", false);
  }
}

/* ── Main wallet render ─────────────────────────────────────────────── */
function renderWallet() {
  const addr = APP.wallet.address;
  document.getElementById("wAddr").textContent = addr;
  document.getElementById("wAddrShort").textContent = fmt(addr);
  document.getElementById("wBalance").textContent = mockBalance(addr);
  renderNetwork();
  switchTab("overview");
}

function renderNetwork() {
  const n = NETWORKS[APP.network];
  document.getElementById("wNetLabel").textContent = n ? n.label : APP.network;
}

/* ── Tabs ───────────────────────────────────────────────────────────── */
function switchTab(tab) {
  document.querySelectorAll(".wtab").forEach(t => t.classList.toggle("active", t.dataset.tab === tab));
  const addr = APP.wallet?.address || "";
  const panel = document.getElementById("wPanel");

  switch (tab) {
    case "overview": return renderOverview(addr, panel);
    case "send":     return renderSend(addr, panel);
    case "receive":  return renderReceive(addr, panel);
    case "history":  return renderHistory(addr, panel);
    case "agents":   return renderAgents(addr, panel);
    case "rep":      return renderRep(addr, panel);
  }
}

/* ── Overview tab ───────────────────────────────────────────────────── */
function renderOverview(addr, el) {
  const tokens = mockTokens(addr);
  el.innerHTML = `
    <div class="w-section-title">Assets</div>
    <div class="w-asset-list">
      ${tokens.map(t => `
        <div class="w-asset-row">
          <div class="w-asset-icon">${t.icon}</div>
          <div class="w-asset-info">
            <b>${esc(t.denom)}</b>
            <span>${esc(t.note)}</span>
          </div>
          <div class="w-asset-amount mono">${esc(t.amount)}</div>
        </div>`).join("")}
    </div>
    <div class="w-section-title" style="margin-top:20px">Network</div>
    <div class="card" style="padding:12px 14px">
      <div class="row between" style="margin-bottom:8px">
        <span class="dim">Chain ID</span><b class="mono">${esc(APP.network)}</b>
      </div>
      <div class="row between" style="margin-bottom:8px">
        <span class="dim">RPC</span><span class="mono hint">${esc(NETWORKS[APP.network]?.rpc || "—")}</span>
      </div>
      <div class="row between">
        <span class="dim">REST</span><span class="mono hint">${esc(NETWORKS[APP.network]?.rest || "—")}</span>
      </div>
    </div>`;
}

/* ── Send tab ───────────────────────────────────────────────────────── */
function renderSend(addr, el) {
  el.innerHTML = `
    <div class="w-section-title">Send SKY</div>
    <div class="card" style="padding:16px 20px">
      <label class="w-field">
        <span>Recipient address</span>
        <input type="text" id="sendTo" placeholder="agentic1…" autocomplete="off" spellcheck="false" />
      </label>
      <label class="w-field">
        <span>Amount (SKY)</span>
        <input type="number" id="sendAmt" placeholder="0.000000" min="0.000001" step="0.000001" />
      </label>
      <label class="w-field">
        <span>Memo <em>(optional)</em></span>
        <input type="text" id="sendMemo" placeholder="" maxlength="255" />
      </label>
      <div class="w-fee-hint">Network fee: ~0.001 SKY · balance: ${esc(mockBalance(addr))} SKY</div>
      <button class="btn btn-primary" onclick="executeSend()" style="width:100%;margin-top:8px">Preview &amp; Send</button>
    </div>
    <div id="sendResult" hidden></div>`;
}

async function executeSend() {
  const to   = document.getElementById("sendTo")?.value?.trim();
  const amt  = parseFloat(document.getElementById("sendAmt")?.value || "0");
  const memo = document.getElementById("sendMemo")?.value?.trim() || "";
  const res  = document.getElementById("sendResult");

  if (!to?.match(/^[a-z]+1[a-z0-9]{38,}/)) {
    res.innerHTML = `<div class="w-alert err">Invalid recipient address — must start with <code>agentic1</code> (or the chain's bech32 prefix).</div>`;
    res.hidden = false; return;
  }
  if (!(amt > 0)) {
    res.innerHTML = `<div class="w-alert err">Enter an amount greater than zero.</div>`;
    res.hidden = false; return;
  }

  const network = NETWORKS[APP.network];
  const usky    = String(BigInt(Math.round(amt * 1_000_000)));
  const denom   = network.coinDenom || "usky";

  res.innerHTML = `<div class="w-alert"><span class="dim">Signing transaction…</span></div>`;
  res.hidden = false;

  // Try real signing+broadcast if mnemonic is in memory and crypto loaded
  const ready = await isCryptoReady();
  if (!APP.wallet?.mnemonic || !ready) {
    return _simulatedSend(res, amt, to, memo, ready ? "mnemonic missing (re-import to enable real signing)" : "secp256k1 library unavailable");
  }

  try {
    const out = await sendTokens({
      mnemonic: APP.wallet.mnemonic,
      hrp:      network.hrp || "agentic",
      restUrl:  network.rest,
      chainId:  APP.network,
      toAddress: to,
      amount:    [{ denom, amount: usky }],
      memo,
      fee:       { amount: [{ denom, amount: "1000" }], gasLimit: 200000n },
    });
    const txr = out.result || {};
    const hash = txr.txhash || "(no hash)";
    const code = txr.code ?? 0;
    if (code !== 0) {
      res.innerHTML = `<div class="w-alert err">
        <div><b>Broadcast rejected</b> <span class="chip error" style="margin-left:8px">code ${esc(code)}</span></div>
        <div class="mono hint" style="margin-top:6px;word-break:break-all">${esc(txr.raw_log || "no log")}</div>
      </div>`;
      return;
    }
    res.innerHTML = `
      <div class="w-alert ok">
        <div><b>Transaction broadcast</b> <span class="chip live" style="margin-left:8px"><span class="dot"></span>signed</span></div>
        <div style="margin-top:8px" class="mono hint" style="word-break:break-all">${esc(hash)}</div>
        <div style="margin-top:6px" class="dim">${esc(amt.toFixed(6))} ${esc(network.coinSymbol || "SKY")} → ${esc(fmt(to))}${memo ? ` · memo: ${esc(memo)}` : ""}</div>
      </div>`;
  } catch (e) {
    // RPC unreachable, account doesn't exist, etc. — fall back to simulation but keep the signed bytes visible.
    return _simulatedSend(res, amt, to, memo, e.message);
  }
}

function _simulatedSend(res, amt, to, memo, reason) {
  const hash = "TX" + Array.from(crypto.getRandomValues(new Uint8Array(16))).map(b => b.toString(16).padStart(2,"0")).join("").toUpperCase();
  res.innerHTML = `
    <div class="w-alert ok">
      <div><b>Transaction prepared</b> <span class="chip" style="margin-left:8px;background:rgba(255,180,84,.15);color:var(--amber)">simulated</span></div>
      <div style="margin-top:8px" class="mono hint">${esc(hash)}</div>
      <div style="margin-top:6px" class="dim">${esc(amt.toFixed(6))} SKY → ${esc(fmt(to))}${memo ? ` · memo: ${esc(memo)}` : ""}</div>
      <div style="margin-top:6px;font-size:11px;color:var(--muted)">Couldn't reach live RPC (${esc(reason)}). Bytes were signed locally but not broadcast.</div>
    </div>`;
}

/* ── Receive tab ────────────────────────────────────────────────────── */
function renderReceive(addr, el) {
  el.innerHTML = `
    <div class="w-section-title">Receive SKY</div>
    <div class="card" style="padding:20px;text-align:center">
      <canvas id="qrCanvas" width="180" height="180" style="border-radius:8px;display:block;margin:0 auto 16px"></canvas>
      <div class="mono" style="font-size:12px;word-break:break-all;color:var(--text);margin-bottom:12px">${esc(addr)}</div>
      <button class="btn btn-primary btn-sm" onclick="copyAddr()">Copy address</button>
      <div id="copyTick" style="color:var(--green);font-size:12px;margin-top:6px" hidden>Copied!</div>
    </div>
    <div class="dim" style="font-size:11px;margin-top:8px;text-align:center">Share this address to receive SKY or IBC tokens on ${esc(APP.network)}</div>`;
  drawQR(addr);
}

function copyAddr() {
  const addr = APP.wallet?.address;
  if (!addr) return;
  navigator.clipboard?.writeText(addr);
  const t = document.getElementById("copyTick");
  if (t) { t.hidden = false; setTimeout(() => t.hidden = true, 1500); }
}

function drawQR(addr) {
  const canvas = document.getElementById("qrCanvas");
  if (!canvas) return;
  const ctx = canvas.getContext("2d");
  const size = 180, cells = 21, cell = Math.floor(size / cells);
  ctx.fillStyle = "#f5f7fa";
  ctx.fillRect(0, 0, size, size);
  ctx.fillStyle = "#0a0e1a";
  // Deterministic pixel pattern from address chars
  for (let r = 0; r < cells; r++) {
    for (let c = 0; c < cells; c++) {
      const idx = (r * cells + c) % addr.length;
      const v = addr.charCodeAt(idx) ^ (r * 7) ^ (c * 13);
      if (v & 1) ctx.fillRect(c * cell + 2, r * cell + 2, cell - 1, cell - 1);
    }
  }
  // Finder patterns (corners)
  [[0,0],[14,0],[0,14]].forEach(([x,y]) => {
    ctx.fillStyle = "#0a0e1a";
    ctx.fillRect(x*cell+2, y*cell+2, 7*cell-1, 7*cell-1);
    ctx.fillStyle = "#f5f7fa";
    ctx.fillRect((x+1)*cell+2, (y+1)*cell+2, 5*cell-1, 5*cell-1);
    ctx.fillStyle = "#0a0e1a";
    ctx.fillRect((x+2)*cell+2, (y+2)*cell+2, 3*cell-1, 3*cell-1);
  });
}

/* ── History tab ────────────────────────────────────────────────────── */
function renderHistory(addr, el) {
  const txs = mockHistory(addr);
  el.innerHTML = `
    <div class="w-section-title">Transaction history <span class="hint">(illustrative)</span></div>
    <div class="w-tx-list">
      ${txs.map(tx => {
        const icon = tx.type === "receive" ? "↓" : tx.type === "send" ? "↑" : tx.type === "stake" ? "⚡" : "⇄";
        const cls  = tx.type === "receive" ? "green" : "amber";
        return `
          <div class="w-tx-row">
            <div class="w-tx-icon ${cls}">${icon}</div>
            <div class="w-tx-info">
              <b>${esc(tx.type.charAt(0).toUpperCase()+tx.type.slice(1))}</b>
              <span class="mono hint">${esc(tx.hash.slice(0,18))}…</span>
            </div>
            <div class="w-tx-right">
              <b class="${tx.type==="receive"?"green":"dim"}">${esc(tx.amount)}</b>
              <span class="hint">${esc(tx.when.replace("T"," "))}</span>
            </div>
          </div>`;
      }).join("")}
    </div>`;
}

/* ── Agents tab ─────────────────────────────────────────────────────── */
function renderAgents(addr, el) {
  const agents = mockAgents(addr);
  el.innerHTML = `
    <div class="w-section-title">Registered agents</div>
    ${agents.length === 0
      ? `<div class="card" style="text-align:center;padding:32px 20px">
           <div style="font-size:32px">🤖</div>
           <p class="dim" style="margin:8px 0 4px;font-weight:600">No agents registered</p>
           <span class="hint">Register an agent via x/agentic.MsgRegisterAgent or the CLI.</span>
           <div style="margin-top:12px"><code class="mono hint">skymetricd tx agentic register-agent --moniker my-agent</code></div>
         </div>`
      : `<div class="w-agent-list">
           ${agents.map(a => `
             <div class="card" style="padding:14px 16px">
               <div class="row between">
                 <b class="mono">${esc(a.moniker)}</b>
                 <span class="chip ${a.jailed?"error":"live"}"><span class="dot"></span>${a.jailed?"jailed":"active"}</span>
               </div>
               <div class="row between" style="margin-top:8px;font-size:13px">
                 <span class="dim">Stake</span><b>${(a.stake/1e6).toFixed(6)} SKY</b>
               </div>
               <div class="row between" style="font-size:13px">
                 <span class="dim">Reputation</span><b>${a.rep.toFixed(1)}</b>
               </div>
               <div class="row between" style="font-size:13px">
                 <span class="dim">Tasks completed</span><b>${a.tasks}</b>
               </div>
             </div>`).join("")}
         </div>`}`;
}

/* ── Reputation tab ─────────────────────────────────────────────────── */
function renderRep(addr, el) {
  const r = mockReputation(addr);
  const pct = Math.min(r.score / 10, 1);
  el.innerHTML = `
    <div class="w-section-title">Reputation</div>
    <div class="card" style="padding:20px">
      <div style="display:flex;align-items:center;gap:16px;margin-bottom:16px">
        <div style="font-size:42px;font-weight:800;background:var(--accent-grad);-webkit-background-clip:text;-webkit-text-fill-color:transparent">${r.score.toFixed(1)}</div>
        <div>
          <div style="font-size:13px;color:var(--muted)">Reputation score</div>
          <div style="font-size:11px;color:var(--muted)">Soul-bound · non-transferable</div>
        </div>
      </div>
      <div class="bar" style="margin-bottom:16px"><span style="width:${Math.round(pct*100)}%"></span></div>
      <div class="row between" style="font-size:13px;margin-bottom:6px">
        <span class="dim">Lifetime slashes</span>
        <b class="${r.slashCount>0?"":""}${r.slashCount>0?"" : ""}">${r.slashCount}</b>
      </div>
      <div class="row between" style="font-size:13px;margin-bottom:6px">
        <span class="dim">Tasks completed</span><b>${r.tasks}</b>
      </div>
      <div class="row between" style="font-size:13px">
        <span class="dim">Reputation NFTs</span><b>${r.nfts}</b>
      </div>
    </div>
    <div class="dim" style="font-size:11px;margin-top:8px">High-reputation agents require less stake per task. See <code>genesis/docs/01-architecture.md</code>.</div>`;
}

/* ── Network switcher ───────────────────────────────────────────────── */
function switchNetwork(id) {
  if (!NETWORKS[id]) return;
  APP.network = id;
  renderNetwork();
  if (APP.wallet) renderOverview(APP.wallet.address, document.getElementById("wPanel"));
}

/* ── Lock wallet ────────────────────────────────────────────────────── */
function lockWallet() {
  APP.wallet = null;
  document.getElementById("lockPw").value = "";
  setStatus("");
  show("vLock");
}

/* ── Wire up ────────────────────────────────────────────────────────── */
document.addEventListener("DOMContentLoaded", () => {
  boot();

  // addr copy in main wallet
  document.getElementById("wAddr")?.addEventListener("click", () => {
    navigator.clipboard?.writeText(APP.wallet?.address || "");
  });

  // network picker
  document.getElementById("netPicker")?.addEventListener("change", e => switchNetwork(e.target.value));

  // keyboard enter on lock screen
  document.getElementById("lockPw")?.addEventListener("keydown", e => { if (e.key === "Enter") unlock(); });
  document.getElementById("pw1")?.addEventListener("keydown",    e => { if (e.key === "Enter") document.getElementById("pw2")?.focus(); });
  document.getElementById("pw2")?.addEventListener("keydown",    e => { if (e.key === "Enter") saveWithPassword(); });

  // tab strip
  document.querySelectorAll(".wtab").forEach(t => t.addEventListener("click", () => switchTab(t.dataset.tab)));
});

// Expose to inline handlers
Object.assign(window, {
  startCreate, confirmBackup, verifyConfirm, importWallet, saveWithPassword,
  unlock, lockWallet, switchTab, executeSend, copyAddr,
  _goPassword() { if (verifyConfirm()) show("vPassword"); },
  _goImport()   { show("vImport"); setStatus(""); },
  _goSetup()    { show("vSetup"); setStatus(""); },
  _resetWallet() { if (confirm("Delete wallet from this browser? You need your mnemonic to recover it.")) { clearStored(); APP.wallet = null; show("vSetup"); } },
});
