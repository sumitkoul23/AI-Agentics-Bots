/**
 * SKYMETRIC Wallet Extension Popup — v1
 *
 * Key generation:  WebCrypto (crypto.getRandomValues + PBKDF2 + AES-GCM)
 * Mnemonic:        24 words from 256-word BIP39 subset (192-bit entropy)
 * Address:         bech32("agentic", SHA-256(entropy)[0:20])
 * Storage:         chrome.storage.local (AES-GCM encrypted)
 */

/* ── Word list ──────────────────────────────────────────────────────── */
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

/* ── Bech32 ─────────────────────────────────────────────────────────── */
const BC = "qpzry9x8gf2tvdw0s3jn54khce6mua7l";
const BG = [0x3b6a57b2,0x26508e6d,0x1ea119fa,0x3d4233dd,0x2a1462b3];
function _poly(v){let c=1;for(const p of v){const t=c>>25;c=((c&0x1ffffff)<<5)^p;for(let i=0;i<5;i++)if((t>>i)&1)c^=BG[i];}return c;}
function _expand(h){const r=[];for(let i=0;i<h.length;i++)r.push(h.charCodeAt(i)>>5);r.push(0);for(let i=0;i<h.length;i++)r.push(h.charCodeAt(i)&31);return r;}
function _chk(h,d){const p=_poly(_expand(h).concat(d,[0,0,0,0,0,0]))^1;return Array.from({length:6},(_,i)=>(p>>(5*(5-i)))&31);}
function _conv(data,f,t){let a=0,b=0;const o=[],m=(1<<t)-1;for(const v of data){a=(a<<f)|v;b+=f;while(b>=t){b-=t;o.push((a>>b)&m);}}if(b>0)o.push((a<<(t-b))&m);return o;}
function bech32(hrp,bytes){const d=_conv(Array.from(bytes),8,5);const c=d.concat(_chk(hrp,d));return hrp+"1"+c.map(x=>BC[x]).join("");}

/* ── Crypto helpers ─────────────────────────────────────────────────── */
async function genEntropy()  { return crypto.getRandomValues(new Uint8Array(24)); }
function  entropyToPhrase(e) { return Array.from(e).map(b=>W[b]).join(" "); }
async function phraseToEntropy(phrase) {
  const words = phrase.trim().toLowerCase().split(/\s+/);
  if (words.length !== 24) throw new Error("Need exactly 24 words");
  const e = new Uint8Array(24);
  for (let i=0;i<24;i++) { const j=W.indexOf(words[i]); if(j<0) throw new Error(`Unknown word: "${words[i]}"`); e[i]=j; }
  return e;
}
async function addressFromEntropy(e) {
  const h = await crypto.subtle.digest("SHA-256", e);
  return bech32("agentic", new Uint8Array(h).slice(0,20));
}
async function _deriveKey(pw, salt) {
  const raw = await crypto.subtle.importKey("raw",new TextEncoder().encode(pw),"PBKDF2",false,["deriveKey"]);
  return crypto.subtle.deriveKey({name:"PBKDF2",salt,iterations:100000,hash:"SHA-256"},raw,{name:"AES-GCM",length:256},false,["encrypt","decrypt"]);
}
async function encryptData(data, pw) {
  const salt=crypto.getRandomValues(new Uint8Array(16)), iv=crypto.getRandomValues(new Uint8Array(12));
  const key=await _deriveKey(pw,salt);
  const ct=await crypto.subtle.encrypt({name:"AES-GCM",iv},key,new TextEncoder().encode(JSON.stringify(data)));
  return {salt:Array.from(salt),iv:Array.from(iv),ct:Array.from(new Uint8Array(ct))};
}
async function decryptData(stored, pw) {
  const key=await _deriveKey(pw,new Uint8Array(stored.salt));
  const pt=await crypto.subtle.decrypt({name:"AES-GCM",iv:new Uint8Array(stored.iv)},key,new Uint8Array(stored.ct));
  return JSON.parse(new TextDecoder().decode(pt));
}

/* ── Storage ────────────────────────────────────────────────────────── */
function load() {
  return new Promise(res => chrome.storage.local.get("wallet", d => res(d.wallet || null)));
}
function save(obj) {
  return new Promise(res => chrome.storage.local.set({wallet: obj}, res));
}
function clear() {
  return new Promise(res => chrome.storage.local.remove("wallet", res));
}

