/**
 * SKYMETRIC Wallet — REST + signing pipeline
 *
 * Talks to a Cosmos SDK REST endpoint (the LCD/`gaiad rest` API) to:
 *   - GET  /cosmos/auth/v1beta1/accounts/{address}   → account_number, sequence
 *   - GET  /cosmos/bank/v1beta1/balances/{address}   → balances
 *   - POST /cosmos/tx/v1beta1/txs                    → broadcast TxRaw
 *
 * Signing uses wallet-crypto.signBytes + wallet-proto encoders.
 */

import { signBytes, accountFromMnemonic, pubkeyFromPriv } from "/assets/js/wallet-crypto.js";
import { Proto, buildMsgSendAny, toBase64 } from "/assets/js/wallet-proto.js";

/* ── REST client ───────────────────────────────────────────────────── */
export class CosmosRest {
  constructor(restUrl) {
    this.url = restUrl.replace(/\/+$/, "");
  }

  async getAccount(address) {
    const r = await fetch(`${this.url}/cosmos/auth/v1beta1/accounts/${address}`);
    if (r.status === 404) return { accountNumber: 0, sequence: 0, exists: false };
    if (!r.ok) throw new Error(`account fetch failed (${r.status})`);
    const j = await r.json();
    const acc = j.account;
    return {
      accountNumber: BigInt(acc.account_number || 0),
      sequence:      BigInt(acc.sequence || 0),
      exists:        true,
    };
  }

  async getBalances(address) {
    const r = await fetch(`${this.url}/cosmos/bank/v1beta1/balances/${address}`);
    if (!r.ok) throw new Error(`balance fetch failed (${r.status})`);
    const j = await r.json();
    return j.balances || [];
  }

  async broadcast(txRawBytes, mode = "BROADCAST_MODE_SYNC") {
    const r = await fetch(`${this.url}/cosmos/tx/v1beta1/txs`, {
      method:  "POST",
      headers: { "Content-Type": "application/json" },
      body:    JSON.stringify({ tx_bytes: toBase64(txRawBytes), mode }),
    });
    const j = await r.json();
    if (!r.ok) throw new Error(j.message || `broadcast failed (${r.status})`);
    return j.tx_response || j;
  }
}

/* ── Signing pipeline ──────────────────────────────────────────────── */

/**
 * Build, sign, and serialize a single-message MsgSend transaction.
 *
 * Returns { txRawBytes, signDocBytes, bodyBytes, authInfoBytes }
 * so callers can broadcast and surface debug info to the UI.
 */
export async function signMsgSend({
  mnemonic,
  hrp = "agentic",
  path = "m/44'/118'/0'/0/0",
  toAddress,
  amount,        // [{ denom, amount }]
  memo = "",
  chainId,
  accountNumber, // bigint
  sequence,      // bigint
  fee,           // { amount: [{denom,amount}], gasLimit: bigint }
}) {
  const acct = await accountFromMnemonic(mnemonic, hrp, path);

  // 1. TxBody (MsgSend)
  const msgAny = buildMsgSendAny({ fromAddress: acct.address, toAddress, amount });
  const bodyBytes = Proto.encodeTxBody({ messages: [msgAny], memo });

  // 2. AuthInfo
  const authInfoBytes = Proto.encodeAuthInfo({
    signerInfo: { pubkey: acct.pubkey, sequence },
    fee,
  });

  // 3. SignDoc
  const signDocBytes = Proto.encodeSignDoc({
    bodyBytes, authInfoBytes, chainId, accountNumber,
  });

  // 4. Sign
  const signature = await signBytes(acct.priv, signDocBytes);

  // 5. TxRaw
  const txRawBytes = Proto.encodeTxRaw({
    bodyBytes, authInfoBytes, signatures: [signature],
  });

  return { txRawBytes, signDocBytes, bodyBytes, authInfoBytes, address: acct.address };
}

/* ── Convenience: full send ────────────────────────────────────────── */
export async function sendTokens({
  mnemonic,
  hrp,
  path,
  restUrl,
  chainId,
  toAddress,
  amount,
  memo,
  fee = { amount: [{ denom: "usky", amount: "1000" }], gasLimit: 200000n },
}) {
  const rest = new CosmosRest(restUrl);
  const acct = await accountFromMnemonic(mnemonic, hrp || "agentic", path || "m/44'/118'/0'/0/0");
  const accInfo = await rest.getAccount(acct.address);

  const { txRawBytes, address } = await signMsgSend({
    mnemonic, hrp, path,
    toAddress, amount, memo,
    chainId,
    accountNumber: accInfo.accountNumber,
    sequence:      accInfo.sequence,
    fee,
  });

  const result = await rest.broadcast(txRawBytes);
  return { result, address };
}
