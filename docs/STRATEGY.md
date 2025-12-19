# Ubik — Complete Strategy

**Legal Entity**: Rastrigin Systems Ltd
**Product**: Ubik
**License**: Business Source License 1.1
**Model**: Source-Available
**Last Updated**: December 2025

---

## Contents

1. [The Goal](#1-the-goal)
2. [Company & Brand](#2-company--brand)
3. [Market Opportunity](#3-market-opportunity)
4. [Product Editions](#4-product-editions)
5. [Source-Available Model](#5-source-available-model)
6. [License (BSL 1.1)](#6-license-bsl-11)
7. [Repository & Development](#7-repository--development)
8. [Enterprise Revenue](#8-enterprise-revenue)
9. [Competitive Positioning](#9-competitive-positioning)
10. [Execution Plan](#10-execution-plan)
11. [Metrics & Kill Switch](#11-metrics--kill-switch)
12. [Exit Optionality](#12-exit-optionality)
13. [The Four Rules](#13-the-four-rules)
14. [Next Actions](#14-next-actions)

---

## 1. The Goal

> **"Build a profitable company that replaces my 9–5 income. Acquisition is an option, not the goal."**

**This is not about:**
- Changing the world
- OSS fame
- Out-featuring competitors
- Building to flip

**This is about:**
- Building something true
- For paying customers
- With independence as the primary outcome

**Personal metric:** Ubik covers 50-70% of personal burn within 12-18 months.

---

## 2. Company & Brand

| Element | Value |
|---------|-------|
| **Legal Entity** | Rastrigin Systems Ltd |
| **Product** | Ubik |
| **Domain** | rastrigin.systems |
| **GitHub** | github.com/rastrigin-systems |
| **Tagline** | "Find your global optimum" |

### Name Origin

**Rastrigin** — Named after the [Rastrigin function](https://en.wikipedia.org/wiki/Rastrigin_function), a famous optimization benchmark with many local minima but one global optimum.

### Product Line

```
RASTRIGIN SYSTEMS LTD
│
├── Ubik Core        Free, self-hosted, BSL 1.1
├── Ubik Enterprise  Paid license + services
└── Ubik Cloud       Managed SaaS (future)
```

---

## 3. Market Opportunity

### Market Size

| Metric | 2025 | 2034 | CAGR |
|--------|------|------|------|
| AI Governance Market | $309M | $4.8B | 35.74% |
| AI Agents Market | $7.7B | $105.6B | — |

### Market Validation

- 85% of enterprises expected to implement AI agents by end of 2025
- 42% fail to scale AI due to fragmented oversight
- 98% plan to increase AI governance budgets

### Target Buyers

| Buyer | Pain Point |
|-------|------------|
| VP Engineering | "How do I standardize AI tools across 50 teams?" |
| Platform Engineering | "How do I distribute approved AI tools?" |
| Security/CISO | "How do I audit what AI tools are doing?" |
| CFO | "How much are we spending on AI tools per team?" |

### Competitor Landscape

| Competitor | Model | Our Advantage |
|------------|-------|---------------|
| MintMCP | Closed source, funded | Source-available, auditable |
| GitHub Copilot Enterprise | Copilot only | Multi-tool (Claude, Cursor, Copilot) |
| Cursor Business | Single tool | Full governance platform |

**MintMCP is validation, not threat.** They prove the market exists.

---

## 4. Product Editions

### Core Edition (Free)

Everything needed for a complete, working product:

- Full authentication (JWT, sessions)
- Organization management
- Team management (unlimited)
- Employee management (no artificial limits)
- Agent configuration (org → team → employee hierarchy)
- MCP server management
- Activity logging
- Usage analytics
- CLI sync
- Docker Compose deployment

**Philosophy:** No artificial limits. Core is complete.

### Enterprise Edition (Paid)

**Year 1 Services (Keep It Narrow):**

| Service | Included | Notes |
|---------|----------|-------|
| Support SLA (48hr response) | ✅ | Realistic for solo founder |
| Security advisories | ✅ | Private email list |
| SSO setup & troubleshooting | ✅ | High-value, complex |
| Named support contact | ✅ | Direct access to founder |
| SOC 2 documentation | ❌ | Year 2 |
| Signed binaries | ❌ | Year 2 |
| Training sessions | ❌ | Year 2 |

**Year 1 focus:** Support + SSO. That's it.

**Code Features (Visible, License Required):**

- SSO/SAML/OIDC
- Advanced analytics
- Compliance exports
- Custom roles

### Enterprise Triggers (Process, Not Limits)

When to consider Enterprise:

- Guaranteed response times (SLA)
- Early access to security advisories
- SSO/SAML configuration and support
- Compliance documentation needs

---

## 5. Source-Available Model

### Why Source-Available (Not Open Source)

| Open Source Promise | Reality |
|---------------------|---------|
| "Community will build features" | 95% of projects get < 5 meaningful PRs/year |
| "Many contributors" | Most have 1-3 active contributors |
| "Free development labor" | Founder does 90%+ of the work |

**Conclusion:** We're not open-sourcing for community development. We're making code visible for **trust and auditability**.

### Source-Available vs Open Source

| Aspect | Open Source | Source-Available (Us) |
|--------|-------------|----------------------|
| Code visible | ✅ | ✅ |
| OSI-approved license | ✅ | ❌ (BSL isn't OSI) |
| Community expectations | High | Low |
| Commercial protection | None | BSL prevents competition |

### Why This Matters for Security Tools

Ubik handles:
- AI agent credentials
- Activity logs
- Sits in the data path

Enterprise security teams need to audit this. Our answer: "Read the code. All of it."

### Contribution Model

We welcome:
- Bug reports (GitHub Issues)
- Feature requests
- Security reports

We accept but don't expect:
- Code contributions (PRs)

**Message:** "We build Ubik. You use it and give feedback."

---

## 6. License (BSL 1.1)

### License Terms

```
Licensor:             Rastrigin Systems Ltd
Licensed Work:        Ubik
Change Date:          4 years from each release
Change License:       Apache License 2.0

Additional Use Grant:
You may use the Licensed Work for any purpose EXCEPT for
providing a commercial AI agent management or MCP governance
service that competes with Ubik.
```

### What BSL Allows

| Use Case | Allowed? |
|----------|----------|
| Use internally at your company | ✅ |
| Modify for internal use | ✅ |
| Audit and review all code | ✅ |
| Self-host on your infrastructure | ✅ |

### What BSL Prohibits

| Use Case | Allowed? |
|----------|----------|
| Sell Ubik as a service | ❌ |
| Host for paying customers | ❌ |
| Create competing commercial product | ❌ |

### 4-Year Conversion

Each version converts to Apache 2.0 four years after release.

**Why this doesn't hurt us:**
- 4-year-old code is outdated
- Customers pay for support, not old code
- Brand and relationships established by then

### CLA (Contributor License Agreement)

Required for all code contributions:
- Clean IP chain
- Acquisition readiness
- Flexibility to use contributions

**Tool:** CLA Assistant (free, automatic).

---

## 7. Repository & Development

### Current Approach (Keep It Simple)

```
ONE REPO: rastrigin-systems/ubik-enterprise (PRIVATE)

- All development happens here
- Stay private during validation
- Go public when you want growth, not before
```

### When to Go Public

| Trigger | Why |
|---------|-----|
| Want inbound leads | HN launch, SEO |
| Want to claim "source-available" | Marketing |
| Have 2-3 paying customers | Validation confirmed |

**Not before.** Public is for growth. You're in validation.

### Pre-Public Checklist

- [ ] Squash/clean history if needed
- [ ] LICENSE file (BSL 1.1)
- [ ] README.md (source-available positioning)
- [ ] SECURITY.md
- [ ] CLA.md

---

## 8. Enterprise Revenue

### What "Support" Actually Means

| Service | What It Means | Time Cost |
|---------|---------------|-----------|
| Priority email response | Reply within 48hr | Low |
| SSO configuration help | Video call to set up SAML/OIDC | 1-2 hrs |
| Deployment assistance | Help with Docker setup | 1-2 hrs |
| Direct Slack channel | Private channel for questions | Async |
| Bug prioritization | Their bugs fixed first | Variable |
| Monthly check-in | 30 min "how's it going?" | 30 min/month |

### The Real Value

| What They Think | What They Actually Want |
|-----------------|-------------------------|
| "Support SLA" | Someone to call when things break |
| "SSO help" | Not figuring it out themselves |
| "Priority bugs" | Feeling like they matter |

**The real product:** "You have a direct line to the person who built this."

### Pricing

| Tier | Price | Target |
|------|-------|--------|
| Core | Free | Evaluators, small teams |
| Enterprise | $500-1500/month | 20-100+ employees |

**At $1000/month:**
- 3 customers = $36K ARR
- 5 customers = $60K ARR
- 10 customers = $120K ARR

### Time Per Customer

~5-10 hrs/month at start, declining over time. With 5 customers: manageable.

---

## 9. Competitive Positioning

### Our Differentiation (Without Naming Competitors)

```
Ubik is source-available. Every line of code is visible.

Your security team can audit:
- What we log
- Where data goes
- That there are no backdoors

Self-host on your infrastructure.
Your data never leaves your network.
```

**Let buyers draw comparisons themselves.** Security buyers distrust vendor wars.

### What NOT to Say

| Avoid | Why |
|-------|-----|
| "Unlike MintMCP..." | Adversarial |
| Direct feature comparisons | Let them discover |

### What TO Say

| Say | Why |
|-----|-----|
| "Source-available" | Factual |
| "Fully auditable" | Security positioning |
| "Self-hosted" | Enterprise requirement |
| "Your data stays yours" | Privacy angle |

---

## 10. Execution Plan

### The Order

```
1. Design partner conversations (NOW)
2. Core development (parallel, informed by conversations)
3. First paid deal (from design partner)
4. Go public (when you want growth)
```

**Don't:** Build → Launch → Find customers
**Do:** Sell → Build (informed)

### Phase 1: Validation (Months 0-6)

| Week | Action |
|------|--------|
| 1-2 | Start design partner outreach |
| 2-8 | Continue Core development + conversations |
| 4+ | Give early access to interested partners |
| 8-12 | Core "done enough" based on feedback |
| 12-24 | First paid enterprise deal |

### Phase 2: Traction (Months 6-18)

| Milestone | Target |
|-----------|--------|
| Paying customers | 5-10 |
| Revenue | $60-150K ARR |
| Positioning | Clear and consistent |

### Phase 3: Outcome (Months 18-36)

Either:
- **Profitable independence** — sustainable revenue, freedom
- **Acquisition** — if attractive offer comes (optional)

### Founder Strategy

**Keep job until:**
- 2-3 paying customers, OR
- 12+ months runway

**Time allocation:**
| Activity | Priority |
|----------|----------|
| Design partner conversations | Highest |
| Core development (informed) | High |
| Strategy documents | Done (stop) |
| Feature additions | Only if customers ask |

---

## 11. Metrics & Kill Switch

### Year 1 Targets

| Metric | Target |
|--------|--------|
| Design partner conversations | 10-20 |
| Paying customers | 3-5 |
| Revenue | $30-100K ARR |

### Year 2-3 Targets

| Metric | Target |
|--------|--------|
| Paying customers | 10-30 |
| Revenue | $100-300K ARR |

### Success Signals

| Signal | Meaning |
|--------|---------|
| 30-60 min calls with leads | Right track |
| "Can you add X for us?" | Real demand |
| Enterprise procurement questions | Serious buyer |

### Kill Switch

> **"If after 20 serious design partner conversations no one indicates willingness to pay, reassess positioning or stop."**

This protects against:
- Sunk cost bias
- Emotional attachment
- Endless iteration

---

## 12. Exit Optionality

### Acquisition is an Option, Not a Dependency

The goal is a **profitable, independent company**. We structure for optionality:

- Clean IP (BSL + CLA)
- Single legal entity
- Clear product wedge
- Paying customers

If attractive offer comes, consider it. If not, sustainable business.

### What Makes Us Attractive (If Acquired)

| Asset | Why It Matters |
|-------|----------------|
| Clear product wedge | Differentiated |
| Paying customers (3-10) | Proven demand |
| Clean IP | No complications |
| Founder | Can transition smoothly |

### Potential Acquirers

| Type | Why They Buy |
|------|--------------|
| Platform vendors | Accelerate roadmap |
| Funded competitors | Acquire credibility |
| Enterprise aggregators | Roll into suite |

---

## 13. The Four Rules

### 1. Don't Try to Win the Market

Build something true for paying customers. Stop expanding scope to feel competitive.

### 2. Conversations Before Code

Start design partner outreach now. You'll learn more in 3 calls than 3 months of coding.

### 3. Services Over Features

Enterprise value comes from support and SSO help — not code unlocks.

### 4. Independence Over Exit

Build for profitability first. Acquisition is optionality, not the goal.

---

## 14. Next Actions

| Priority | Action | When |
|----------|--------|------|
| **1** | Draft design partner outreach message | Today |
| **2** | Identify 10 target orgs | This week |
| **3** | Start 3-5 conversations | Next 2 weeks |
| **4** | Continue Core development | Ongoing |
| **5** | Stop iterating strategy docs | Now |

---

## Summary

| Question | Answer |
|----------|--------|
| What's the goal? | Profitable company that replaces 9-5 |
| Is acquisition the goal? | No — it's an option |
| Are we open source? | No — source-available |
| Why make code visible? | Trust and auditability |
| How do we make money? | Enterprise services (support + SSO) |
| What's the personal metric? | 50-70% of burn in 12-18 months |
| What's the kill switch? | 20 conversations, no pay signal = reassess |
| When to go public? | When you want growth, not before |
| When to quit job? | 2-3 paying customers or 12+ months runway |

---

*Strategy finalized. Goal: profitable independence. Time to execute.*
