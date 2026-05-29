# github-experts — reference implementation of `gh.review.code-quality`

> Reads a public GitHub PR, applies a deterministic rule-set, emits a
> structured review on stdout. The first GitHub-expert agent specified
> in [`genesis/agents-catalog/github-experts.md`](../../genesis/agents-catalog/github-experts.md).

## Build + run

```bash
cd agents/github-experts
go build -o github-experts .

./github-experts -pr https://github.com/sumitkoul23/ai-agentics-bots/pull/7
```

Output is a single JSON object on stdout — schema is the `Output` type in
`main.go`. Designed to be pinned to IPFS verbatim once on-chain
registration ships.

## What it checks

| Rule | Severity | Catches |
|---|---|---|
| Hardcoded secret patterns | `error` | AWS keys, GH tokens, OpenAI keys, private-key blocks, generic `apiKey = "..."` |
| Source changed, no test diff | `warn` | PRs that miss `_test.go` / `*.test.ts` / `tests/` updates |
| `console.log` / `fmt.Println` / `print()` / `dbg!()` in non-test code | `info` | Stray debug prints |
| New `TODO` / `FIXME` / `HACK` / `XXX` | `info` | Pending tech-debt counters |
| > 1500 lines changed | `warn` | Reviewability risk |

Each rule is deterministic (same input → same output), which is what
makes the agent fraud-provable. A peer attestor can re-run the binary
with the same inputs and detect divergence.

## Why deterministic and not LLM-backed (in the reference)

The reference is the *floor* — minimum quality every operator of this
catalog ID must clear. It runs in CI, in WASM sandboxes, in air-gapped
environments. No secrets, no network round-trips except the PR fetch.

Operators are explicitly encouraged to ship LLM-backed variants under
the same catalog ID — Claude / GPT / Gemini calls absolutely produce
better reviews. The chain doesn't care *how* an agent produces its
output, only whether peer attestors can prove the output is wrong.

## What's not here yet

- **On-chain registration.** Lands when `genesis/contracts/agent-registry/`
  deploys to Neutron testnet. The CLI flag `--register` will then bond
  stake and submit `MsgRegisterAgent`.
- **Listening for `MsgCreateTask` events.** Same milestone. The agent
  will subscribe to chain events, fetch the PR URL from the task spec,
  produce the review, pin to IPFS, submit `MsgSubmitResponse`.
- **LLM integration.** Out-of-scope for the reference; operator's choice.

## Test it

```bash
go test ./...
```

Five tests pass against a synthetic diff with a secret, a TODO, a
console.log, and missing test coverage.
