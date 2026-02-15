# Ink MCP / ml.ink — Marketing Playbook

Your product is unusually demo-friendly: it turns *code* into a *public URL* with minimal human involvement. Lean into "proof of autonomy" and make every piece of content show **time-to-URL** and **error-to-green** recovery.

---

## The Core Story (repeat everywhere)

1. Agents can write apps now.
2. They still can't deploy without humans.
3. Ink MCP makes deploys agent-native: one tool, one URL, no dashboards.

**Signature format:** *agent sees failure → reads logs → fixes → redeploys → green*.

---

## Video Content

### Stack Demos (SEO workhorses — do these first)

- [x] "Claude deploys a full-stack Next.js app in 47 seconds"
- [x] "Cursor builds and deploys a Flask API with SQLite — zero config"
- [x] "Agent builds a SaaS from a napkin sketch" (give it a screenshot of a hand-drawn wireframe)
- [x] "Deploying a Go backend + React frontend with one conversation"
- [x] Every popular framework gets its own video. These are long-tail search magnets.
- [x] Debug live deployments that fail video.

### Self-Referential (these get shared because they're meta)

- [x] "I asked Claude to deploy itself" — clonebot idea, great hook
- "AI agent deploys an MCP server that deploys other MCP servers" — recursion is inherently shareable
- [x] "I gave an agent my startup idea and went to lunch. It was live when I got back." — timelapse, real clock on screen
- [x] "Claude fixes its own production bug" — agent sees crash in logs, reads the error, pushes a fix, redeploys. Full loop, no human.
- [x] AI agent deploys a claudebot that deploys another claudebot

### Competition Bait (controversial = engagement)

- [x] "Deploying to ml.ink vs Railway vs Vercel — which one can an agent actually use?" — side by side, same prompt, three platforms. You win because yours is purpose-built.
- [x] "How many tool calls does it take to deploy on Railway MCP?" — count them. Then show yours. The gap is the marketing.
- [x] "I tried to deploy a backend on Vercel" — short, painful, ends with switching to ml.ink. Relatable frustration content.

### Constraint Videos (sports / rules = shareable)

- "Deploy a full-stack app using only **3 tool calls**." (tool-call counter on screen)
- "Deploy a SaaS where the agent **can't use the word 'Docker'**."
- "Deploy an app in **90 seconds** before I cut the feed." (timer + hard stop)
- "Deploy a project where every commit message must be a **haiku**."
- "Deploy from **one sentence**… then ship a bugfix from **a second sentence**."

### Error-to-Green Mini-Dramas (highest conversion)

- "Port detection fails → agent reads error → sets port → redeploys."
- "Missing env var → agent discovers it in logs → provisions resource → auto-wires → redeploys."
- "Crash loop speedrun: fastest fix wins." (you vs model vs model)
- "Deliberately broken app deployment… watch the agent self-heal it."

### Agent-as-SRE Series (recurring storyline)

- "Production incident #1: OOM → agent scales + redeploys."
- "Incident #2: bad deploy → agent rolls back."
- "Incident #3: DB token rotated → agent re-wires secrets + confirms health."

### Economics / Scale (infra crowd + trust)

- "100 deploys: time, cost, failures, retries." (real stats)
- "Build cache hits: why the 10th deploy is 5x faster."
- "What rate limits look like when an agent goes feral (and how the platform contains it)."

### Reverse Demo (start with the URL)

Open with the live URL, then reveal:
- "This app didn't exist 60 seconds ago."
- Rewind/timelapse the agent creating → pushing → deploying.

### The Absurd (these go viral on Twitter/TikTok)

- [xx] "Agent deploys 100 apps in 10 minutes" — speedrun. Show the counter going up.
- [x] "I let an AI run my infrastructure for a week" — daily diary format. What did it deploy? Did anything break?
- [x] "Non-technical person deploys a full-stack app by just describing it" — get your mom/friend to talk to Claude. The less technical the better.
- [x] "Agent vs junior developer: who deploys faster?" — race format. The agent wins. Funny and slightly threatening.
- [x] "I asked 5 different AI models to deploy an app. Only one succeeded." — model comparison gets massive engagement.

