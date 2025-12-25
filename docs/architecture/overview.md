# AI Agent Security Gateway

**Vision**: Enterprise-grade visibility and control for AI coding agents, integrated with existing security infrastructure.

---

## Executive Summary

Arfa provides the missing security layer for AI coding agents. As enterprises adopt Claude Code, Cursor, and Copilot, security teams have zero visibility into what these agents actually do. Arfa captures every tool invocation, enforces policies, and forwards structured events to existing SIEM systems.

**One-liner**: "See and control every tool your AI agents use. Export to your existing SIEM."

---

## Market Problem

| Challenge | Impact |
|-----------|--------|
| AI agents execute code autonomously | Security teams can't audit actions |
| No tool-level visibility | "What did Claude do?" → "No idea" |
| Existing SIEMs blind to AI | Datadog/Splunk see nothing |
| Compliance gaps | SOC2/HIPAA require audit trails |
| Shadow AI usage | Employees use agents without oversight |

---

## Solution Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    ENTERPRISE SIEM                              │
│              (Kibana / Splunk / Datadog / etc.)                 │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Dashboard: AI Agent Activity                           │    │
│  │  • Tool calls per hour                                  │    │
│  │  • Blocked actions by policy                            │    │
│  │  • Token usage by team                                  │    │
│  │  • Anomaly alerts                                       │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
                              ▲
                              │ OpenTelemetry / Webhook / Kafka
                              │
┌─────────────────────────────────────────────────────────────────┐
│                     ARFA SECURITY GATEWAY                       │
│                                                                 │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────────┐    │
│  │    CAPTURE    │  │    ENFORCE    │  │     FORWARD       │    │
│  │               │  │               │  │                   │    │
│  │ • Tool calls  │  │ • Block       │  │ • Webhook         │    │
│  │ • Parameters  │  │ • Audit-only  │  │ • Kafka           │    │
│  │ • Results     │  │ • Conditional │  │ • OpenTelemetry   │    │
│  │ • Token usage │  │ • Alert       │  │ • S3/GCS          │    │
│  │ • Session ctx │  │ • Approve     │  │ • Syslog          │    │
│  └───────────────┘  └───────────────┘  └───────────────────┘    │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    POLICY ENGINE                        │    │
│  │  • Per-org policies    • Conditional rules              │    │
│  │  • Per-team overrides  • Time-based policies            │    │
│  │  • Per-employee        • Approval workflows             │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
                              ▲
                              │ HTTPS Proxy (transparent)
                              │
┌─────────────────────────────────────────────────────────────────┐
│                       AI AGENTS                                 │
│                                                                 │
│     Claude Code    Cursor    Windsurf    GitHub Copilot         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Core Features

### 1. Tool Call Capture (Current: MVP Complete)

**Status**: ✅ Implemented

What we capture for every tool invocation:

```json
{
  "event_type": "tool_call",
  "timestamp": "2025-01-15T10:30:00Z",
  "employee_id": "ae848cb1-7c8a-41eb-b164-bd176dd934e4",
  "org_id": "8b58e482-737e-4145-b0e8-69162a6b5db1",
  "session_id": "c704df8e-0126-4814-a07f-334de83c017f",
  "agent_id": "claude-code",
  "payload": {
    "tool_name": "Bash",
    "tool_id": "toolu_01ABC123",
    "tool_input": {
      "command": "rm -rf /tmp/test"
    },
    "blocked": true,
    "block_reason": "Destructive commands blocked by policy"
  }
}
```

**Unique data points**:
- Exact tool name (Bash, Read, Write, Edit, Glob, Grep, etc.)
- Full input parameters
- Block status and reason
- Employee/team attribution
- Session context for replay

---

### 2. Policy Engine (Current: Basic Blocking)

**Status**: 🟡 Partial (unconditional + conditional blocking works)

#### Current Capabilities
- [x] Block tools by name (exact match)
- [x] Block tools by glob pattern (`mcp__*`)
- [x] Conditional blocking (parameter-based rules)
- [x] Audit-only mode (log without blocking)

