# Ubik Open Source Strategy Research

**Date**: December 2025
**Purpose**: Research on existing OSS competitors, licensing strategies, and how to structure Ubik for commercial success / acquisition

---

## Part 1: Existing Open Source Competitors

### Direct Competitors (MCP Management)

| Project | Stars | License | Focus | Threat Level |
|---------|-------|---------|-------|--------------|
| **MCPJungle** | 737 | MPL-2.0 | MCP gateway + registry | 🔴 High |
| **Obot MCP Gateway** | ~200 | OSS | SSO, RBAC, audit logging | 🟡 Medium |
| **MCP Context Forge (IBM)** | Unknown | OSS | FastAPI gateway, auth, rate-limiting | 🟡 Medium |
| **Microsoft MCP Gateway** | Unknown | OSS | Kubernetes, session routing | 🟡 Medium |
| **MetaMCP** | Unknown | OSS | MCP aggregation, OAuth | 🟢 Low |

### MCPJungle Deep Dive (Closest OSS Competitor)

**Repository**: https://github.com/mcpjungle/MCPJungle

| Aspect | Details |
|--------|---------|
| **Language** | Go |
| **Database** | SQLite (default), PostgreSQL |
| **License** | MPL-2.0 (weak copyleft) |
| **Stars** | 737 |
| **Focus** | MCP server registry + gateway |

**Features MCPJungle HAS:**
- ✅ Unified MCP gateway (single endpoint for all servers)
- ✅ Register HTTP and STDIO MCP servers
- ✅ Tool discovery and management
- ✅ Enable/disable tools per server
- ✅ Tool Groups (expose subset of tools)
- ✅ Bearer token authentication
- ✅ Claude + Cursor integration
- ✅ Docker deployment
- ✅ PostgreSQL support

**Features MCPJungle LACKS:**
- ❌ Agent configuration management (system prompts, settings)
- ❌ Hierarchical config (org → team → employee)
- ❌ Multi-AI-tool support (only MCP, not Copilot/Cursor configs)
- ❌ Approval workflows
- ❌ Usage/cost tracking
- ❌ CLI distribution (`ubik sync`)
- ❌ Hybrid API token model
- ❌ SSO/SAML (basic auth only)
- ❌ Activity logging (beyond basic)
- ❌ Team management (users, roles)

**Current Limitations (from their docs):**
- "New sub-process started for every tool call" — no persistent connections
- "SSE support exists but is not mature"
- macOS binaries not notarized

### Self-Hosted AI Coding Assistants

| Project | Stars | What It Is |
|---------|-------|------------|
| **Tabby** | 24K+ | Self-hosted Copilot alternative (code completion) |
| **Aider** | 25K+ | CLI-based AI coding assistant |
| **Continue** | 20K+ | Open-source AI code assistant for IDEs |

**Note**: These are AI assistants themselves, not management/governance tools. Different category than Ubik.

---

## Part 2: Ubik's Differentiation Opportunity

### What Ubik Has That No OSS Project Has

| Feature | MCPJungle | Obot | IBM | Ubik |
|---------|-----------|------|-----|------|
| MCP gateway | ✅ | ✅ | ✅ | ✅ |
| **Agent config management** | ❌ | ❌ | ❌ | ✅ |
| **Hierarchical config (org/team/employee)** | ❌ | ❌ | ❌ | ✅ |
| **System prompt management** | ❌ | ❌ | ❌ | ✅ |
| **Multi-AI-tool configs** | ❌ | ❌ | ❌ | ✅ |
| **Approval workflows** | ❌ | ⚠️ | ❌ | ✅ |
| **CLI sync for developers** | ❌ | ❌ | ❌ | ✅ |
| **Usage/cost tracking** | ❌ | ❌ | ❌ | ✅ |
| **Hybrid token model** | ❌ | ❌ | ❌ | ✅ |
| Team management | ❌ | ✅ | ❌ | ✅ |

### Ubik's Unique Positioning

```
Existing OSS landscape:

MCPJungle/Obot/etc = "MCP Server Gateway"
                      (How do I expose MCP servers?)

Ubik = "AI Coding Tool Governance Platform"
       (How do I manage ALL AI tools across my organization?)
```

**The gap**: No open-source project manages the full AI coding tool lifecycle:
- What tools each team can use
- How they're configured (system prompts, policies)
- Configuration inheritance (org → team → employee)
- Approval workflows for access
- Developer onboarding (`ubik sync`)
- Cost visibility