### Educational (build trust, establish authority)

- [x] "How MCP actually works — explained by building a deploy platform" — technical deep dive
- [x] "Why every AI agent will need a deploy tool" — thought leadership
- [x] "The architecture behind ml.ink" — behind-the-scenes infra. Show k3s, Longhorn, Temporal.
- [x] MCP tool with its own docs and bug reports that agent can submit and another agent will fix.
- [x] Different AI as a Skill for double-checking architecture

---

## Written Content

### Blog Posts That Earn Backlinks

- [x] "The Agent Deploy Problem: Why AI needs infrastructure designed for it" — the manifesto. Founding narrative. Post on blog, crosspost to HN.
- [x] "We analyzed 10,000 agent deployments. Here's what they build." — once you have usage data, this is gold.
- [x] "MCP Tool Design: What we learned building for AI agents" — practical lessons. Developers will bookmark this.
- [x] "How to make your app deployable by an agent in 5 minutes" — practical guide.

### "Agent-First Infra" Essay Series

- Why agent deployments need **task semantics** (progress/cancel/resume)
- Create vs retry-safe (idempotency keys vs magical upserts)
- Error messages as UX for non-human users
- Anatomy of an autonomous deploy loop

### SEO Content (boring but effective)

- [x] "How to deploy [framework] with AI" for every framework: Next.js, Flask, Django, Express, FastAPI, Go, Rails, Laravel, Rust Axum, etc.
- [x] "MCP server for deployment" — own this search term
- "Claude deploy app", "Cursor deploy backend", "AI deploy full stack" — every search term an early adopter might type

### Publish Reason Codes as Public Docs

`BUILD_FAILED`, `PORT_NOT_LISTENING`, `HEALTHCHECK_TIMEOUT`, etc. Agents love enums. Developers love clarity.

### "Make Your Repo Agent-Deployable" Checklist

Useful even if they don't use you (but suggests you as the obvious destination):
- PORT handling
- Health endpoint
- Env var patterns
- Build/runtime commands
- Migration/seed conventions

Bonus: ship a GitHub Action that runs the checklist and comments on PRs.

---

## Getting Ink Tokens Into The World

### MCP Directory Listings

- [xx] Get listed on every MCP directory, registry, and awesome-list. Smithery, Glama, awesome-mcp-servers on GitHub.
- [xx] Write the best README in every directory. The README is your storefront.

### Client Integration Guides

- [xx] Write dedicated setup guides for each client: "Add ml.ink to Cursor in 30 seconds", "One-click ml.ink for Claude Desktop". Make it trivially easy.
- [xx] If any client has featured/recommended MCP servers, get on that list. Reach out to their DevRel.

### Open Source Presence

- The MCP server itself should be open source. Developers trust what they can read.
- Contribute to MCP ecosystem projects. PRs, issues, discussions.
- [xx] Create example repos: `ml-ink-nextjs-template`, `ml-ink-flask-starter`, etc.

### Developer Community Seeding

- [xx] Post every video and blog post to HN, r/programming, r/artificial, r/ChatGPT, r/cursor, relevant Discord servers.
- Don't spam — contribute. Become the person people associate with "agent deployment."
- [xx] Find threads where people complain about deploying from AI agents. That's your opening.

### Referral / Word-of-Mouth

- [x] Every deployed app at `*.ml.ink` is a walking billboard. The subdomain IS the marketing.
- "Deployed with ml.ink" badge/footer — opt-in, extra free tier for displaying it.
- [x] Free tier must be generous enough that people deploy things they wouldn't have bothered deploying otherwise.

### "Deploy to ml.ink" Button (agent-native)

A README badge that opens a page with:
- Copy/paste MCP config for Claude/Cursor/Windsurf
- A prompt snippet that deploys the repo

Each starter repo becomes a distribution channel.

