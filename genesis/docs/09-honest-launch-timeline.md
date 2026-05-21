# Honest launch timeline — what stands between this scaffold and mainnet

> Plain language. No persona, no "Genesis System." Just the truth about
> what it would take to launch AGENTIC as a real Cosmos SDK L1, written
> from the position of having shipped the scaffold in this repo.

## TL;DR

There are five honest scenarios. Pick one:

| Scenario | Realistic timeline to mainnet |
|---|---|
| Solo founder, learning Cosmos SDK as you go, no capital | **24–36 months** — and ~50 % of such projects never ship |
| Solo founder who is already an experienced Cosmos SDK engineer, no capital | **12–18 months** — testnet at month 6, mainnet at month 12+ |
| 2–3 experienced engineers + ~$500k for one audit and infra | **9–14 months** |
| Funded team (5+ engineers, $2–5M raised) | **6–9 months** to mainnet, plus the same number again to ecosystem maturity |
| Skip mainnet entirely; ship as a CosmWasm app on Neutron / Osmosis | **3–4 months** |

What this scaffold is worth on those timelines: it compresses the
**architecture + product design** phase by roughly **2–4 months**. The
remaining work is engineering, security, operations, and community.

## Where this scaffold actually puts us

To give credit where it's due, here is what's real and what's not.

### Real (≈ 8,000 lines of code + docs)

- Architecture decisions made and defensible (Cosmos SDK, PoS, 4 modules).
- Tokenomics encoded as concrete numbers (1B supply, 1–7 % inflation, 20 %
  burn per task, no VC allocation).
- Four module skeletons with the right shape — types, msgs, keepers,
  events. Anyone with Cosmos SDK experience can read this and understand
  the intent in 30 minutes.
- Unit tests for the two most important invariants (50/30/20 split math,
  AMM `x*y=k` conservation).
- A landing page that deploys today.
- A wallet brand surface that loads in Chrome today.
- A full strategy stack — DEX, CEX, growth, social, financial
  instruments, wallet. Eight documents that anyone could reuse.
- CI/CD wired up so the moment the chain compiles, every build is
  reproducible and signed.

### Not real (deliberately marked as such throughout the repo)

- **The chain does not compile end-to-end yet.** The hand-rolled `Msg`
  types in `types/msgs.go` files are placeholders; the keepers reference
  proto-generated types that don't exist. `go mod tidy` will fail.
- **No proto pipeline.** `buf generate` is not wired. The first real
  compile pass will reveal a dozen API mismatches against Cosmos SDK
  v0.50.
- **`app/app.go::New(...)` is incomplete.** It lists every module in
  `ModuleBasics` but the actual `runtime.AppBuilder` keeper-construction
  call sites have not been written. Adding them is ~200–400 LOC of
  careful copy-from-`simapp` work.
- **No genesis ceremony.** Real validator keys, real seed-node IDs, real
  `genesis.json` published — none of this has been done. Cannot be done
  remotely in a session.
- **No audit.** Cosmos SDK chains that launch without external review
  have a > 50 % rate of consensus-halting bugs in the first 90 days.
- **No external validators committed.** The "free-tier quartet" is a
  plan, not a registration list of actual operators.
- **No real social media presence.** No accounts have been claimed.
- **No real users.** Zero people have run an agent against this chain.

## The five gaps between scaffold and mainnet

### Gap 1 — Engineering (3–9 months depending on team)

What's needed:

1. **Proto pipeline.** Add `buf.work.yaml`, `buf.gen.yaml`, the protoc
   plugins, and the makefile target. Generates ~10k lines of typed Go.