#### Needed for Enterprise
- [ ] Policy inheritance (org → team → employee)
- [ ] Time-based policies (block after hours)
- [ ] Approval workflows (request access to blocked tool)
- [ ] Policy versioning and rollback
- [ ] Policy testing/dry-run mode

#### Policy Examples

```yaml
# Block destructive commands
- tool: Bash
  action: block
  conditions:
    - param: command
      operator: matches
      value: "rm -rf|mkfs|dd if="
  reason: "Destructive commands require approval"

# Audit all file writes (don't block, just log)
- tool: Write
  action: audit
  alert: slack

# Block external network access after hours
- tool: Bash
  action: block
  conditions:
    - param: command
      operator: contains
      value: "curl|wget|nc"
  schedule:
    deny: "18:00-09:00"
  reason: "External network access blocked outside business hours"
```

---

### 3. SIEM Integration (Current: Not Started)

**Status**: ❌ Not implemented

**Priority**: HIGH - This is the enterprise differentiator

#### Integration Options

| Method | Use Case | Complexity |
|--------|----------|------------|
| **Webhook** | Real-time, any endpoint | Low |
| **Kafka** | High-volume, streaming | Medium |
| **OpenTelemetry** | Standard observability | Medium |
| **S3/GCS** | Batch, compliance archive | Low |
| **Syslog** | Legacy SIEM integration | Low |

#### Webhook Integration (Priority 1)

```yaml
# Admin configuration
destinations:
  - name: security-siem
    type: webhook
    url: https://siem.company.com/api/events
    headers:
      Authorization: "Bearer ${SIEM_TOKEN}"
    events:
      - tool_call
      - policy_violation
    format: json
    retry:
      max_attempts: 3
      backoff: exponential
```

#### Event Schema (OpenTelemetry-compatible)

```json
{
  "resourceLogs": [{
    "resource": {
      "attributes": [
        {"key": "service.name", "value": {"stringValue": "arfa-gateway"}},
        {"key": "org.id", "value": {"stringValue": "8b58e482-..."}},
        {"key": "employee.id", "value": {"stringValue": "ae848cb1-..."}}
      ]
    },
    "scopeLogs": [{
      "logRecords": [{
        "timeUnixNano": "1705312200000000000",
        "severityText": "INFO",
        "body": {"stringValue": "tool_call:Bash blocked"},
        "attributes": [
          {"key": "tool.name", "value": {"stringValue": "Bash"}},
          {"key": "tool.blocked", "value": {"boolValue": true}},
          {"key": "policy.rule", "value": {"stringValue": "no-destructive-commands"}}
        ]
      }]
    }]
  }]
}
```

---

### 4. Alerting (Current: Not Started)

**Status**: ❌ Not implemented

Real-time alerts for security events:

| Event | Alert Channel | Example |
|-------|---------------|---------|
| Policy violation | Slack/PagerDuty | "Employee X attempted blocked command" |
| Anomaly detected | Email/Webhook | "Unusual tool usage pattern" |
| New tool first use | Slack | "Employee X used Bash for first time" |
| High token usage | Email | "Team Y exceeded daily token budget" |

#### Alert Configuration

```yaml
alerts:
  - name: policy-violation
    trigger:
      event: tool_call
      condition: blocked == true
    channels:
      - slack: "#security-alerts"
      - pagerduty: P1
    throttle: 5m  # Max 1 alert per 5 minutes per employee

  - name: destructive-command-attempt
    trigger:
      event: tool_call
      condition: |
        tool_name == "Bash" &&
        tool_input.command matches "rm -rf|drop table|truncate"
    channels:
      - slack: "#security-critical"
      - email: security@company.com
    severity: critical
```

---

### 5. CLI (Current: Minimal)

**Status**: 🟡 Basic (employee-focused, limited management)

**Design principle:** No "admin" namespace. Commands check user role at runtime.