### Preview URLs with TTL

Every PR/branch deploy gets `https://repo-pr-42.ml.ink` with auto-cleanup after 72h. Every PR comment thread carries your brand.

---

## Partnerships & Integrations

- [x] Reach out to AI coding tool creators (Cursor, Windsurf, Cline, Aider, Claude Code). Offer to be their recommended deploy target.
- [x] Reach out to AI tutorial creators on YouTube. Offer free accounts.
- [x] Turso partnership — cross-promote: "The best SQLite deploy stack: Turso + ml.ink."
- Micro-partnerships: ship a starter + video + post for pairings (Turso + ml.ink, SvelteKit + ml.ink, FastAPI + ml.ink).

---

## Weird Stunts That Can Go Viral

### The "URL Receipt" Share Card

After every deploy, generate a shareable card image: app name, URL, time-to-URL, framework, "deployed by an agent" stamp. People share images more than text.

### Deploy Roulette (curated)

Public page: submit a public repo + prompt "make it deployable." Agent attempts it live with a scoreboard.

### The Museum of Useless Apps

Gallery of tiny hilarious apps at `*.ml.ink`: "rate my houseplant", "excuse generator for meetings", "a button that apologizes". Each with opt-in "deployed on ml.ink" footer.

### The 1-Minute Startup (weekly)

Viewers submit ideas → pick one → agent builds + deploys in 60–180 seconds → post the URL.

### The "No Dashboard" Pledge

"If it requires clicking around a provider UI, it doesn't count." Then deploy real things without opening dashboards.

### Agent-Built Landing Pages

- [xx] Let people type a description and get a live landing page at `something.ml.ink` in 30 seconds. Viral because people share generated pages.

### Deploy Challenge

- [x] "Deploy something cool → post the ml.ink URL → best one wins [prize]."

### MCP Benchmark/Leaderboard

- [xx] Public benchmark: tool calls, time to deploy, error recovery. You win your own benchmark (designed for it), but openness builds credibility.

### "Deploy It For Me" Bot (Twitter/Reddit)

- [xx] People post code screenshots or repo links, tag your bot, it deploys and replies with URL.
- **Safe version:** explicit opt-in, strict rate limits, public repos only, TTL previews, abuse reporting.

### Agent Hackathon Sponsorship

- [xxx] Sponsor AI hackathons. Offer ml.ink as the deployment layer.

### Public Directory — Ink Builders

- [xxx] Every app deployed via Ink MCP (opt-in/out) automatically hosted on Ink Builders gallery.

---

## Branding: "ml.ink Everywhere" Without Being Annoying

### Petname Subdomains

Lean into weird/cute default names ("rename later"). People share weird names.

### Receipt Endpoint

Every service exposes `/.mlink` showing: deployed at, commit SHA, framework, status, "deployed by agent" stamp.

### Stickers

QR stickers: "Add ml.ink to Claude/Cursor in 30 seconds."

### Celebrate the "First Deploy" Moment

Make it shareable: confetti, share card, "your first URL is live" story.

---

## Distribution: Creator Kits

Make a "demo kit" for YouTubers:
- 5 prompts + 5 repos + 5 failure scenarios that recover cleanly
- Thumbnails, B-roll, suggested titles

### Deploy Clinic (live weekly stream)

Viewers submit repos → pick 3 → agent tries to deploy → narrate results and lessons.

---

## Priority Order

**Week 1–2:** MCP directory listings, open source the server, client setup guides. Distribution infrastructure — everything else builds on it. 10 stack demos (SEO). 5 error-to-green loops (conversion). 1 speedrun (viral).

**Week 3–4:** Share cards. Gallery (opt-in). Preview TTL. The manifesto blog post. The self-referential video. The competition video.

**Month 2:** Publish reason codes. Publish "agent-deployable repo" checklist. Run weekly deploy clinic.

**Ongoing:** One video per week, one blog post every two weeks, constant community presence. Consistency beats virality.

---