2. **Module instance wiring.** Construct each keeper, hand it its
   dependencies (bank, account, staking), register the AppModule with
   the runtime composer. ~400 LOC across `app/app.go` and
   `app/modules.go` (the latter file doesn't exist yet).
3. **Real Msg server registration.** Today's keepers expose methods like
   `RegisterAgent` but they aren't connected to the SDK's tx router —
   that's a couple of hundred lines of boilerplate.
4. **gRPC query server.** Every Query handler in every module. Today
   these are mentioned in `module.go` files but not implemented.
5. **CLI commands.** `agenticd tx agentic register-agent` etc. — these
   are referenced in the docs but the cobra command graph isn't built.
6. **Integration tests.** A `simapp_test.go` per module that exercises
   the happy path end-to-end at the chain level. ~500 LOC per module.
7. **IBC integration.** Required for any cross-chain story. Adds the
   `x/ibc` modules, the transfer keeper, the channel keeper. ~200 LOC
   of wiring + a working light-client config.
8. **Real fraud-proof attestation store.** The current stub means we
   can only slash on quorum = 1; needs a multi-signer collections.Map.

Honest hours: **400–800 engineer-hours**. At one person full-time,
3–6 months. At three people, 6–10 weeks.

### Gap 2 — Security (2–4 months, $50–200k)

- Internal code review by someone *other* than the original author.
  Catches ~30 % of bugs.
- External audit. Realistic vendors at this scale: Informal Systems
  (Cosmos-native), Oak Security, Halborn, Trail of Bits, NCC Group.
  Expect **6–12 weeks turnaround**, **$80–200k**.
- Bug bounty program — Immunefi takes a percentage of payouts, so
  it's $0 upfront but commits the treasury.
- Penetration testing on the validator infra.

Cosmos SDK chains that skip the external audit have a > 50 % rate of
consensus-halting bugs in the first 90 days. This is the **single
non-negotiable line item**.

### Gap 3 — Operations (1–2 months)

- Validator-key ceremony — air-gapped machines, multi-party computation
  for genesis multisig keys, off-site recovery shards. Takes a long
  weekend if you've done it before; takes 3 weeks if you haven't.
- Actually provisioning the four validators on Oracle Cloud, Fly.io,
  GitHub Codespaces, AWS — and getting them in sync.
- Publishing the `genesis.json` to a CDN-fronted endpoint.
- Running a public devnet for at least 2 weeks to shake out config
  bugs.
- Running a public testnet for at least 4 weeks to validate the actual
  state machine under adversarial conditions.

### Gap 4 — Community + Validators (3–6 months)

- Claim the social handles (you, ~30 minutes from
  `growth/social/signup-checklist.md`).
- Post the build-in-public content (you, ongoing, 30 minutes / day).
- Engage in Cosmos Discord, forums, X — earn reputation in the
  ecosystem **we are about to ask to validate our chain**.
- Recruit 12–20 external validator operators willing to commit
  hardware for our launch.
- Get one or two "name" validators (Polychain, Imperator, Cosmostation)
  to commit — table stakes for credibility.
- Run the airdrop snapshot + claim flow before mainnet so day-1 has
  liquid stakers.

### Gap 5 — Legal (1–3 months, $20–80k)

- Entity formation if there's a foundation (BVI or Cayman, see
  `exchange/cex/jurisdictions.md`).
- Token-issuance memo from competent crypto counsel covering the
  airdrop + the contributor allocation.
- IP / contributor agreements with anyone who has contributed code,
  including AI-generated code — the legal status of which is
  *unsettled* in most jurisdictions as of 2026 and worth a 30-minute
  conversation with a lawyer before the genesis block.

## Three realistic scenarios — full breakdown

### A. Solo, $0, learning Cosmos SDK as you go

**Total time to mainnet: 24–36 months. ~50 % drop-off rate.**

- Months 1–6: learn Cosmos SDK from the official tutorials. Re-do every
  scaffold in this repo as a learning exercise.