/* ── Mock data ──────────────────────────────────────────────────────── */
function seed(addr) { return addr.split("").reduce((a,c)=>(a*31+c.charCodeAt(0))>>>0,7); }
function mockBal(addr) { const s=seed(addr); return ((s%89999)+100+(s%999)/1000).toFixed(6); }
function mockTokens(addr) {
  const s=seed(addr);
  return [
    {denom:"SKY",      amount:mockBal(addr),                         note:"native"},
    {denom:"USDC.axl", amount:((s%4999)+1+(s%99)/100).toFixed(2),   note:"IBC · Axelar"},
    {denom:"ATOM",     amount:((s%199)+(s%99)/100).toFixed(6),       note:"IBC · cosmoshub-4"},
  ];
}
function mockAgents(addr) {
  const s=seed(addr); if(s%3===0) return [];
  return [{moniker:`agent-${(s%9999).toString(16)}`, stake:(s%49999+1000), jailed:false, tasks:s%12}];
}
function mockRep(addr) { const s=seed(addr); return {score:((s%980)+10)/10, slashes:s%2, tasks:s%47+1}; }

/* ── App state ──────────────────────────────────────────────────────── */
const APP = { draft: null, wallet: null };
let _quiz = [];

/* ── Screen helpers ─────────────────────────────────────────────────── */
function showScreen(id) {
  document.querySelectorAll(".screen").forEach(s => s.hidden = true);
  document.getElementById(id).hidden = false;
}
function setErr(id, msg) {
  const el = document.getElementById(id);
  el.textContent = msg; el.hidden = !msg;
}
function fmt(addr) { return addr.slice(0,10)+"…"+addr.slice(-6); }
function esc(s) { return String(s).replace(/[&<>"']/g,c=>({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;"}[c])); }

/* ── Boot ───────────────────────────────────────────────────────────── */
async function boot() {
  const stored = await load();
  showScreen(stored ? "vLock" : "vSetup");
}

/* ── Create flow ────────────────────────────────────────────────────── */
document.getElementById("btnCreate").addEventListener("click", async () => {
  const entropy = await genEntropy();
  const mnemonic = entropyToPhrase(entropy);
  const address = await addressFromEntropy(entropy);
  APP.draft = { entropy: Array.from(entropy), mnemonic, address };
  renderMnemonicGrid();
  showScreen("vBackup");
});

function renderMnemonicGrid() {
  const words = APP.draft.mnemonic.split(" ");
  document.getElementById("mnemonicGrid").innerHTML =
    words.map((w,i) => `<span class="mword"><em>${i+1}</em>${esc(w)}</span>`).join("");
}

document.getElementById("btnConfirmBackup").addEventListener("click", () => {
  const words = APP.draft.mnemonic.split(" ");
  _quiz = [3,9,17].map(i => ({idx:i, word:words[i]}));
  document.getElementById("confirmHint").textContent =
    `Confirm words #${_quiz.map(q=>q.idx+1).join(", ")} from your phrase.`;
  document.getElementById("confirmInputs").innerHTML =
    _quiz.map((q,i) => `<label class="field" style="margin-bottom:8px">Word #${q.idx+1}<input type="text" id="cw${i}" autocomplete="off" spellcheck="false" /></label>`).join("");
  setErr("confirmErr","");
  showScreen("vConfirm");
});

document.getElementById("btnVerify").addEventListener("click", () => {
  const ok = _quiz.every((q,i) => (document.getElementById("cw"+i)?.value?.trim().toLowerCase()) === q.word);
  if (!ok) { setErr("confirmErr","Words don't match. Check your backup and try again."); return; }
  setErr("confirmErr","");
  document.getElementById("pw1").value = "";
  document.getElementById("pw2").value = "";
  document.getElementById("pwTitle")?.remove();
  setErr("pwErr","");
  showScreen("vPassword");
});

document.getElementById("backFromBackup").addEventListener("click", () => showScreen("vSetup"));
document.getElementById("backFromConfirm").addEventListener("click", () => showScreen("vBackup"));

/* ── Import flow ────────────────────────────────────────────────────── */
document.getElementById("btnImport").addEventListener("click", () => {
  document.getElementById("importPhrase").value = "";
  setErr("importErr","");
  showScreen("vImport");
});

document.getElementById("btnImportConfirm").addEventListener("click", async () => {
  const phrase = document.getElementById("importPhrase").value;
  try {
    const entropy = await phraseToEntropy(phrase);
    const address = await addressFromEntropy(entropy);
    APP.draft = { entropy: Array.from(entropy), mnemonic: phrase.trim(), address };
    setErr("importErr","");
    document.getElementById("pw1").value = "";
    document.getElementById("pw2").value = "";
    setErr("pwErr","");
    showScreen("vPassword");
  } catch(e) { setErr("importErr", e.message); }
});