#### Current Commands
```bash
arfa login/logout       # Authentication
arfa sync              # Sync configs
arfa logs view/stream  # View own logs
arfa policies list     # View policies
```

#### Needed Commands (Permission-Based)
```bash
# Webhook destinations (admin/manager only for write ops)
arfa webhooks list
arfa webhooks add -f webhook.yaml
arfa webhooks test <name>
arfa webhooks delete <name>

# Policy management (admin/manager only for write ops)
arfa policies list
arfa policies create -f policy.yaml
arfa policies test --dry-run
arfa policies enable/disable <id>

# Employee management (admin only)
arfa employees list
arfa employees logs <email>
arfa employees revoke <email>

# Audit (admin/manager only)
arfa audit export --since 30d --format csv
arfa audit report --type compliance
```

**Permission model:**
| Role | Read (list) | Write (add/delete) |
|------|-------------|-------------------|
| admin | ✅ | ✅ |
| manager | ✅ | ✅ |
| developer | ✅ | ❌ |

---

### 6. Web Dashboard (Current: Basic)

**Status**: 🟡 Basic logs page exists

#### Current
- [x] View logs (flat list)
- [x] Basic filtering

#### Needed for Enterprise
- [ ] Real-time dashboard with metrics
- [ ] Policy management UI
- [ ] Destination configuration UI
- [ ] Employee activity overview
- [ ] Anomaly visualization
- [ ] Compliance reports

**Note**: Dashboard is secondary. Enterprises will use their SIEM. Our dashboard is for:
1. Initial setup/configuration
2. Quick debugging
3. Companies without existing SIEM

---

## Competitive Differentiation

| Feature | Arfa | Datadog | Snyk | Lakera |
|---------|------|---------|------|--------|
| AI tool-level visibility | ✅ | ❌ | ❌ | ❌ |
| Real-time policy blocking | ✅ | ❌ | ❌ | 🟡 (input only) |
| SIEM integration | ✅ | N/A | ❌ | ❌ |
| Multi-agent support | ✅ | ❌ | ❌ | ❌ |
| On-prem option | 🔜 | ❌ | ❌ | ❌ |

**Our unique position**: We sit between AI agents and LLM APIs, capturing data no one else can see.

---

## Implementation Roadmap

### Phase 1: Foundation (Current)
**Goal**: Prove core capture and blocking works

- [x] HTTPS proxy intercepts LLM traffic
- [x] Tool call extraction from SSE streams
- [x] Basic policy blocking (unconditional)
- [x] Conditional policy blocking
- [x] Log storage in PostgreSQL
- [x] Basic CLI for employees

### Phase 2: Enterprise Integration (Next)
**Goal**: Connect to enterprise security infrastructure

- [ ] Webhook destination support
- [ ] Configurable event forwarding
- [ ] Admin CLI commands
- [ ] Slack/PagerDuty alerting
- [ ] OpenTelemetry export format
- [ ] Kibana dashboard template

### Phase 3: Advanced Policies
**Goal**: Sophisticated policy engine

- [ ] Policy inheritance (org → team → employee)
- [ ] Time-based policies
- [ ] Approval workflows
- [ ] Policy versioning
- [ ] Anomaly detection rules

### Phase 4: Scale & Compliance
**Goal**: Enterprise-ready deployment

- [ ] High-availability deployment
- [ ] On-premises option
- [ ] SOC2 compliance documentation
- [ ] SSO/SAML integration
- [ ] Audit log retention policies
- [ ] Data residency options

---

## Success Metrics

### Technical
- [ ] <100ms latency overhead from proxy
- [ ] 99.9% uptime for proxy service
- [ ] <5s event delivery to SIEM

### Business
- [ ] 3 enterprise pilots with SIEM integration
- [ ] Security team approval (not just dev team)
- [ ] Compliance checkbox for AI agent usage

---

## Demo Script

**"Watch what happens when Claude tries to run a dangerous command"**