## The Core Insight

**Every ml.ink URL in the wild is marketing.** The product markets itself every time someone shares what they built. Your job is to make it so easy that people deploy things they wouldn't have bothered deploying otherwise — and then share the URL because it just works.

---

## AI Coding Client Support Matrix

Prioritized list of clients to support with Ink MCP integration guides, tested workflows, and marketing content.

### Tier 1 — Must-Have (millions of users, highest ROI)

| Client | Est. Users | MCP Support | OpenRouter Usage | Priority | Notes |
|--------|-----------|-------------|-----------------|----------|-------|
| **GitHub Copilot** | 1.5M+ paid | Yes | N/A (closed ecosystem) | **P0** | The 800-lb gorilla. Works across VS Code, JetBrains, Neovim, CLI. If your MCP works here, instant credibility. Microsoft distribution = unmatched reach. |
| **Cursor** | Millions | Deep native | N/A (own routing) | **P0** | $9B valuation, $900M raised. THE AI-native IDE. Multi-model, deep MCP. Dominant among "vibe coders." Every Cursor tutorial is a potential ml.ink funnel. |
| **Claude Code** | Millions | Native, central | 28.6B tokens (#6) | **P0** | Best AI coding assistant (Jan 2026 consensus). MCP is core to how it works. Our product is literally built on this ecosystem. Huge mindshare among serious devs. |

### Tier 2 — High Priority (hundreds of thousands, fast growth)

| Client | Est. Users | MCP Support | OpenRouter Usage | Priority | Notes |
|--------|-----------|-------------|-----------------|----------|-------|
| **Windsurf** (ex-Codeium) | 500K+ | One-click setup | N/A | **P1** | Acquired by OpenAI — signals massive future growth. "Cascade" agent, generous free tier. OpenAI backing means resources + distribution. |
| **Cline** | 300K+ | Yes (VS Code ext) | 31.1B tokens (#5) | **P1** | Open-source, human-in-the-loop approval. Popular with security-conscious devs. BYOK model = broad model compatibility means Ink MCP works across many LLMs through one client. |
| **Roo Code** (Cline fork) | 200K+ | Deep native | 15.6B tokens (#10) | **P1** | Fastest-growing fork. Multi-agent modes (Code/Architect/Orchestrator). Agent orchestration = complex deploy scenarios = more Ink MCP usage. Very active community. |

### Tier 3 — Important (significant, growing, strategic)

| Client | Est. Users | MCP Support | OpenRouter Usage | Priority | Notes |
|--------|-----------|-------------|-----------------|----------|-------|
| **Gemini CLI** | Growing fast | Yes | N/A | **P2** | Google's terminal agent. Free tier + 1M token context. Zero cost = massive adoption potential. Good for "deploy with Gemini" content series. |
| **OpenAI Codex CLI** | Growing | Yes (GA) | N/A | **P2** | MCP client built-in at GA. Cloud sandboxes. OpenAI brand = trust + tutorials. Good for "deploy with GPT" content. |
| **Augment Code** | Enterprise | Yes (Context Engine MCP) | N/A | **P2** | #1 SWE-Bench Pro. IDE ext + CLI. Enterprise teams = paid tier conversions. |
| **Amazon Q Developer** | AWS shops | Yes | N/A | **P2** | Strong in AWS-centric orgs. "Deploy without AWS complexity" angle could resonate. |
| **Continue** | Growing | Yes | N/A | **P3** | Open-source, self-hosted/local model friendly. Good for privacy-first devs. |
| **OpenCode** | Niche | Yes | N/A | **P3** | 75+ providers, TUI, privacy-first. Small but loyal community. |

### Tier 4 — Watch List (OpenRouter rising stars, evaluate quarterly)

| Client | OpenRouter Usage | MCP Support | Priority | Notes |
|--------|-----------------|-------------|----------|-------|
| **Kilo Code** | 182B tokens (#2!) | VS Code ext | **P3** | Enormous token volume on OpenRouter — second only to OpenClaw. Worth testing integration and creating a setup guide. Growing fast. |
| **OpenClaw** | 273B tokens (#1) | Unknown | **P3** | Highest OpenRouter volume by far. "The AI that actually does things" — if it supports MCP or tool use, could be a dark horse. Investigate. |
| **BLACKBOXAI** | 39B tokens (#4) | Unknown | **P4** | "AI agent for builders" — relevant positioning. Monitor for MCP support. |
| **Z Code** | 5.77B tokens (new) | Unknown | **P4** | Brand new, growing. Watch. |
| **Agent Zero** | 3.54B tokens | Unknown | **P4** | Autonomous agent framework. If it gains MCP support, natural fit for Ink MCP. |

### Client Strategy Notes

- **Content per client:** Each Tier 1–2 client should have a dedicated setup guide, demo video, and blog post. Framework this as "[Client] + ml.ink" series.
- **OpenRouter insight:** Cline + Claude Code + Roo Code together account for ~75B tokens/week on OpenRouter alone. These three are disproportionately important for the "bring your own model" crowd.
- **Windsurf/OpenAI acquisition:** With OpenAI acquiring Windsurf, expect a unified Codex+Windsurf experience. Early Ink MCP support on both means you're ready when they merge.
- **Copilot's MCP adoption:** GitHub Copilot adding MCP is the single biggest distribution event for MCP tools in 2026. Priority #1 is ensuring Ink MCP works flawlessly here the day it launches widely.
- **Agent orchestration trend:** Roo Code's multi-agent modes and Agent Zero's autonomous agents point toward multi-step deploy workflows. Build Ink MCP to handle agent chains (build → test → deploy → verify).

---

## LLM Leaderboard — Models to Feature in Content

The models people actually use determine which "Deploy with [Model]" videos get views. Prioritize content around high-usage models.

### This Week's OpenRouter Top Models (Feb 2026)

| Rank | Model | Provider | Weekly Tokens | Trend | Deploy Content Priority | Notes |
|------|-------|----------|--------------|-------|------------------------|-------|
| 1 | **Kimi K2.5** | Moonshot AI | 1.47T | +71% | **High** | Dominant volume. Chinese AI breakout star — huge international audience. "Deploy with Kimi" taps into a massive, underserved market. Strong coding capabilities. |
| 2 | **Gemini 3 Flash Preview** | Google | 761B | +4% | **High** | Google's latest. Free tier drives adoption. Fast + cheap = agent-friendly. Great for "budget deploy" content angle. |
| 3 | **DeepSeek V3.2** | DeepSeek | 742B | +19% | **High** | Open-weight champion. Huge in Asia + cost-conscious devs. "Deploy for free with DeepSeek + ml.ink" = viral combo. |
| 4 | **Claude Sonnet 4.5** | Anthropic | 649B | +9% | **Must-do** | Our ecosystem. Best balance of speed/quality for agent coding. Every ml.ink demo should work flawlessly with Sonnet. |
| 5 | **Grok 4.1 Fast** | xAI | 542B | +72% | **Medium** | Surging. xAI/Twitter audience = viral distribution channel. "Deploy with Grok" could get Elon engagement (free impressions). |
| 6 | **MiniMax M2.5** | MiniMax | 492B | **New** | **Low** | Brand new entry. Monitor. If coding capabilities are strong, create content early to own the SEO. |
| 7 | **Claude Opus 4.6** | Anthropic | 489B | **+1,006%** | **Must-do** | Explosive growth. Best reasoning model. Showcase complex multi-step deploys (monorepos, microservices, DB migrations). Premium content. |
| 8 | **MiniMax M2.1** | MiniMax | 481B | +19% | **Low** | Predecessor to M2.5, still popular. Monitor coding quality. |
| 9 | **Gemini 2.5 Flash** | Google | 445B | +7% | **Medium** | Workhorse model. Solid for "reliable deploy" content. |
| 10 | **Trinity Large Preview** | Arcee AI | 423B | +49% | **Low** | Free tier driving adoption. Open model. Niche but growing. |
| 11 | **Gemini 2.5 Flash Lite** | Google | 372B | +16% | **Low** | Ultra-cheap. "Deploy on a budget" angle. |
| 12 | **gpt-oss-120b** | OpenAI | 300B | +14% | **Medium** | OpenAI's open-source play. If it codes well, "Deploy with open-source GPT" is compelling content. |
| 13 | **GPT-5 Nano** | OpenAI | 297B | +39% | **Medium** | Small + fast. Good for "instant deploy" speedrun content. Shows Ink MCP works even with lightweight models. |
| 14 | **Grok Code Fast 1** | xAI | 290B | +10% | **High** | Coding-specific Grok variant. Purpose-built for code = perfect for deploy demos. Natural pairing. |
| 15 | **Claude Opus 4.5** | Anthropic | 258B | +36% | **Medium** | Previous-gen Opus still popular. Good for "model comparison" content alongside Opus 4.6. |
| 16 | **Step 3.5 Flash** | StepFun | 201B | **+381%** | **Low** | Explosive growth, free tier. Chinese AI. Watch for coding capabilities. |
| 17 | **Pony Alpha** | OpenRouter | 196B | **+19,094%** | **Skip** | Creative/RP model, not relevant for deploy content. |
| 18 | **GLM 5** | Z-AI | 183B | **New** | **Low** | Zhipu's latest. Monitor. Could be big in Chinese developer market. |
| 19 | **Gemini 3 Pro Preview** | Google | 176B | +7% | **High** | Google's flagship reasoning model. Complex deploy scenarios. |
| 20 | **Gemini 2.0 Flash** | Google | 163B | +7% | **Low** | Older but stable. Good baseline for benchmark content. |

### Model Content Strategy

**Must-produce content (Tier 1 models):**
- Claude Sonnet 4.5 / Opus 4.6 — flagship demos, these should be perfect
- Kimi K2.5 — tap into massive Chinese/international audience, underserved by Western deploy tools
- DeepSeek V3.2 — "deploy for free" angle, open-weight community loves this
- Gemini 3 Flash/Pro — Google ecosystem, free tier, massive reach

**High-opportunity content:**
- Grok Code Fast 1 — coding-specific model + Twitter/X virality. Tag @xAI in demos.
- GPT-5 Nano / gpt-oss-120b — OpenAI brand recognition drives clicks regardless
- "Model shootout" videos — deploy same app with 5 different models, compare tool calls, time, success rate. These get massive engagement.

**Trends to exploit:**
- **Opus 4.6 at +1,006%** — ride this wave NOW. Every "Opus 4.6 deploys X" video will get boosted by algorithmic interest.
- **Kimi K2.5 at #1** — Western developers are discovering Moonshot. First-mover on "Kimi deploys" content.
- **Step 3.5 Flash at +381%** — free models drive experimentation. "Deploy your first app for $0" using free models + ml.ink free tier.
- **OpenAI open-source (gpt-oss-120b)** — huge narrative shift. "Open-source GPT deploys to open infrastructure" story writes itself.
- **Google's Gemini 3 family** — 3 Flash Preview + 3 Pro Preview both in top 20. Google I/O content timing opportunity.

**Cross-reference with clients:**
- Cursor users → tend toward Claude, GPT models
- Claude Code users → Claude models (obviously), but also DeepSeek via OpenRouter
- Cline/Roo Code users → model-agnostic via OpenRouter, so content for ANY top model drives Ink MCP adoption through these clients
- Gemini CLI users → Gemini models exclusively. Dedicated "Gemini + ml.ink" series.

---

## Caution Flags

### "Deploy It for Me" Bot
Safe version: explicit opt-in, strict rate limits, public repos only, default TTL previews, abuse reporting.

### "Replicates Itself" Gimmicks
Make it a joke *with* safety rails: quotas, TTL, cleanup, deterministic names. The guardrails become part of the flex.
