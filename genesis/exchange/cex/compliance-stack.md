# Compliance vendor stack

Activated at Tier 4 only. All costs are approximate market rates as of
2025–2026; check during procurement.

## Identity verification (KYC)

| Vendor | What | Pricing | Notes |
|---|---|---|---|
| **Persona** | KYC + document verification | $1.00–2.50 / verification | Best DX, weak in APAC |
| **Sumsub** | Global KYC + AML | $0.80–1.80 / verification | Strong APAC + LATAM coverage |
| **Onfido** | KYC + biometric | $1.50–3.00 / verification | Highest EU acceptance rate |
| **Jumio** | Enterprise KYC | $2.00–4.00 / verification | Required for some banking partners |

**Default:** Persona for v0, Sumsub layered in by Y2 once Asian markets
open. Multi-vendor is unavoidable at scale — no single vendor satisfies
every jurisdiction's specific requirements.

## AML transaction monitoring

| Vendor | What | Pricing | Notes |
|---|---|---|---|
| **Chainalysis KYT** | Real-time transaction risk scoring | $200k–800k/yr enterprise | The default; required by most banking partners |
| **Elliptic** | Wallet screening + transaction monitoring | $150k–500k/yr | Strong in EU; better OFAC sanctions coverage |
| **TRM Labs** | Travel-rule + cross-chain monitoring | $100k–400k/yr | Best for Cosmos-native chains |
| **Merkle Science** | Behavioural analytics + threshold rules | $50k–200k/yr | Cheapest of the four; sufficient for early-stage |

**Default:** TRM Labs + Chainalysis dual-source. Bank counterparties
universally require Chainalysis on the AML pipeline; TRM gives us
Cosmos-native intelligence that Chainalysis underweights.

## Sanctions screening

| Vendor | What | Pricing |
|---|---|---|
| **ComplyAdvantage** | OFAC / EU / UK / UN sanctions watchlists | $30k–80k/yr |
| **Refinitiv World-Check** | Politically exposed persons (PEP) + sanctions | $80k–200k/yr |
| **Dow Jones Risk & Compliance** | Sanctions + adverse media | $60k–150k/yr |

**Default:** ComplyAdvantage at launch; add Refinitiv when institutional
desks come online (institutional clients often *require* World-Check).

## Travel rule (FATF Recommendation 16)

- **Sygna Bridge** (TRM-owned)
- **Notabene**
- **TRP / OpenVASP** (open standard, free reference impl)

We adopt the **TRP open standard** because it's free, interoperable, and
the future direction of the industry. Sygna remains a fallback for
counterparties stuck on proprietary endpoints.

## Audit + assurance

| Need | Provider | Notes |
|---|---|---|
| Financial audit | One of the Big Four (PwC, KPMG, EY, Deloitte) | Required for licensing reciprocity in EU + APAC |
| Proof-of-Reserves | Armanino historically; since they exited crypto, Mazars or BDO | Quarterly cadence |
| SOC 2 Type II | A-LIGN, Coalfire, or KirkpatrickPrice | Annual, ~$60–120k |
| Penetration testing | Trail of Bits, NCC Group, Cure53 | Quarterly + pre-launch full-stack |

## Operational tooling

| Tool | Purpose | Pricing |
|---|---|---|
| Hummingbot Enterprise | Market-making automation (when we run our own MM book) | OSS + paid support |
| Slack Enterprise Grid | Internal communications (subject to retention) | $25/user/mo |
| Egress / Material Security | Email DLP | $30/user/mo |
| 1Password Business / HashiCorp Vault | Secrets management | $19/user/mo |
| Atlassian Jira + Confluence | Compliance ticket + record retention | $7/user/mo |
| DocuSign + Ironclad | Contract management | $20–80/user/mo |

## What we don't pay for

- "Compliance-as-a-service" turnkey vendors. Every BSA officer we trust
  says the same thing: outsourced compliance fails the moment a regulator
  asks a clarifying question. We hire in-house from day 1.
- KYC vendor lock-in contracts > 12 months. Pricing and quality both
  change yearly — renegotiate every contract.
- Custodial insurance that excludes social engineering. Read every
  exclusion clause; the cheap policies are cheap because they exclude
  the 90 % of scenarios that have actually happened to exchanges.
