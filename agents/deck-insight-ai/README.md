# Deck Insight AI

**NLP/AI agent that reads a presentation, understands its content, and summarises it in the voice of a Head of Department.**

Give it a `.key` (Apple Keynote) or `.pptx` (PowerPoint) file. It extracts the
text from every slide, works out the topic / steps / critical values, and writes
a structured leadership briefing — grounded strictly in what the deck actually
says.

```
deck.key / deck.pptx  →  [ parse ]  →  [ understand ]  →  [ summarise as persona ]  →  brief.md
```

---

## Quick start

No install needed — pure Node (v18+), **zero dependencies**.

```bash
cd agents/deck-insight-ai

# Head of Department briefing (default)
node main.js path/to/deck.key

# C-suite summary, written to a file
node main.js path/to/deck.pptx --persona exec --out brief.md

# Machine-readable output (summary + extracted slides)
node main.js path/to/deck.key --json
```

### Options

| Flag | Meaning |
|------|---------|
| `--persona <hod\|exec>` | Summary voice. `hod` = Head of Department (default), `exec` = C-suite. |
| `--out <path>` | Also write the Markdown briefing to a file. |
| `--json` | Emit a JSON envelope instead of Markdown. |
| `--help` | Show usage. |

---

## How it works

| Stage | File | Notes |
|-------|------|-------|
| ZIP reader | `lib/zip.js` | Reads `.pptx`/`.key` containers with Node's built-in `zlib` only. |
| Snappy decoder | `lib/snappy.js` | Pure-JS Snappy decompressor for Keynote IWA chunks. |
| Keynote parser | `lib/keynote.js` | Decompresses `Index/Slide*.iwa` and harvests the text runs. |
| PowerPoint parser | `lib/pptx.js` | Pulls `<a:t>` runs from `ppt/slides/slideN.xml`. |
| Persona prompts | `lib/persona.js` | Head-of-Department & Executive prompt templates. |
| Model client | `lib/llm.js` | Ollama → Groq → Gemini → Claude → OpenAI → offline. |
| CLI | `main.js` | Orchestration + flags. |

### Model selection (local-first)

Like the rest of the Bodhi swarm, it runs **free and local by default** and
upgrades when keys are present. It tries each provider in order and uses the
first that answers:

1. **Ollama** (local) — `OLLAMA_HOST` (default `http://localhost:11434`), `OLLAMA_MODEL` (default `llama3.2`)
2. **Groq** — `GROQ_API_KEY`
3. **Gemini** — `GEMINI_API_KEY`
4. **Anthropic Claude** — `ANTHROPIC_API_KEY`
5. **OpenAI** — `OPENAI_API_KEY`
6. **Offline extractive** — always available; no model, no network needed.

```bash
# Fully local
ollama serve & ollama pull llama3.2
node main.js deck.key

# Or use a hosted model
GROQ_API_KEY=... node main.js deck.key
```

---

## Example output

Running against an SAP **Customer Creation** SOP deck (T-Code `XD01`) produces a
Head-of-Department briefing that surfaces:

- the fixed keys — Account Group `Z002`, Company Code `2100`, Sales Org `Z210`, Reconciliation Account `11000000`;
- the regulated fields — GST number in *Tax Number 3*, Drug License in *Tax Number 5*, GST code `0`/`1`;
- the Sold-to / Ship-to partner codes `Z01` / `Z02`;
- and the controls a department head should put in place.

See [`samples/customer-creation-summary.md`](samples/customer-creation-summary.md)
for the LLM-written briefing, and
[`samples/customer-creation-summary.offline.md`](samples/customer-creation-summary.offline.md)
for the no-model extractive version.

---

## Supported formats

| Format | Extension | Status |
|--------|-----------|--------|
| Apple Keynote | `.key` | ✅ |
| Microsoft PowerPoint | `.pptx` | ✅ |
| PDF | `.pdf` | planned |