document.getElementById("backFromImport").addEventListener("click", () => showScreen("vSetup"));

/* ── Password & save ────────────────────────────────────────────────── */
document.getElementById("btnSavePw").addEventListener("click", async () => {
  const pw  = document.getElementById("pw1").value;
  const pw2 = document.getElementById("pw2").value;
  if (pw.length < 8) { setErr("pwErr","Password must be at least 8 characters."); return; }
  if (pw !== pw2)    { setErr("pwErr","Passwords do not match."); return; }
  setErr("pwErr","");
  const encrypted = await encryptData({ entropy: APP.draft.entropy, address: APP.draft.address }, pw);
  await save({ encrypted, address: APP.draft.address });
  APP.wallet = { address: APP.draft.address };
  APP.draft = null;
  renderMain();
  showScreen("vMain");
});

/* ── Unlock ─────────────────────────────────────────────────────────── */
document.getElementById("btnUnlock").addEventListener("click", async () => {
  const pw = document.getElementById("lockPw").value;
  const stored = await load();
  if (!stored) { showScreen("vSetup"); return; }
  try {
    const data = await decryptData(stored.encrypted, pw);
    APP.wallet = { address: data.address };
    setErr("lockErr","");
    renderMain();
    showScreen("vMain");
  } catch { setErr("lockErr","Wrong password."); }
});
document.getElementById("lockPw").addEventListener("keydown", e => {
  if (e.key === "Enter") document.getElementById("btnUnlock").click();
});

document.getElementById("btnReset").addEventListener("click", async () => {
  if (!confirm("Delete wallet from this device? You need your phrase to recover it.")) return;
  await clear();
  APP.wallet = null;
  showScreen("vSetup");
});

/* ── Main wallet render ─────────────────────────────────────────────── */
function renderMain() {
  const addr = APP.wallet.address;
  document.getElementById("addr").textContent = fmt(addr);
  document.getElementById("addr").title = addr;
  document.getElementById("balance").textContent = mockBal(addr);
  renderTab("assets");
}

document.getElementById("addr").addEventListener("click", () => {
  const addr = APP.wallet?.address || "";
  navigator.clipboard?.writeText(addr).catch(()=>{});
});

document.querySelectorAll(".tab").forEach(t =>
  t.addEventListener("click", () => {
    document.querySelectorAll(".tab").forEach(x => x.classList.toggle("active", x===t));
    renderTab(t.dataset.tab);
  })
);

document.querySelectorAll(".actions button").forEach(b =>
  b.addEventListener("click", () => {
    if (b.dataset.action === "lock") { APP.wallet=null; showScreen("vLock"); return; }
    renderAction(b.dataset.action);
  })
);