---

## Part 3: License Options for Commercial OSS

### License Comparison

| License | Type | Can Sell Enterprise? | Can Be Forked by Competitor? | Used By |
|---------|------|---------------------|------------------------------|---------|
| **MIT** | Permissive | ✅ Yes | ✅ Yes (fully) | React, VS Code |
| **Apache 2.0** | Permissive | ✅ Yes | ✅ Yes (fully) | Kubernetes, Spark |
| **MPL-2.0** | Weak copyleft | ✅ Yes | ⚠️ Modified files must stay open | Firefox, MCPJungle |
| **AGPL-3.0** | Strong copyleft | ✅ Yes | ⚠️ SaaS must open source | Grafana, MongoDB (old) |
| **BSL 1.1** | Source-available | ✅ Yes | ❌ No competing service | HashiCorp, Sentry, CockroachDB |
| **SSPL** | Source-available | ✅ Yes | ❌ Must open entire stack | MongoDB, Elastic (old) |

### Recommended: Business Source License (BSL 1.1)

**Why BSL is best for your situation:**

1. **Protects against cloud competitors** — AWS/GCP can't offer "Ubik-as-a-service"
2. **Still source-available** — Users can self-host, see code, contribute
3. **Converts to open source** — After 4 years, becomes Apache 2.0 (or sooner)
4. **Proven model** — HashiCorp, Sentry, CockroachDB all successful with BSL
5. **Acquisition-friendly** — Acquirers understand BSL, it's business-friendly

**BSL Success Stories:**

| Company | License Change | Result |
|---------|---------------|--------|
| MongoDB | AGPL → SSPL (2018) | Usage grew 6x, IPO at $25B |
| Elastic | Apache → SSPL (2021) | Usage grew 4x |
| HashiCorp | MPL → BSL (2023) | Acquired by IBM for $6.4B |
| Sentry | BSD → BSL (2019) | Continued growth, raised $200M+ |
| CockroachDB | Apache → BSL | Raised $600M+ |

### BSL Key Terms

```
Business Source License 1.1

- You CAN: View, modify, self-host the code
- You CAN: Use it for internal business purposes
- You CANNOT: Offer it as a competing commercial service
- AFTER 4 YEARS: Converts to Apache 2.0 (or earlier date you set)
```

**Example BSL header:**
```
Licensed under the Business Source License 1.1 (the "License");
you may not use this file except in compliance with the License.

Change Date: December 2029 (4 years from now)
Change License: Apache License, Version 2.0

For non-production use, you may use this software freely.
For production use as a service, contact licensing@ubik.dev
```

---

## Part 4: Structuring for Acquisition / Monetization

### The Open Core Model

```
┌─────────────────────────────────────────────────────────────────┐
│                    UBIK COMMUNITY EDITION                        │
│                         (BSL 1.1)                                │
├─────────────────────────────────────────────────────────────────┤
│  Core Features (Free, self-hosted)                              │
│  ─────────────────────────────────                              │
│  • CLI tool (ubik sync)                                         │
│  • Self-hosted API server                                       │
│  • Agent configuration (Claude Code, Cursor, Copilot)           │
│  • MCP server management                                        │
│  • Basic team management (users, roles)                         │
│  • Hierarchical config (org → team → employee)                  │
│  • System prompt management                                     │
│  • Basic audit logging                                          │
│  • PostgreSQL backend                                           │
│  • Docker deployment                                            │
│  • Web UI (basic)                                               │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    UBIK ENTERPRISE                               │
│               (Commercial License Required)                      │
├─────────────────────────────────────────────────────────────────┤
│  Enterprise Features (Paid)                                     │
│  ───────────────────────────                                    │
│  • SSO/SAML/OIDC integration                                    │
│  • Advanced audit logs + compliance exports                     │
│  • Usage analytics dashboard                                    │
│  • Cost tracking and attribution                                │
│  • Advanced approval workflows                                  │
│  • Real-time policy enforcement                                 │
│  • PII detection and redaction                                  │
│  • Multi-region deployment                                      │
│  • Priority support + SLA                                       │
│  • Custom integrations                                          │
│  • Ubik Cloud (managed hosting)                                 │
└─────────────────────────────────────────────────────────────────┘
```

### What Makes Enterprise Features "Enterprise"