1. Show Kibana dashboard (empty)
2. Employee runs: `arfa` → asks Claude to "clean up temp files"
3. Claude attempts: `rm -rf /tmp/*`
4. Arfa blocks it, shows policy message to employee
5. Kibana dashboard updates in real-time:
   - Event: `tool_call:Bash BLOCKED`
   - Employee: `sarah.cto@acme.com`
   - Reason: `Destructive commands blocked`
6. Slack alert fires: "Policy violation detected"

**Time**: 30 seconds. **Impact**: Visceral understanding of value.

---

## Business Strategy

### What We Sell

**Three value propositions:**

| Value | Buyer Pain | Our Solution |
|-------|------------|--------------|
| **Visibility** | "What is AI doing?" | Tool calls, parameters, attribution |
| **Control** | "Stop bad things" | Policy blocking, approvals, alerts |
| **Compliance** | "Prove it to auditors" | Audit trail, SIEM export, reports |

### Buyer Personas

| Persona | Pain Point | What They Buy |
|---------|------------|---------------|
| **CISO** | "I can't audit AI usage" | Compliance + SIEM integration |
| **Security Engineer** | "I need to see what AI does" | Visibility + Alerts |
| **Engineering Manager** | "Devs use AI without guardrails" | Control + Policies |
| **Compliance Officer** | "SOC2 requires audit trails" | Export + Reports |

### Revenue Model

**Per-Seat SaaS (Recommended)**

| Tier | Price | Includes |
|------|-------|----------|
| Starter | $15/user/month | 5 users, basic policies |
| Team | $30/user/month | Unlimited users, SIEM export |
| Enterprise | Custom | On-prem, SSO, SLA, support |

---

## Company Roadmap

### Year 1: Prove Value (Current → +12 months)

**Goal**: 10 paying enterprise customers

```
Q1: Foundation                          ← CURRENT
├── Webhook export to SIEM
├── Kibana dashboard template
├── 3 pilot customers
└── Basic alerting (Slack)

Q2: Enterprise Ready
├── SSO/SAML integration
├── Policy management UI
├── On-prem deployment option
└── SOC2 Type 1

Q3: Scale
├── Multi-agent support (Cursor, Copilot)
├── Advanced policies (time-based, approvals)
├── 10 paying customers
└── Series A fundraise

Q4: Expand
├── Anomaly detection
├── Cost management features
├── SOC2 Type 2
└── Partner integrations (ServiceNow, Jira)
```

### Year 2: Market Leadership

- AI Agent marketplace (curated, secure agents)
- Industry compliance packs (HIPAA, PCI, FedRAMP)
- Developer SDK (embed Arfa in custom agents)

### Year 3+: Platform Evolution

- "Arfa Runtime" - Secure execution environment
- Multi-agent orchestration with guardrails
- Expand beyond coding agents

---

## Competitive Moat Evolution

| Timeline | Moat |
|----------|------|
| **Today** | Proxy captures unique tool-level data |
| **Year 1** | Enterprise integrations + policy library + customer lock-in |
| **Year 2+** | Network effects + platform + largest AI behavior dataset |

---

## Exit Scenarios

| Acquirer | Strategic Rationale |
|----------|---------------------|
| **Datadog** | Add AI observability to platform |
| **CrowdStrike** | Endpoint security + AI security |
| **Palo Alto** | Expand security portfolio |
| **Microsoft** | Secure Copilot ecosystem |
| **Anthropic/OpenAI** | Enterprise trust layer |

**IPO path**: $50M+ ARR, category leader in AI security

---

## Immediate Priorities

| # | Action | Why | Status |
|---|--------|-----|--------|
| 1 | **Webhook export** | Prove SIEM integration story | Not started |
| 2 | **Kibana template** | Tangible demo artifact | Not started |
| 3 | **Slack alerting** | Real-time policy violation alerts | Not started |
| 4 | **3 pilot customers** | Validate with real enterprises | Not started |

---

*Last updated: 2025-12-22*
