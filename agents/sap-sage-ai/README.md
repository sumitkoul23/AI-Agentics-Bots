# SAP Sage AI

**An autonomous NLP/AI agent that specialises in SAP ECC — it understands the SAP ECC model, remembers everything it learns, trains itself every day, reports its own confidence, and answers accurately.**

```
knowledge/ (SAP ECC brain)  +  memory.json (everything learned)
        │
        ▼
  retrieve → answer (Ollama → cloud → offline)   ·   daily train → confidence ↑
```

---

## Quick start

Pure Node (v18+), **zero dependencies**.

```bash
cd agents/sap-sage-ai

node main.js status                                   # self-awareness + confidence
node main.js ask "How do I view a customer in SAP?"   # accurate answer + confidence
node main.js train                                     # one learning cycle
node main.js remember "Plant for Mumbai DC is 2110"   # memorise a fact forever
node main.js knowledge                                 # inventory dump
```

---

## What it knows (SAP ECC model)

- **Enterprise structure** — Client, Company Code, Controlling Area, Sales Area (Sales Org + Distribution Channel + Division), Plant, Storage Location, Purchasing Org, Shipping Point …
- **Modules** — FI, CO, SD, MM, PP, QM, PM, WM/EWM, LE, HCM, PS.
- **End-to-end processes** — Order-to-Cash, Procure-to-Pay, Record-to-Report, Plan-to-Produce.
- **Master vs transaction vs configuration data**, key **T-codes** (XD01/02/03, VA01, VL01N, VF01, ME21N, MIGO, MIRO, FS00, SPRO …) and **tables** (KNA1/KNB1/KNVV/KNVP, MARA, VBAK/VBAP, BKPF/BSEG …).
- The **Customer Master SOP** (XD01) seeded from the Keynote deck analysed by Deck Insight AI — including the org's fixed values and compliance rules.

Knowledge lives as editable JSON in [`knowledge/`](knowledge/). Add a file → it's instantly part of the brain and counts toward confidence.

---

## Self-awareness protocol

`node main.js status` prints an honest introspection report — identity, knowledge inventory, **confidence with rationale**, learning state, and **explicit known gaps** (e.g. "I reason about the SAP model; I do not connect to a live SAP instance"). This isn't sentience — it's disciplined self-reporting grounded in measurable state, which is what keeps an autonomous agent trustworthy.

## Confidence (deterministic, not a vibe)

```
confidence = 0.45 · knowledge coverage
           + 0.40 · self-test pass rate
           + 0.15 · training maturity
```

Bands: Novice < 30 ≤ Developing < 50 ≤ Competent < 70 ≤ Proficient < 85 ≤ Expert.
Out of the box it self-tests **12/12** and reports **~75 % (Proficient)**.

## Daily learning / training

Each training cycle (`node main.js train`):
1. **Self-test** against the seed Q&A and score the pass-rate.
2. **Ingest** any `.key`/`.pptx` dropped into `learn-inbox/` (via Deck Insight AI) and remember new facts.
3. **Recompute confidence** and append to the training history so growth is visible day over day.

Run it automatically every day with the included systemd units:

```bash
sudo cp deploy/linux/sap-sage-train.* /etc/systemd/system/
sudo systemctl enable --now sap-sage-train.timer    # trains daily at 02:00
```

## Memory ("memorise everything")

The durable brain is `memory.json` — learned facts, ingested documents, training and confidence history. Teach it anything with `remember`, and it persists across runs and reboots.

---

## Model strategy

Local-first, like the rest of the swarm: **Ollama → Groq → Gemini → Claude → offline knowledge**. With no model and no network it still answers accurately straight from its knowledge base.

| Env | Purpose |
|-----|---------|
| `OLLAMA_HOST`, `OLLAMA_MODEL` | local model (default `llama3.2`) |
| `GROQ_API_KEY` / `GEMINI_API_KEY` / `ANTHROPIC_API_KEY` | hosted fallbacks |
| `SAP_SAGE_MEMORY` | override memory file path |
| `SAP_SAGE_INBOX` | override learn-inbox path |