- Months 6–12: get the chain compiling. Run a single-node devnet.
- Months 12–18: public testnet with a handful of validators.
- Months 18–24: audit (this is where the $50–80k bill arrives, since
  "no audit" is the line we won't cross). The audit usually surfaces
  ~10 high-severity issues; fixing them takes 6–8 weeks.
- Months 24–36: mainnet candidate, soft launch, mainnet.

Honest critique: the project is unlikely to survive 24+ months of solo
work because the market and the narrative will have moved. **This path is
not recommended.**

### B. You + 1–2 experienced Cosmos SDK engineers + ~$500k

**Total time to mainnet: 9–14 months.**

- Months 1–4: engineering catches up to the scaffold's intent.
  Everything compiles. Internal review.
- Months 5–6: public devnet.
- Months 6–8: audit + fix. ~$100–150k.
- Months 8–10: public testnet. Validator recruitment. Airdrop snapshot.
- Months 10–14: mainnet.

This is the realistic path if you have or can raise modest capital.

### C. Funded team — 5+ engineers, $2–5M raised

**Total time to mainnet: 6–9 months.**

The team can run engineering / audit / community in parallel rather
than serial. The audit is bigger and faster ($200k, accelerated). The
testnet has more validators and runs harder load. This is the path
every Cosmos chain you've heard of took.

### D. Skip mainnet — CosmWasm on Neutron or Osmosis

**Total time to live product: 3–4 months.**

The agent registry, the AMM, even the perps logic — all of it could
ship as CosmWasm contracts on Neutron (purpose-built smart-contract
chain) or Osmosis. Trade-offs:

- **Lose:** sovereign fee market, the ability to enforce protocol-level
  slashing, the "we have our own chain" narrative.
- **Win:** ~70 % less engineering, no validator recruitment, no
  consensus-halt risk, instant IBC connectivity, existing audited
  base layer.

Honest take: **this is the right choice for ~80 % of projects** that
start with "let's build a chain." It is worth seriously considering
whether AGENTIC needs to be sovereign. Re-read
[`docs/01-architecture.md`](01-architecture.md) §1 and ask "is the
slashable-stake-in-consensus argument actually load-bearing?" If the
honest answer is "we'd be fine with smart-contract slashing," scenario
D saves you a year.

## The minimum viable testnet — ordered checklist

Independent of which scenario, this is the *ordered* list of what has
to happen between today and the first public testnet block. Each item
unblocks the next; don't skip ahead.

1. [ ] Wire `buf.work.yaml` + `buf.gen.yaml`, generate proto for all
       four modules.
2. [ ] `go mod tidy` succeeds.
3. [ ] Hand-write `app/app.go::New(...)` keeper construction —
       reference Cosmos SDK's `simapp/app.go` line-by-line.
4. [ ] `make build` produces a working `agenticd` binary.
5. [ ] `./scripts/init-chain.sh` produces a valid `genesis.json`.
6. [ ] `./scripts/start-node.sh` produces blocks.
7. [ ] Wire each module's Msg server into the tx router.
8. [ ] `agenticd tx agentic register-agent` successfully registers an
       agent in a local devnet.
9. [ ] Repeat #8 for every Msg in every module — register, create,
       submit, settle, swap, open, liquidate, route.
10. [ ] Write integration tests at the chain level (`simapp_test.go`).
11. [ ] Internal code review by a different person.
12. [ ] Set up the seed node on Oracle Cloud Always-Free ARM.
13. [ ] Validator-key ceremony — generate genesis keys air-gapped,
        split with Shamir.
14. [ ] Publish `genesis.json` + RPC endpoint publicly.
15. [ ] Recruit ≥ 4 external validators willing to join testnet.
16. [ ] Run testnet for ≥ 4 weeks. Track downtime, slashing events,
        unexpected state-machine behaviour.
17. [ ] Engage audit firm with the testnet codebase.
18. [ ] Fix audit findings.
19. [ ] Bug bounty open.
20. [ ] Mainnet `genesis.json` ceremony.

Honest expectation: items 1–6 alone are **2–4 weeks** of focused work
even for someone experienced. Items 7–10 are another **3–6 weeks**.
Item 16 is **calendar-bound to 4 weeks minimum** no matter how good
the team. Item 17 is **calendar-bound to 6–12 weeks**.

## The single decision that defines the timeline

It's not technical. It's this:

> **Do we build a sovereign L1, or do we ship as a CosmWasm app on a
> chain that already exists?**

If sovereign L1 — accept the 9–14 month timeline minimum (with capital)
or 18–24 months (without). The reason is consensus-level slashing,
which is genuinely load-bearing for the agent-staking primitive if
we believe the slashable-bond argument is what makes AGENTIC defensible.

If CosmWasm on Neutron — accept the loss of the "our own chain"
narrative in exchange for 3–4 months to live product. The agent
registry, the DEX, even the perps module can all be contracts. The
slashing happens via contract logic rather than consensus, which is
strictly weaker but probably good enough for v0.

This is the conversation worth having before any more code gets
written. Until that decision is made explicit, every additional batch
of files is hedging both directions.

## What I can do from here, today, in a session

In rough order of leverage:

1. Write the `buf.work.yaml` + `buf.gen.yaml` configs. Doesn't run
   them — that needs the protoc binary on a machine — but the configs
   are correct.
2. Write the `simapp_test.go`-style integration test skeletons for
   each module. They won't compile yet, but they document the expected
   chain-level behaviours and become real the moment proto-gen runs.
3. Re-write `app/app.go::New(...)` against Cosmos SDK v0.50's
   `simapp/app.go` as a reference, even though the keeper constructors
   it calls don't yet exist. Closer to compilable than today.
4. Add a `Makefile` target that runs the whole pipeline locally on a
   developer's laptop.
5. Decision-doc: write a one-page brief weighing sovereign vs CosmWasm
   for your specific bet. You — not me — make the call.

What I cannot do in a session:

- Provision real validators.
- Generate real keys.
- Conduct an audit.
- Recruit external validators.
- Run a testnet for 4 weeks.
- Make the decision.

## Conclusion

The right next step is **not** to ship more files. It's to:

1. **Decide sovereign L1 or CosmWasm app.** (You.)
2. **Decide solo vs team vs funded.** (You.)
3. **Decide on your real timeline given the above.** (You.)
4. **Then I can help most by either:**
   - Filling in the proto pipeline + integration tests (path A),
   - Or rewriting the modules as CosmWasm contracts (path D).

I will not promise a launch date. Anyone who has shipped a Cosmos SDK
chain will tell you that promised-date schedules are how chains die
under technical debt or rushed audits. The realistic ranges in this
document are achievable; specific dates inside those ranges are not.