| Feature | Why It's Paid |
|---------|---------------|
| **SSO/SAML** | Enterprises MUST have it, will pay |
| **Compliance exports** | Legal/audit requirements |
| **Advanced analytics** | Management reporting |
| **Real-time blocking** | Security teams need it |
| **Priority support** | SLA guarantees |
| **Multi-region** | Global enterprises need it |

### Contributor License Agreement (CLA)

**Why you need a CLA:**
- Gives you legal right to relicense code (including contributions)
- Required if you want to offer commercial licenses
- Required if you want to be acquired (acquirer needs clean IP)

**CLA Options:**

| Type | What It Does | Example |
|------|--------------|---------|
| **Copyright Assignment** | Contributors give you ownership | Rare, contributor-unfriendly |
| **License Grant (CLA)** | Contributors keep copyright, grant you license | Elastic, MongoDB |
| **DCO (Developer Certificate of Origin)** | Contributors certify they can contribute | GitLab, Linux kernel |

**Recommendation**: Use a **CLA with license grant** (not copyright assignment)

Example CLA terms:
> "You retain copyright to your contributions, but grant Ubik a perpetual,
> worldwide, non-exclusive, royalty-free license to use, modify, and
> distribute your contributions under any license."

**CLA Tools:**
- [CLA Assistant](https://cla-assistant.io/) — Free, GitHub integration
- [CLA Bot](https://colineberhardt.github.io/cla-bot/) — Alternative

---

## Part 5: Acquisition Readiness Checklist

### What Acquirers Look For

| Factor | What They Want | How to Build It |
|--------|---------------|-----------------|
| **Clean IP** | No licensing issues | CLA from all contributors |
| **No legal risk** | Clear ownership | BSL license, trademark registered |
| **Traction** | Users, stars, activity | Marketing, community building |
| **Technology** | Quality code | Good architecture, tests, docs |
| **Team** | Talent to acquire | You + any contributors |
| **Strategic fit** | Solves their problem | Position for specific acquirers |

### Acquisition-Ready Checklist

**Legal / IP:**
- [ ] BSL license applied to all code
- [ ] CLA signed by all contributors
- [ ] Trademark filed for "Ubik" (if possible)
- [ ] No GPL-contaminated dependencies
- [ ] Clean third-party license audit

**Technical:**
- [ ] Well-documented codebase
- [ ] Test coverage >70%
- [ ] CI/CD pipeline
- [ ] Security best practices
- [ ] Scalable architecture

**Traction:**
- [ ] GitHub stars (target: 1,000+)
- [ ] Active users (track installs)
- [ ] Community engagement (Discord, discussions)
- [ ] Enterprise inquiries documented
- [ ] Case studies / testimonials

**Business:**
- [ ] Clear positioning vs. competitors
- [ ] Enterprise feature roadmap
- [ ] Pricing model defined
- [ ] Revenue (even small) from enterprise
- [ ] Relationship with potential acquirers

---

## Part 6: Potential Acquirers

### Tier 1: Strategic Fit (Most Likely)

| Company | Why They'd Acquire | Your Angle |
|---------|-------------------|------------|
| **Anthropic** | Native Claude Code governance | "Official governance layer for Claude Code" |
| **GitHub/Microsoft** | Copilot Enterprise enhancement | "Multi-tool governance including Copilot" |
| **GitLab** | AI governance for DevSecOps | "Complete AI coding governance" |

### Tier 2: Adjacent Players

| Company | Why They'd Acquire | Your Angle |
|---------|-------------------|------------|
| **MintMCP** | Eliminate OSS competitor | "Broader feature set, community" |
| **Datadog** | AI observability play | "Governance + observability" |
| **Snyk** | AI security governance | "Security-first AI governance" |
| **JetBrains** | IDE ecosystem expansion | "Governance for all AI assistants" |

### Tier 3: Infrastructure Players

| Company | Why They'd Acquire |
|---------|-------------------|
| **Cloudflare** | Edge AI governance |
| **Hashicorp/IBM** | Infrastructure governance expansion |
| **Atlassian** | Developer tools ecosystem |

---

## Part 7: Action Plan

### Phase 1: Prepare for Open Source (Weeks 1-4)

| Week | Tasks |
|------|-------|
| **1** | Choose BSL 1.1 license, add to all files |
| **1** | Set up CLA (cla-assistant.io) |
| **1** | Audit dependencies for license compatibility |
| **2** | Clean up code, improve documentation |
| **2** | Write compelling README with screenshots/demo |
| **3** | Set up GitHub Discussions, issue templates |
| **3** | Create CONTRIBUTING.md |
| **4** | Soft launch: Twitter, Reddit, communities |

### Phase 2: Build Community (Months 2-6)

| Month | Focus |
|-------|-------|
| **2** | Product Hunt launch |
| **2-3** | Engage early users, ship fixes fast |
| **3-4** | Add requested features |
| **4-5** | First enterprise feature (SSO) |
| **5-6** | Cloud hosting option (beta) |

### Phase 3: Monetize (Months 6-12)

| Month | Focus |
|-------|-------|
| **6-7** | Enterprise license available |
| **7-8** | First paying customers |
| **8-9** | Case studies published |
| **9-12** | Scale enterprise sales |

### Phase 4: Exit Options (Year 2+)

| Path | Trigger |
|------|---------|
| **Acquisition** | Strategic interest from Tier 1/2 |
| **Raise funding** | Strong traction, want to grow faster |
| **Profitable indie** | Sustainable revenue, enjoy the work |

---

## Part 8: Recommended License Text

### BSL 1.1 License Header

```
Business Source License 1.1

Licensor: [Your Name / Ubik Inc.]
Licensed Work: Ubik [version]
Change Date: [4 years from release date]
Change License: Apache License, Version 2.0

The Licensed Work is provided under the terms of the Business Source
License 1.1. As stated in Section 1.1 of the Business Source License 1.1:

"Production Use" means any use of the Licensed Work in a production
environment to provide a commercial product or service to third parties,
including any use that is competitive with the Licensed Work.

For information about alternative licensing arrangements, contact:
licensing@ubik.dev

Notice:
The Business Source License (this document, or the "License") is not
an Open Source license. However, the Licensed Work will eventually
be made available under an Open Source License, as stated in this License.
```

### Additional Use Grant (Recommended)

Add this to allow self-hosting for internal use:

```
Additional Use Grant:

You may use the Licensed Work for your internal business purposes,
including self-hosting for your own organization's use, without a
commercial license.

A commercial license is required only if you offer the Licensed Work
to third parties as a managed service or as part of a commercial
product.
```

---

## Summary

### Key Decisions

| Decision | Recommendation | Rationale |
|----------|----------------|-----------|
| **License** | BSL 1.1 | Protects commercial interest, proven model |
| **CLA** | Yes, license grant | Required for relicensing, acquisition |
| **Open Core** | Yes | Free community, paid enterprise |
| **Change Date** | 4 years | Standard, provides protection |
| **Change License** | Apache 2.0 | Most business-friendly |

### Ubik's Unique Position

**No other open-source project offers:**
1. Full AI coding tool governance (not just MCP)
2. Hierarchical configuration (org → team → employee)
3. System prompt management across hierarchy
4. Approval workflows
5. CLI distribution for developers
6. Usage/cost tracking

**This is your moat.** MCPJungle and others only do MCP gateway. You do the full governance stack.

---

## Sources

### Open Source Competitors
- [MCPJungle GitHub](https://github.com/mcpjungle/MCPJungle)
- [Awesome MCP Enterprise](https://github.com/bh-rat/awesome-mcp-enterprise)
- [Tabby Self-Hosted AI](https://github.com/TabbyML/tabby)

### Licensing
- [BSL Wikipedia](https://en.wikipedia.org/wiki/Business_Source_License)
- [HashiCorp BSL Announcement](https://www.hashicorp.com/en/blog/hashicorp-adopts-business-source-license)
- [MongoDB SSPL FAQ](https://www.mongodb.com/legal/licensing/server-side-public-license/faq)
- [Sentry BSL Success](https://www.dotcms.com/blog/bsl-in-action-whos-doing-it-and-does-it-work)

### CLA Strategy
- [Elastic Contributor Agreement](https://www.elastic.co/contributor-agreement)
- [GitLab DCO Transition](https://about.gitlab.com/press/releases/2017-11-01-gitlab-transitions-contributor-license/)
- [CLA Assistant Tool](https://cla-assistant.io/)

### Acquisition Examples
- [HashiCorp acquired by IBM ($6.4B)](https://www.hashicorp.com)
- [MongoDB IPO ($25B peak)](https://www.mongodb.com)
- [Elastic license change, 4x growth](https://www.elastic.co)
