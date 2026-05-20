# 30-minute signup checklist

> Tick these off in order. Direct URLs included. Use the same email
> (`agentic.chain@protonmail.com` or similar — set up a dedicated alias **first**)
> for every platform so password resets all land in one inbox.

## Pre-flight (one-time, 3 min)

- [ ] Create a dedicated email alias: e.g. ProtonMail or `+agentic` Gmail alias.
      Every account below uses this email.
- [ ] Save the recovery codes / 2FA seeds in a password manager
      (Bitwarden is free).
- [ ] Open [`handles.md`](handles.md) in another tab so you can pick the
      rank-1 handle and fall back to rank-2 if taken.
- [ ] Open [`bios.md`](bios.md) — paste the matching bio when prompted.
- [ ] Have [`logo.svg`](logo.svg) exported to PNG at the sizes in
      [`avatar-banner-spec.md`](avatar-banner-spec.md) — or just upload the
      SVG and let each platform downscale (works everywhere except
      Telegram and Discord, which need PNG).

## Step 1 · X / Twitter (4 min)

1. Sign up: https://x.com/i/flow/signup
2. Try handles in order from `handles.md` → X section.
3. Display name: `AGENTIC`
4. Bio: paste the "Primary (158 chars)" from `bios.md`.
5. Profile pic: `logo.png` (400×400).
6. Banner: `banner-x.png` (1500×500).
7. Pin a tweet: paste the "Pinned-tweet template (280 chars)" from `bios.md`.
8. Enable 2FA → settings → security → 2FA (authenticator app, **not** SMS).

## Step 2 · GitHub organisation (2 min)

1. Create the org: https://github.com/account/organizations/new (Free plan).
2. Name: try `agentic-chain` → fallbacks in `handles.md` → GitHub section.
3. Upload `logo.png` (500×500) as the org avatar.
4. Add the LinkedIn-variant bio (from `bios.md`) as the org description.
5. Org README: create a public repo `<orgname>/.github`, add
   `profile/README.md` with the tagline + GitHub banner.

## Step 3 · Telegram (3 min)

1. Channel (broadcast): https://t.me/+ → "New Channel"
   - Name: `AGENTIC`
   - Description: from `bios.md` Telegram section
   - Type: Public, link `t.me/agenticchain` (or fallback)
   - Photo: `logo.png` (512×512)
2. Group (chat): https://t.me/+ → "New Group"
   - Name: `AGENTIC · Chat`
   - Public link: `t.me/agenticchainchat` (or fallback)
   - Permissions: Allow members to send messages; disable forwarded messages
     from new accounts; enable slow-mode 5s to deter spam.
3. Pin in both: link to the X handle + the GitHub repo.

## Step 4 · Discord (5 min)

1. Sign in at https://discord.com/ and create a server.
2. Server name: `AGENTIC`
3. Upload `logo.png` (512×512) as server icon.
4. Create the channel structure from `bios.md` → Discord section.
5. Roles to create:
   - `@Validator` (manual approval; for confirmed validator operators)
   - `@AgentOperator` (anyone who's registered an agent on testnet)
   - `@Builder` (anyone who's opened a PR to the repo)
   - `@Mod` (you + co-maintainers)
6. Use the free [MEE6](https://mee6.xyz) bot for moderation; turn on
   anti-spam at "Strict".
7. Generate an invite link (`Server Settings → Invites → Create Invite →
   Never expire`). Paste this link into the X bio.

## Step 5 · Farcaster (2 min, $5 in ETH)

1. Sign up via Warpcast: https://warpcast.com/~/signup
2. Pay the registration fee (~$5 in ETH on Base) — this is the **only paid
   step** in the entire playbook. Skip if strict-$0.
3. Bio: from `bios.md` → Farcaster section.
4. Set custody address to the same wallet that will deploy the chain (so the
   social identity is bound to the genesis key).

## Step 6 · YouTube (3 min)

1. Sign in to YouTube → Create channel: https://youtube.com/create_channel
2. Channel handle: try `@agenticchain` first.
3. Display name: `AGENTIC`
4. About: paste from `bios.md` → YouTube section.
5. Channel art: `banner-youtube.png` (2560×1440).
6. Avatar: `logo.png` (800×800).
7. Verify the account (raises upload limit + unlocks custom thumbnails) —
   takes 24h.

## Step 7 · LinkedIn page (4 min)

1. https://www.linkedin.com/company/setup/new/
2. Page type: Company → Small business
3. Name: `AGENTIC`
4. Vanity URL: `agentic-chain` (fallbacks in `handles.md`)
5. Industry: `Blockchain Services`
6. Tagline: from `bios.md` → LinkedIn tagline.
7. About: from `bios.md` → LinkedIn About section.
8. Logo: `logo.png` (400×400). Cover: `banner-linkedin.png` (1128×191).
9. Add the maintainer accounts as page admins.

## Step 8 · Medium / Mirror (3 min)

**Medium** — https://medium.com/m/signin → settings → public profile name →
set to `agenticchain`. Create a Publication: Settings → Publications → New →
name `AGENTIC`, URL slug `agenticchain`. Add bio from `bios.md`.

**Mirror.xyz** — https://mirror.xyz/dashboard → "Create entry". Sign in with
the maintainer wallet. Claim subdomain `agentic.mirror.xyz`. The first post
should be the architecture doc cross-posted from `genesis/docs/01-architecture.md`.

## Step 9 · Reddit (2 min)

1. Create the subreddit: https://www.reddit.com/subreddits/create
2. Name: `r/AgenticChain` (fallbacks in `handles.md`)
3. Type: Public
4. Topic: Cryptocurrency
5. Add the sidebar text from `bios.md` → Reddit section.
6. Create the rules from the brand voice in `brand.md`.

## Step 10 · TikTok / Instagram (optional, 4 min)

Skip in v0 unless you have time. When you do them, paste the matching bio
from `bios.md` and use the same `logo.png` + cover image.

---

## After signup — send the confirmed handles back to me

DM me a list like:

```
X:         @agenticchain
GitHub:    agentic-chain
Telegram:  @agenticchain (channel), @agenticchainchat (group)
Discord:   discord.gg/xY3z9pQ
Farcaster: @agentic
YouTube:   @agenticchain
LinkedIn:  linkedin.com/company/agentic-chain
Medium:    @agenticchain
Mirror:    agentic.mirror.xyz
Reddit:    r/AgenticChain
```

I will then:
1. Update every `agentic.dev` / `@agenticchain` placeholder in the repo's
   docs, bios, and READMEs to point at the real handles.
2. Generate the first week of platform-specific posts in
   [`launch-week-posts.md`](launch-week-posts.md) with the correct handles.
3. Open a follow-up PR with the link-update diff.