function renderAction(action) {
  const panel = document.getElementById("panel");
  if (action === "send") {
    panel.innerHTML = `
      <h3>Send SKY</h3>
      <label style="display:block;font-size:11px;color:var(--ash);margin-bottom:4px">Recipient</label>
      <input type="text" id="sendTo" style="width:100%;background:rgba(8,10,22,.6);border:1px solid rgba(255,255,255,.08);color:var(--bone);border-radius:8px;padding:8px 10px;font:12px monospace;margin-bottom:8px;outline:none" placeholder="agentic1…" />
      <label style="display:block;font-size:11px;color:var(--ash);margin-bottom:4px">Amount (SKY)</label>
      <input type="number" id="sendAmt" style="width:100%;background:rgba(8,10,22,.6);border:1px solid rgba(255,255,255,.08);color:var(--bone);border-radius:8px;padding:8px 10px;font:12px inherit;margin-bottom:10px;outline:none" placeholder="0.000000" min="0.000001" step="0.000001" />
      <button id="doSend" style="width:100%;padding:9px;border-radius:8px;background:linear-gradient(135deg,#7c5cff,#18d3ff);color:#fff;border:0;font:600 13px inherit;cursor:pointer">Preview &amp; Send</button>
      <div id="sendRes" style="margin-top:8px;font-size:11px"></div>`;
    document.getElementById("doSend").addEventListener("click", () => {
      const to=document.getElementById("sendTo").value.trim(), amt=parseFloat(document.getElementById("sendAmt").value||"0");
      const res=document.getElementById("sendRes");
      if (!to.startsWith("agentic1") && !to.match(/^[a-z]+1[a-z0-9]{38}/)) { res.style.color="var(--slash)"; res.textContent="Invalid address."; return; }
      if (!(amt>0)) { res.style.color="var(--slash)"; res.textContent="Enter amount > 0."; return; }
      const hash="TX"+Array.from(crypto.getRandomValues(new Uint8Array(10))).map(b=>b.toString(16).padStart(2,"0")).join("").toUpperCase();
      res.style.color="var(--stake)"; res.innerHTML=`✓ Broadcast (simulated)<br><span style="font-family:monospace">${hash}</span>`;
    });
    return;
  }
  if (action === "receive") {
    const addr = APP.wallet?.address || "";
    panel.innerHTML = `
      <h3>Receive SKY</h3>
      <div style="background:rgba(255,255,255,.04);border:1px solid rgba(255,255,255,.08);border-radius:8px;padding:12px;word-break:break-all;font-family:monospace;font-size:10px;margin-bottom:8px">${esc(addr)}</div>
      <button id="copyAddr" style="width:100%;padding:8px;border-radius:8px;background:rgba(125,249,255,.1);border:1px solid rgba(125,249,255,.2);color:var(--gen);font:500 12px inherit;cursor:pointer">Copy address</button>
      <div id="copyTick" style="color:var(--stake);font-size:11px;margin-top:4px;text-align:center"></div>`;
    document.getElementById("copyAddr").addEventListener("click", () => {
      navigator.clipboard?.writeText(addr).catch(()=>{});
      document.getElementById("copyTick").textContent = "Copied!";
      setTimeout(()=>document.getElementById("copyTick").textContent="",1500);
    });
    return;
  }
  panel.innerHTML = `<div class="empty">"${esc(action)}" will be wired to x/agentic in the next release.</div>`;
}

/* ── Tab rendering ──────────────────────────────────────────────────── */
function renderTab(tab) {
  const addr = APP.wallet?.address || "";
  const panel = document.getElementById("panel");
  switch (tab) {
    case "assets":     return renderAssets(addr, panel);
    case "registry":   return renderRegistry(addr, panel);
    case "tasks":      return renderTasks(panel);
    case "reputation": return renderReputation(addr, panel);
    case "streams":    return renderStreams(panel);
  }
}

function renderAssets(addr, panel) {
  const tokens = mockTokens(addr);
  panel.innerHTML = `<h3>Assets</h3>` +
    tokens.map(t => `
      <div class="row">
        <div><div>${esc(t.denom)}</div><div class="muted mono">${esc(t.note)}</div></div>
        <div class="mono">${esc(t.amount)}</div>
      </div>`).join("");
}

function renderRegistry(addr, panel) {
  const agents = mockAgents(addr);
  if (!agents.length) {
    panel.innerHTML = `<h3>Your Agents</h3><div class="empty">No agents registered.<br><br>Use <span class="mono">skymetricd tx agentic register-agent</span> to create your first agent.</div>`;
    return;
  }
  panel.innerHTML = `<h3>Your Agents</h3>` +
    agents.map(a => `
      <div class="row">
        <div><div>${esc(a.moniker)}</div><div class="muted mono">stake ${(a.stake/1e6).toFixed(2)} SKY · ${a.tasks} tasks</div></div>
        <div class="pill ${a.jailed?"bad":"good"}">${a.jailed?"JAILED":"active"}</div>
      </div>`).join("");
}

function renderTasks(panel) {
  panel.innerHTML = `<h3>Open tasks</h3><div class="empty">No active tasks.<br>New tasks arrive via x/agentic.MsgCreateTask.</div>`;
}

function renderReputation(addr, panel) {
  const r = mockRep(addr);
  panel.innerHTML = `
    <h3>Reputation</h3>
    <div class="row"><div>Score</div><div class="mono">${r.score.toFixed(1)}</div></div>
    <div class="row"><div>Lifetime slashes</div><div class="mono ${r.slashes>0?"muted":""}">${r.slashes}</div></div>
    <div class="row"><div>Tasks completed</div><div class="mono">${r.tasks}</div></div>
    <div class="muted" style="margin-top:10px;font-size:11px">High reputation = less stake required per task. Soul-bound · non-transferable.</div>`;
}

function renderStreams(panel) {
  panel.innerHTML = `<h3>Streaming payments</h3><div class="empty">No active streams.<br>Streams land with the financial-instrument module (docs/06).</div>`;
}

/* ── Init ───────────────────────────────────────────────────────────── */
boot();
