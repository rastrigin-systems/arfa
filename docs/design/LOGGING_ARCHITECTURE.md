# Logging Architecture Design Document

**Status:** Draft
**Version:** 1.0
**Last Updated:** 2025-12-10
**Author:** Ubik Engineering Team

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Current State](#current-state)
3. [MVP Phase 1: Human-Readable Log Classification](#mvp-phase-1-human-readable-log-classification) ⬅️ **START HERE**
4. [Business Requirements](#business-requirements)
5. [Proposed Architecture](#proposed-architecture)
6. [Log Classification System](#log-classification-system)
7. [PII Reduction Strategy](#pii-reduction-strategy)
8. [Compliance Framework](#compliance-framework)
9. [Implementation Phases](#implementation-phases)
10. [API Design](#api-design)
11. [Security Considerations](#security-considerations)
12. [Future Extensibility](#future-extensibility)

---

## Executive Summary

This document outlines the architecture for Ubik's enterprise logging system, designed to capture, classify, sanitize, and analyze AI agent interactions for compliance, security, and operational visibility.

**Key Goals:**
- Provide enterprises with complete visibility into AI agent usage
- Enable compliance with regulatory requirements (SOC 2, GDPR, HIPAA, etc.)
- Reduce PII exposure while maintaining audit trail integrity
- Create an extensible classification system for log analysis
- Support real-time monitoring and historical analysis

---

## Current State

### What We Have

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   CLI Client    │────▶│   MITM Proxy     │────▶│  AI Provider    │
│   (ubik sync)   │     │  (goproxy:8082)  │     │  (Anthropic,    │
│                 │◀────│                  │◀────│   OpenAI, etc.) │
└─────────────────┘     └──────────────────┘     └─────────────────┘
         │                      │
         ▼                      ▼
┌─────────────────────────────────────────────┐
│              Logging System                  │
│  • Session tracking (start/end)             │
│  • API request/response capture             │
│  • Basic header redaction                   │
│  • Offline queue with retry                 │
│  • Batch sending (100 entries / 5s)         │
└─────────────────────────────────────────────┘
         │
         ▼
┌─────────────────┐
│  Platform API   │
│  /api/v1/logs   │
└─────────────────┘
```

### Current Capabilities

| Feature | Status | Notes |
|---------|--------|-------|
| Request/Response Interception | ✅ Working | Via MITM proxy for Anthropic, OpenAI, Google |
| Session Tracking | ✅ Working | UUID-based session management |
| Header Redaction | ⚠️ Basic | Only Auth/API-Key/Token/Cookie headers |
| Body Redaction | ❌ Missing | Comment notes "should be more sophisticated" |
| Log Classification | ❌ Missing | All logs treated uniformly |
| PII Detection | ❌ Missing | No content scanning |
| Real-time Streaming | ✅ Working | WebSocket-based log streaming |
| Offline Queue | ✅ Working | Persists to ~/.ubik/log_queue/ |

### Current Log Structure

```go
type LogEntry struct {
    SessionID     string                 // UUID for session tracking
    AgentID       string                 // UUID for agent identification
    EventType     string                 // input|output|error|session_start|session_end|api_request|api_response
    EventCategory string                 // cli|session|proxy
    Content       string                 // Log message content
    Payload       map[string]interface{} // Structured metadata
    Timestamp     time.Time              // Event timestamp
}
```

---

## MVP Phase 1: Human-Readable Log Classification

> **This is the immediate next step.** Everything below this section is future roadmap.

### Goal

Transform raw API JSON logs into human-readable, classified entries that show:
- 💬 **User Prompts** - What the user asked
- 🤖 **AI Responses** - What the AI said (text only)
- 🔧 **Tool Calls** - Which tools the AI invoked
- 📤 **Tool Results** - What the tools returned
- ❌ **Errors** - Any failures

### Why This First?

| Current State | After MVP Phase 1 |
|---------------|-------------------|
| Raw JSON blob in logs | Structured, readable entries |
| Can't distinguish prompt from response | Clear classification |
| Hard to audit "what happened" | Easy timeline view |
| No business value from logs | Actionable insights |

### Input: Raw Anthropic API Format

We already capture this via MITM proxy:

```json
// Request to /v1/messages
{
  "model": "claude-sonnet-4-20250514",
  "max_tokens": 8096,
  "messages": [
    {"role": "user", "content": "Fix the bug in auth.go"},
    {"role": "assistant", "content": "I'll help fix that..."},
    {"role": "user", "content": "Thanks, now add tests"}
  ],
  "tools": [...]
}

// Response from /v1/messages
{
  "id": "msg_123",
  "type": "message",
  "role": "assistant",
  "content": [
    {"type": "text", "text": "I'll read the file first..."},
    {"type": "tool_use", "id": "tool_1", "name": "Read", "input": {"file_path": "/app/auth.go"}}
  ],
  "usage": {"input_tokens": 1500, "output_tokens": 200}
}
```

### Output: Classified Log Entries

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ SESSION: abc-123                                                             │
│ AGENT: Claude Code                                                           │
│ STARTED: 2024-01-15 10:30:00                                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│ [10:30:01] 💬 USER_PROMPT                                                   │
│ Fix the bug in auth.go                                                       │
│                                                                              │
│ [10:30:02] 🤖 AI_TEXT                                                       │
│ I'll read the file first to understand the issue...                         │
│                                                                              │
│ [10:30:02] 🔧 TOOL_CALL: Read                                               │
│ file_path: /app/auth.go                                                      │
│                                                                              │
│ [10:30:03] 📤 TOOL_RESULT: Read                                             │
│ [278 lines of code]                                                          │
│                                                                              │
│ [10:30:05] 🤖 AI_TEXT                                                       │
│ I found the issue on line 45. The token validation is missing...            │
│                                                                              │
│ [10:30:05] 🔧 TOOL_CALL: Edit                                               │
│ file_path: /app/auth.go                                                      │
│ old_string: "if token != nil {"                                              │
│ new_string: "if token != nil && token.Valid() {"                            │
│                                                                              │
│ [10:30:06] 📤 TOOL_RESULT: Edit                                             │
│ ✓ File updated successfully                                                  │
│                                                                              │
│ [10:30:06] 🤖 AI_TEXT                                                       │
│ I've fixed the bug. The issue was that we weren't checking token validity.  │
│                                                                              │
├─────────────────────────────────────────────────────────────────────────────┤
│ SESSION SUMMARY                                                              │
│ Duration: 6 seconds                                                          │
│ Tokens: 1,500 input / 450 output                                            │
│ Tool Calls: 2 (Read: 1, Edit: 1)                                            │
│ Cost Estimate: $0.003                                                        │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Implementation Plan

#### Step 1: Define Classification Types

```go
// pkg/types/log_classification.go

type ClassifiedLogEntry struct {
    // Identity
    ID            string    `json:"id"`
    SessionID     string    `json:"session_id"`
    Timestamp     time.Time `json:"timestamp"`

    // Classification
    EntryType     LogEntryType `json:"entry_type"`

    // Content (varies by type)
    Content       string            `json:"content,omitempty"`        // For text entries
    ToolName      string            `json:"tool_name,omitempty"`      // For tool calls
    ToolInput     map[string]any    `json:"tool_input,omitempty"`     // For tool calls
    ToolOutput    string            `json:"tool_output,omitempty"`    // For tool results
    ErrorMessage  string            `json:"error_message,omitempty"`  // For errors

    // Metadata
    TokensInput   int     `json:"tokens_input,omitempty"`
    TokensOutput  int     `json:"tokens_output,omitempty"`
    Model         string  `json:"model,omitempty"`
    Provider      string  `json:"provider,omitempty"`  // anthropic|openai|google
}

type LogEntryType string

const (
    LogTypeUserPrompt   LogEntryType = "user_prompt"
    LogTypeAIText       LogEntryType = "ai_text"
    LogTypeToolCall     LogEntryType = "tool_call"
    LogTypeToolResult   LogEntryType = "tool_result"
    LogTypeError        LogEntryType = "error"
    LogTypeSessionStart LogEntryType = "session_start"
    LogTypeSessionEnd   LogEntryType = "session_end"
)
```

#### Step 2: Anthropic JSON Parser

```go
// services/cli/internal/logparser/anthropic.go

type AnthropicParser struct{}

// ParseRequest parses an Anthropic /v1/messages request
func (p *AnthropicParser) ParseRequest(body []byte) ([]ClassifiedLogEntry, error) {
    var req struct {
        Model    string `json:"model"`
        Messages []struct {
            Role    string `json:"role"`
            Content any    `json:"content"` // string or []ContentBlock
        } `json:"messages"`
    }

    if err := json.Unmarshal(body, &req); err != nil {
        return nil, err
    }

    var entries []ClassifiedLogEntry

    for _, msg := range req.Messages {
        if msg.Role == "user" {
            entries = append(entries, ClassifiedLogEntry{
                EntryType: LogTypeUserPrompt,
                Content:   extractTextContent(msg.Content),
                Model:     req.Model,
                Provider:  "anthropic",
            })
        }
        // Note: assistant messages in request are conversation history
    }

    return entries, nil
}

// ParseResponse parses an Anthropic /v1/messages response
func (p *AnthropicParser) ParseResponse(body []byte) ([]ClassifiedLogEntry, error) {
    var resp struct {
        ID      string `json:"id"`
        Content []struct {
            Type  string          `json:"type"`
            Text  string          `json:"text,omitempty"`
            ID    string          `json:"id,omitempty"`
            Name  string          `json:"name,omitempty"`
            Input json.RawMessage `json:"input,omitempty"`
        } `json:"content"`
        Usage struct {
            InputTokens  int `json:"input_tokens"`
            OutputTokens int `json:"output_tokens"`
        } `json:"usage"`
    }

    if err := json.Unmarshal(body, &resp); err != nil {
        return nil, err
    }

    var entries []ClassifiedLogEntry

    for _, block := range resp.Content {
        switch block.Type {
        case "text":
            entries = append(entries, ClassifiedLogEntry{
                EntryType:    LogTypeAIText,
                Content:      block.Text,
                TokensInput:  resp.Usage.InputTokens,
                TokensOutput: resp.Usage.OutputTokens,
            })
        case "tool_use":
            var input map[string]any
            json.Unmarshal(block.Input, &input)
            entries = append(entries, ClassifiedLogEntry{
                EntryType: LogTypeToolCall,
                ToolName:  block.Name,
                ToolInput: input,
            })
        }
    }

    return entries, nil
}
```

#### Step 3: Integrate with MITM Proxy

```go
// services/cli/internal/httpproxy/server.go

func (s *ProxyServer) logRequest(r *http.Request) {
    // ... existing code ...

    // NEW: Parse and classify
    if strings.Contains(r.URL.Host, "anthropic.com") {
        entries, err := s.parser.ParseRequest(bodyBytes)
        if err == nil {
            for _, entry := range entries {
                entry.SessionID = s.sessionID
                entry.Timestamp = time.Now()
                s.logger.LogClassified(entry)
            }
        }
    }

    // Still log raw for debugging (optional)
    s.logger.LogEvent("api_request", "proxy", ...)
}
```

#### Step 4: Human-Readable Formatter

```go
// services/cli/internal/logparser/formatter.go

func FormatEntry(entry ClassifiedLogEntry) string {
    timestamp := entry.Timestamp.Format("15:04:05")

    switch entry.EntryType {
    case LogTypeUserPrompt:
        return fmt.Sprintf("[%s] 💬 USER_PROMPT\n%s\n", timestamp, entry.Content)

    case LogTypeAIText:
        return fmt.Sprintf("[%s] 🤖 AI_TEXT\n%s\n", timestamp, entry.Content)

    case LogTypeToolCall:
        inputStr := formatToolInput(entry.ToolInput)
        return fmt.Sprintf("[%s] 🔧 TOOL_CALL: %s\n%s\n", timestamp, entry.ToolName, inputStr)

    case LogTypeToolResult:
        return fmt.Sprintf("[%s] 📤 TOOL_RESULT: %s\n%s\n", timestamp, entry.ToolName, truncate(entry.ToolOutput, 200))

    case LogTypeError:
        return fmt.Sprintf("[%s] ❌ ERROR\n%s\n", timestamp, entry.ErrorMessage)

    default:
        return fmt.Sprintf("[%s] %s\n", timestamp, entry.EntryType)
    }
}
```

### Deliverables

- [ ] `ClassifiedLogEntry` type in `pkg/types/`
- [ ] `AnthropicParser` that extracts user prompts, AI text, tool calls
- [ ] `OpenAIParser` (same structure, different JSON)
- [ ] `Formatter` for human-readable output
- [ ] Integration with existing MITM proxy
- [ ] `ubik logs view --format=pretty` command
- [ ] Unit tests for parsers (80%+ coverage)

### What This Enables (But Doesn't Block)

| Future Feature | How This Helps | Still Independent |
|----------------|----------------|-------------------|
| PII Detection | Can scan classified content | Yes - add later |
| Compliance Reports | Have structured data | Yes - add later |
| Cost Analytics | Already tracking tokens | Yes - add later |
| Search | Can index by type | Yes - add later |

### Timeline

This is a focused, achievable scope:

1. **Day 1-2**: Define types, write parser tests
2. **Day 3-4**: Implement Anthropic parser
3. **Day 5**: Integrate with proxy, add formatter
4. **Day 6**: CLI command `ubik logs view`
5. **Day 7**: Testing, polish, documentation

---

## Business Requirements

### Who Needs These Logs?

| Stakeholder | Primary Needs | Key Questions They Ask |
|-------------|---------------|------------------------|
| **CISO / Security** | Audit trails, incident response | "What data did the AI access?" "Was there data exfiltration?" |
| **Compliance Officer** | Regulatory evidence, policy enforcement | "Can we prove GDPR compliance?" "Do we have audit logs for SOC 2?" |
| **Legal / Risk** | Liability documentation, IP protection | "Did the AI generate code that violates licenses?" "Is our IP being leaked?" |
| **IT Operations** | Usage monitoring, cost allocation | "Which teams use AI the most?" "What's our API spend?" |
| **Engineering Manager** | Productivity insights, best practices | "How are agents being used?" "What tasks are most common?" |
| **HR / Policy** | Policy compliance, acceptable use | "Is the AI being used appropriately?" "Any policy violations?" |

### Feature Requirements by Priority

#### P0 - MVP (Must Have)

1. **Complete Audit Trail**
   - Every AI interaction logged with timestamp, user, session
   - Immutable log storage (append-only)
   - Minimum 90-day retention

2. **Basic Log Classification**
   - Distinguish: tool calls, user prompts, AI responses, errors
   - Tag by AI provider (Anthropic, OpenAI, Google)
   - Session boundary markers

3. **PII Redaction**
   - Detect and redact common PII patterns
   - Configurable redaction rules per organization
   - Preserve log utility while reducing exposure

4. **Access Controls**
   - Role-based log access (admin, auditor, manager)
   - Organization-scoped data isolation
   - Audit log of who accessed logs

#### P1 - Post-MVP (Should Have)

5. **Advanced Classification (AI-Powered)**
   - Semantic classification of prompts/responses
   - Intent detection (coding, research, writing, etc.)
   - Sensitive topic detection (legal, financial, HR)

6. **Compliance Reporting**
   - Pre-built reports for SOC 2, GDPR, HIPAA
   - Automated compliance checks
   - Export in audit-friendly formats

7. **Real-time Alerts**
   - Policy violation detection
   - Anomaly detection (unusual usage patterns)
   - Webhook/email notifications

8. **Cost Analytics**
   - Token usage tracking per user/team/agent
   - Cost allocation and chargeback
   - Budget alerts

#### P2 - Future (Nice to Have)

9. **Content Analysis**
   - Code quality assessment
   - IP/license detection in generated code
   - Sentiment analysis

10. **Integrations**
    - SIEM integration (Splunk, Datadog, etc.)
    - Ticketing systems (Jira, ServiceNow)
    - Identity providers (SSO audit correlation)

---

## Proposed Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              EMPLOYEE WORKSTATION                                │
│  ┌──────────────┐     ┌──────────────────┐     ┌──────────────────────────────┐ │
│  │  AI Agent    │────▶│   Ubik CLI       │────▶│      MITM Proxy              │ │
│  │  (Claude,    │     │   (ubik sync)    │     │   • Request capture          │ │
│  │   Cursor)    │◀────│                  │◀────│   • Response capture         │ │
│  └──────────────┘     └──────────────────┘     │   • Pre-flight redaction     │ │
│                                                 └──────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ HTTPS (Batched Logs)
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              UBIK PLATFORM                                       │
│                                                                                  │
│  ┌────────────────────────────────────────────────────────────────────────────┐ │
│  │                         INGESTION LAYER                                     │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌────────────────┐  │ │
│  │  │  Log Receiver │  │  Validator   │  │  Enricher    │  │  Router        │  │ │
│  │  │  /api/v1/logs │──│  • Schema    │──│  • Org ID    │──│  • Classification│ │ │
│  │  │              │  │  • Auth      │  │  • Employee  │  │  • Priority     │  │ │
│  │  └──────────────┘  └──────────────┘  └──────────────┘  └────────────────┘  │ │
│  └────────────────────────────────────────────────────────────────────────────┘ │
│                                    │                                             │
│                     ┌──────────────┴──────────────┐                             │
│                     ▼                              ▼                             │
│  ┌────────────────────────────────┐  ┌────────────────────────────────────────┐ │
│  │     FAST PATH (Sync)           │  │     ASYNC PATH (Queue)                 │ │
│  │  • Audit log write             │  │  ┌──────────────────────────────────┐  │ │
│  │  • Real-time streaming         │  │  │         Message Queue            │  │ │
│  │  • Session state               │  │  │      (PostgreSQL/Redis)          │  │ │
│  └────────────────────────────────┘  │  └──────────────────────────────────┘  │ │
│                                       │                    │                    │ │
│                                       │     ┌──────────────┴──────────────┐    │ │
│                                       │     ▼                              ▼    │ │
│                                       │  ┌─────────────┐  ┌──────────────────┐ │ │
│                                       │  │ Classifier  │  │   PII Processor  │ │ │
│                                       │  │ • Rules     │  │   • Detection    │ │ │
│                                       │  │ • AI Model  │  │   • Redaction    │ │ │
│                                       │  └─────────────┘  └──────────────────┘ │ │
│                                       │         │                  │            │ │
│                                       │         └────────┬─────────┘            │ │
│                                       │                  ▼                      │ │
│                                       │  ┌──────────────────────────────────┐  │ │
│                                       │  │        Processed Log Store       │  │ │
│                                       │  │     (Searchable, Analyzed)       │  │ │
│                                       │  └──────────────────────────────────┘  │ │
│                                       └────────────────────────────────────────┘ │
│                                                                                  │
│  ┌────────────────────────────────────────────────────────────────────────────┐ │
│  │                         STORAGE LAYER                                       │ │
│  │  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────────────┐  │ │
│  │  │  Raw Log Store   │  │ Processed Store  │  │   Analytics Store        │  │ │
│  │  │  (Immutable)     │  │ (Searchable)     │  │   (Aggregated)           │  │ │
│  │  │  • Full content  │  │ • Classified     │  │   • Usage metrics        │  │ │
│  │  │  • Legal hold    │  │ • PII-redacted   │  │   • Cost data            │  │ │
│  │  │  • 7-year retain │  │ • Indexed        │  │   • Trends               │  │ │
│  │  └──────────────────┘  └──────────────────┘  └──────────────────────────┘  │ │
│  └────────────────────────────────────────────────────────────────────────────┘ │
│                                                                                  │
│  ┌────────────────────────────────────────────────────────────────────────────┐ │
│  │                         ACCESS LAYER                                        │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌────────────────┐  │ │
│  │  │  Admin UI    │  │  API         │  │  Streaming   │  │  Exports       │  │ │
│  │  │  • Search    │  │  • Query     │  │  • WebSocket │  │  • CSV/JSON    │  │ │
│  │  │  • Reports   │  │  • Filter    │  │  • Real-time │  │  • SIEM        │  │ │
│  │  │  • Alerts    │  │  • Paginate  │  │  • Subscribe │  │  • Compliance  │  │ │
│  │  └──────────────┘  └──────────────┘  └──────────────┘  └────────────────┘  │ │
│  └────────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Data Flow

```
1. CAPTURE        2. TRANSMIT       3. INGEST         4. PROCESS        5. SERVE
   ┌───────┐         ┌───────┐         ┌───────┐         ┌───────┐         ┌───────┐
   │ MITM  │────────▶│ Batch │────────▶│Receive│────────▶│Classify│───────▶│ Query │
   │ Proxy │         │ Queue │         │Validate│        │Redact  │        │ API   │
   └───────┘         └───────┘         └───────┘         └───────┘         └───────┘
       │                 │                 │                 │                 │
       ▼                 ▼                 ▼                 ▼                 ▼
   Raw capture      Offline-safe      Schema valid     Enriched +        Role-based
   with basic       batching +        + org-scoped     classified +      access +
   redaction        compression                        PII-safe          audit trail
```

---

## Log Classification System

### Classification Taxonomy

```
LogEntry
├── EventClass (L1)
│   ├── INTERACTION        # Human-AI conversation
│   ├── TOOL_USE           # Agent tool invocations
│   ├── SYSTEM             # Session/connection events
│   └── ERROR              # Failures and exceptions
│
├── EventType (L2)
│   ├── INTERACTION
│   │   ├── user_prompt           # User input to AI
│   │   ├── ai_response           # AI output to user
│   │   ├── context_injection     # System prompts, context
│   │   └── clarification         # Follow-up questions
│   │
│   ├── TOOL_USE
│   │   ├── file_read             # Reading files
│   │   ├── file_write            # Writing/editing files
│   │   ├── file_search           # Glob, grep operations
│   │   ├── command_execute       # Bash/shell commands
│   │   ├── web_fetch             # HTTP requests
│   │   ├── web_search            # Search queries
│   │   └── mcp_tool              # MCP server tools
│   │
│   ├── SYSTEM
│   │   ├── session_start         # New session began
│   │   ├── session_end           # Session terminated
│   │   ├── api_request           # Outbound API call
│   │   ├── api_response          # Inbound API response
│   │   └── config_change         # Settings modified
│   │
│   └── ERROR
│       ├── api_error             # Provider API errors
│       ├── tool_error            # Tool execution failures
│       ├── auth_error            # Authentication failures
│       └── rate_limit            # Rate limiting events
│
└── Metadata
    ├── intent                    # Detected intent (coding, research, etc.)
    ├── sensitivity               # low | medium | high | critical
    ├── pii_detected              # Boolean + types found
    ├── tokens_used               # Input/output token counts
    └── cost_estimate             # Estimated API cost
```

### Classification Pipeline

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     CLASSIFICATION PIPELINE                              │
│                                                                          │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐              │
│  │   Stage 1    │    │   Stage 2    │    │   Stage 3    │              │
│  │   RULES      │───▶│   PATTERNS   │───▶│   AI MODEL   │              │
│  │              │    │              │    │  (Optional)  │              │
│  └──────────────┘    └──────────────┘    └──────────────┘              │
│         │                   │                   │                       │
│         ▼                   ▼                   ▼                       │
│  • API endpoint        • Regex for        • Intent detection           │
│  • HTTP method           tool patterns    • Sensitivity scoring        │
│  • Event markers       • File paths       • Topic classification       │
│  • Session events      • Code patterns    • Anomaly detection          │
│                        • PII patterns                                   │
│                                                                          │
│  Cost: Free            Cost: Free         Cost: ~$0.001/log            │
│  Latency: <1ms         Latency: <5ms      Latency: <100ms              │
│  Accuracy: 70%         Accuracy: 85%      Accuracy: 95%+               │
└─────────────────────────────────────────────────────────────────────────┘
```

### Stage 1: Rule-Based Classification

```go
// Deterministic rules based on known patterns
type ClassificationRule struct {
    Condition   func(entry *LogEntry) bool
    EventClass  string
    EventType   string
    Priority    int  // Higher priority rules evaluated first
}

var rules = []ClassificationRule{
    // System events
    {
        Condition:  func(e *LogEntry) bool { return e.EventType == "session_start" },
        EventClass: "SYSTEM",
        EventType:  "session_start",
    },
    // API events
    {
        Condition:  func(e *LogEntry) bool {
            return strings.Contains(e.Payload["url"].(string), "/v1/messages")
        },
        EventClass: "INTERACTION",
        EventType:  "api_request",
    },
    // Tool use patterns
    {
        Condition:  func(e *LogEntry) bool {
            content := e.Content
            return strings.Contains(content, "Read tool") ||
                   strings.Contains(content, "file_path")
        },
        EventClass: "TOOL_USE",
        EventType:  "file_read",
    },
}
```

### Stage 2: Pattern-Based Classification

```go
// Regex patterns for deeper classification
var patterns = map[string]*regexp.Regexp{
    // Tool invocation patterns
    "bash_command":    regexp.MustCompile(`(?i)bash|shell|terminal|command`),
    "file_operation":  regexp.MustCompile(`(?i)read|write|edit|create|delete.*file`),
    "search_operation": regexp.MustCompile(`(?i)grep|find|search|glob`),

    // Content sensitivity patterns
    "code_generation": regexp.MustCompile(`(?i)function|class|def |import |package `),
    "database_query":  regexp.MustCompile(`(?i)SELECT|INSERT|UPDATE|DELETE|CREATE TABLE`),
    "api_key_mention": regexp.MustCompile(`(?i)api.?key|secret|password|credential`),

    // PII patterns (see PII section for full list)
    "email":           regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
    "phone":           regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`),
    "ssn":             regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
}
```

### Stage 3: AI-Powered Classification (Optional)

For organizations requiring deeper analysis, we offer AI-powered classification using a small, cost-efficient model.

**Model Selection Criteria:**
- Low latency (<100ms)
- Low cost (<$0.001 per classification)
- Good accuracy for classification tasks
- Can run on-premise if required

**Candidate Models:**
| Model | Cost/1K tokens | Latency | Accuracy | Notes |
|-------|---------------|---------|----------|-------|
| Claude Haiku | $0.00025 | ~50ms | High | Best balance |
| GPT-4o-mini | $0.00015 | ~40ms | High | Cost-effective |
| Local LLaMA | $0 (infra) | ~100ms | Medium | Privacy-first |

**Classification Prompt:**

```
Classify this AI agent log entry. Return JSON only.

Entry:
{log_content}

Classify into:
1. intent: coding | research | writing | analysis | other
2. sensitivity: low | medium | high | critical
3. topics: [list of relevant topics]
4. risk_flags: [any compliance/security concerns]

JSON response:
```

---

## PII Reduction Strategy

### Critical Constraint: PII Must Never Leave the Machine

> **Key Insight:** We cannot use cloud AI models (Claude Haiku, GPT-4o-mini) for PII detection because that would require sending PII over the network to detect PII - defeating the entire purpose.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     THE PII DETECTION PARADOX                                │
│                                                                              │
│   ❌ WRONG: Send to cloud AI for detection                                  │
│   ┌──────────────┐     ┌──────────────┐                                     │
│   │ "My SSN is   │────▶│ Claude Haiku │  ← PII already leaked!              │
│   │  123-45-6789"│     │ "detect PII" │                                     │
│   └──────────────┘     └──────────────┘                                     │
│                                                                              │
│   ✅ CORRECT: Local-only detection                                          │
│   ┌──────────────┐     ┌──────────────┐     ┌──────────────┐               │
│   │ "My SSN is   │────▶│ Local Regex  │────▶│ "[REDACTED]" │───▶ LLM       │
│   │  123-45-6789"│     │ + Local NER  │     │              │               │
│   └──────────────┘     └──────────────┘     └──────────────┘               │
│                         NEVER LEAVES MACHINE                                 │
└─────────────────────────────────────────────────────────────────────────────┘
```

### PII Detection Approaches Comparison

| Approach | Latency | Network? | Accuracy | MVP Ready |
|----------|---------|----------|----------|-----------|
| **Regex patterns** | <1ms | ❌ No | 70% (structured PII) | ✅ Yes |
| **Local NER (prose/spaCy)** | 5-50ms | ❌ No | 85% | ✅ Yes |
| **Dictionary/Bloom filter** | 1-5ms | ❌ No | Medium | ✅ Yes |
| **Cloud AI (Haiku)** | 50-200ms | ⚠️ Yes - leaks PII! | 95% | ❌ No |
| **Self-hosted LLM** | 50ms | ❌ No (within org) | 90% | ⚠️ Complex |

### Recommended: Tiered Local Detection

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     TIERED LOCAL PII DETECTION                               │
│                                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                   │
│  │   TIER 1     │    │   TIER 2     │    │   TIER 3     │                   │
│  │   REGEX      │───▶│  LOCAL NER   │───▶│  LEARN MODE  │                   │
│  │   (Always)   │    │  (Optional)  │    │   (Async)    │                   │
│  └──────────────┘    └──────────────┘    └──────────────┘                   │
│        │                   │                    │                            │
│        ▼                   ▼                    ▼                            │
│   ~0.5ms              ~5-20ms              Background                        │
│   100% of requests    Configurable         (on redacted data)               │
│                                                                              │
│  Catches:             Catches:             Purpose:                          │
│  • SSN, CC, Phone     • Person names       • Discover new patterns          │
│  • Email, IP          • Locations          • Improve rules                  │
│  • API Keys           • Organizations      • Can use AI safely              │
│  • Structured IDs     • Medical terms      • (already redacted)             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### PII Detection Categories

| Category | Examples | Detection Method | Default Action |
|----------|----------|------------------|----------------|
| **Direct Identifiers** | SSN, Passport, Driver's License | Regex + Checksum | REDACT |
| **Contact Info** | Email, Phone, Address | Regex | REDACT |
| **Financial** | Credit Card, Bank Account, Tax ID | Regex + Luhn | REDACT |
| **Health** | Medical Record #, Insurance ID | Regex | REDACT |
| **Credentials** | Passwords, API Keys, Tokens | Regex + Entropy | REDACT |
| **Names** | Person names in context | NER Model | MASK (configurable) |
| **Locations** | Specific addresses | Regex + NER | MASK (configurable) |
| **Custom** | Org-specific patterns | Admin-defined regex | Configurable |

### PII Detection Patterns

```go
var PIIPatterns = map[string]PIIPattern{
    // Direct Identifiers
    "ssn": {
        Pattern:     regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
        Category:    "direct_identifier",
        Sensitivity: "critical",
        Action:      "redact",
        Replacement: "[SSN-REDACTED]",
    },
    "passport_us": {
        Pattern:     regexp.MustCompile(`\b[A-Z]\d{8}\b`),
        Category:    "direct_identifier",
        Sensitivity: "critical",
        Action:      "redact",
    },

    // Contact Information
    "email": {
        Pattern:     regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
        Category:    "contact",
        Sensitivity: "high",
        Action:      "redact",
        Replacement: "[EMAIL-REDACTED]",
    },
    "phone_us": {
        Pattern:     regexp.MustCompile(`\b(?:\+1[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}\b`),
        Category:    "contact",
        Sensitivity: "high",
        Action:      "redact",
        Replacement: "[PHONE-REDACTED]",
    },
    "address": {
        Pattern:     regexp.MustCompile(`\b\d+\s+[\w\s]+(?:Street|St|Avenue|Ave|Road|Rd|Boulevard|Blvd|Drive|Dr|Lane|Ln|Court|Ct)\b`),
        Category:    "contact",
        Sensitivity: "medium",
        Action:      "mask",
    },

    // Financial
    "credit_card": {
        Pattern:     regexp.MustCompile(`\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|6(?:011|5[0-9]{2})[0-9]{12})\b`),
        Category:    "financial",
        Sensitivity: "critical",
        Action:      "redact",
        Validator:   luhnCheck,  // Additional validation
        Replacement: "[CC-REDACTED]",
    },
    "bank_account": {
        Pattern:     regexp.MustCompile(`\b[0-9]{8,17}\b`),  // Context-aware
        Category:    "financial",
        Sensitivity: "critical",
        Action:      "redact",
        ContextRequired: true,  // Only match in financial context
    },

    // Credentials
    "api_key_generic": {
        Pattern:     regexp.MustCompile(`(?i)(?:api[_-]?key|apikey|secret[_-]?key|access[_-]?token)[\s:=]+['\"]?([a-zA-Z0-9_-]{20,})['\"]?`),
        Category:    "credential",
        Sensitivity: "critical",
        Action:      "redact",
        Replacement: "[API-KEY-REDACTED]",
    },
    "aws_key": {
        Pattern:     regexp.MustCompile(`(?:AKIA|ABIA|ACCA|ASIA)[0-9A-Z]{16}`),
        Category:    "credential",
        Sensitivity: "critical",
        Action:      "redact",
        Replacement: "[AWS-KEY-REDACTED]",
    },
    "private_key": {
        Pattern:     regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`),
        Category:    "credential",
        Sensitivity: "critical",
        Action:      "redact",
        Replacement: "[PRIVATE-KEY-REDACTED]",
    },
    "jwt_token": {
        Pattern:     regexp.MustCompile(`eyJ[a-zA-Z0-9_-]*\.eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*`),
        Category:    "credential",
        Sensitivity: "high",
        Action:      "redact",
        Replacement: "[JWT-REDACTED]",
    },
}
```

### Multi-Layer Redaction Strategy

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    PII REDACTION PIPELINE                                │
│                                                                          │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐              │
│  │   Layer 1    │    │   Layer 2    │    │   Layer 3    │              │
│  │   CLIENT     │───▶│   INGESTION  │───▶│   STORAGE    │              │
│  │   (Pre-send) │    │  (On-receive)│    │ (Background) │              │
│  └──────────────┘    └──────────────┘    └──────────────┘              │
│         │                   │                   │                       │
│         ▼                   ▼                   ▼                       │
│  • API key headers     • Fast regex        • Deep NER scan             │
│  • Known secrets         patterns          • ML-based detection        │
│  • Env variables       • Credential        • Cross-reference           │
│  • Obvious patterns      formats           • Audit trail               │
│                        • Critical PII                                   │
│                                                                          │
│  Location: CLI         Location: API       Location: Worker             │
│  Sync: Yes             Sync: Yes           Sync: No (async)             │
│  Latency: <1ms         Latency: <10ms      Latency: Minutes             │
└─────────────────────────────────────────────────────────────────────────┘
```

### Redaction Actions

| Action | Description | Use Case |
|--------|-------------|----------|
| **REDACT** | Replace with placeholder | Sensitive data that must be removed |
| **MASK** | Partial replacement (show last 4) | Data needed for debugging |
| **HASH** | One-way hash for correlation | Track patterns without exposing data |
| **ENCRYPT** | Reversible for authorized users | Audit/legal requirements |
| **FLAG** | Mark but don't modify | Alert without disrupting |

### Organization-Configurable Rules

```yaml
# Organization PII Configuration
pii_config:
  # Global settings
  default_action: redact

  # Category overrides
  categories:
    contact:
      email:
        action: hash  # Allow correlation without exposure
    financial:
      action: redact
      alert: true     # Notify security team
    credential:
      action: redact
      alert: true
      block: true     # Prevent log from being stored

  # Custom patterns for this organization
  custom_patterns:
    - name: employee_id
      pattern: "EMP-[0-9]{6}"
      action: mask
      sensitivity: medium
    - name: internal_project_code
      pattern: "PRJ-[A-Z]{3}-[0-9]{4}"
      action: hash
      sensitivity: low

  # Allowlist (never redact)
  allowlist:
    - "support@company.com"  # Public support email
    - "api.company.com"      # Public API domain
```

---

## Compliance Framework

### Supported Compliance Standards

| Standard | Requirements | Ubik Features |
|----------|--------------|---------------|
| **SOC 2 Type II** | Audit trails, access controls, encryption | Complete logging, RBAC, encryption at rest/transit |
| **GDPR** | Data minimization, right to erasure, consent | PII redaction, data retention policies, export |
| **HIPAA** | PHI protection, audit trails, access controls | Healthcare PII detection, BAA support, audit logs |
| **CCPA** | Consumer data rights, disclosure | Data inventory, export, deletion workflows |
| **ISO 27001** | Information security management | Comprehensive controls, risk management |
| **FedRAMP** | Government security requirements | Encryption, access controls, audit (future) |

### Compliance Features Matrix

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         COMPLIANCE FEATURE MATRIX                                │
├─────────────────────────┬─────────┬─────────┬─────────┬─────────┬──────────────┤
│ Feature                 │ SOC 2   │ GDPR    │ HIPAA   │ CCPA    │ ISO 27001    │
├─────────────────────────┼─────────┼─────────┼─────────┼─────────┼──────────────┤
│ Immutable Audit Logs    │   ✓     │   ✓     │   ✓     │   ✓     │     ✓        │
│ Access Control (RBAC)   │   ✓     │   ✓     │   ✓     │   ✓     │     ✓        │
│ Encryption at Rest      │   ✓     │   ✓     │   ✓     │   ✓     │     ✓        │
│ Encryption in Transit   │   ✓     │   ✓     │   ✓     │   ✓     │     ✓        │
│ PII Detection/Redaction │   ○     │   ✓     │   ✓     │   ✓     │     ○        │
│ Data Retention Policies │   ✓     │   ✓     │   ✓     │   ✓     │     ✓        │
│ Right to Erasure        │   ○     │   ✓     │   ○     │   ✓     │     ○        │
│ Data Export             │   ○     │   ✓     │   ✓     │   ✓     │     ○        │
│ Consent Management      │   ○     │   ✓     │   ○     │   ✓     │     ○        │
│ Breach Notification     │   ✓     │   ✓     │   ✓     │   ✓     │     ✓        │
│ Vendor Management       │   ✓     │   ✓     │   ✓     │   ○     │     ✓        │
├─────────────────────────┴─────────┴─────────┴─────────┴─────────┴──────────────┤
│ ✓ = Required   ○ = Recommended                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Pre-Built Compliance Reports

```go
type ComplianceReport struct {
    Standard       string    // SOC2, GDPR, HIPAA, etc.
    Period         TimeRange
    GeneratedAt    time.Time
    GeneratedBy    string

    Sections       []ReportSection
    Findings       []Finding
    Recommendations []Recommendation

    // Attestation
    SignedHash     string
    Verifiable     bool
}

// Example: SOC 2 Report Sections
var SOC2Sections = []string{
    "CC1: Control Environment",
    "CC2: Communication and Information",
    "CC3: Risk Assessment",
    "CC4: Monitoring Activities",
    "CC5: Control Activities",
    "CC6: Logical and Physical Access Controls",
    "CC7: System Operations",
    "CC8: Change Management",
    "CC9: Risk Mitigation",
}
```

### Data Retention Policies

```yaml
retention_policies:
  # Raw logs (immutable, for legal/compliance)
  raw_logs:
    default: 7 years
    legal_hold: indefinite
    storage: cold_storage  # Cost-optimized
    encryption: AES-256
    access: legal_team_only

  # Processed logs (searchable, day-to-day use)
  processed_logs:
    default: 90 days
    extended: 1 year  # Optional add-on
    storage: hot_storage
    access: role_based

  # Analytics/aggregates (anonymized)
  analytics:
    default: 2 years
    storage: warm_storage
    anonymized: true
    access: all_admins

  # Per-regulation overrides
  gdpr:
    max_retention: 3 years  # Unless legal basis
    erasure_sla: 30 days
  hipaa:
    min_retention: 6 years
    audit_logs: 6 years
```

---

## Implementation Phases

### Phase 1: Human-Readable Log Classification (NOW) ⬅️ START HERE

**Goal:** Parse API JSON into readable, classified log entries

**See [MVP Phase 1](#mvp-phase-1-human-readable-log-classification) for detailed implementation plan.**

```
Day 1-2: Types & Tests
├── [ ] ClassifiedLogEntry type in pkg/types/
├── [ ] Unit tests for Anthropic parser
└── [ ] Unit tests for formatter

Day 3-4: Anthropic Parser
├── [ ] Parse /v1/messages requests (extract user prompts)
├── [ ] Parse /v1/messages responses (extract AI text, tool calls)
├── [ ] Handle streaming responses (SSE)
└── [ ] Error handling

Day 5: Integration
├── [ ] Integrate parser with MITM proxy
├── [ ] LogClassified() method in logger
└── [ ] Store classified entries

Day 6: CLI Command
├── [ ] `ubik logs view --format=pretty`
├── [ ] `ubik logs view --format=json`
└── [ ] Session summary display

Day 7: Polish
├── [ ] Integration tests
├── [ ] Documentation
└── [ ] 80%+ test coverage
```

**Deliverables:**
- [ ] `ClassifiedLogEntry` type
- [ ] Anthropic JSON parser
- [ ] Human-readable formatter
- [ ] `ubik logs view` command
- [ ] 80% test coverage

---

### Phase 2: Basic PII Handling (Next)

**Goal:** Local-only PII detection and redaction

```
├── [ ] PII regex pattern library (20+ patterns)
├── [ ] Client-side pre-redaction in MITM proxy
├── [ ] Configurable redaction rules
├── [ ] Unit tests for all patterns
└── [ ] Optional: Local NER integration
```

---

### Phase 3: Platform Integration

**Goal:** Send classified logs to platform, view in admin UI

```
├── [ ] API endpoint for classified logs
├── [ ] Database schema for classified entries
├── [ ] Admin UI log viewer
├── [ ] Search and filter functionality
└── [ ] Real-time streaming updates
```

---

### Phase 4: Compliance & Analytics (Future)

**Goal:** Compliance reporting and usage analytics

```
├── [ ] SOC 2 report template
├── [ ] Token usage tracking & cost allocation
├── [ ] Usage dashboards
└── [ ] Full-text search & export
```

---

### Phase 5: AI-Powered Features (Future)

**Goal:** Intelligent classification and anomaly detection (using local models or self-hosted)

```
├── [ ] Local NER for name/location detection
├── [ ] Anomaly detection for unusual patterns
├── [ ] Self-hosted LLM for advanced classification (optional)
└── [ ] Pattern learning from redacted data
```

---

### Phase 6: Enterprise Features (Future)

**Goal:** Enterprise-grade integrations and scale

```
├── [ ] SIEM Integration (Splunk, Datadog)
├── [ ] HIPAA/FedRAMP compliance
├── [ ] Multi-region support
└── [ ] High availability
```

---

## API Design

### Enhanced Log Schema

```go
// Enhanced LogEntry for MVP
type LogEntry struct {
    // Identity
    ID            string    `json:"id"`             // UUID
    SessionID     string    `json:"session_id"`     // Session UUID
    AgentID       string    `json:"agent_id"`       // Agent UUID
    EmployeeID    string    `json:"employee_id"`    // Employee UUID
    OrganizationID string   `json:"organization_id"` // Org UUID (multi-tenancy)

    // Classification
    EventClass    string    `json:"event_class"`    // INTERACTION|TOOL_USE|SYSTEM|ERROR
    EventType     string    `json:"event_type"`     // Specific event type
    EventCategory string    `json:"event_category"` // cli|session|proxy|tool

    // Content
    Content       string    `json:"content"`        // Log message (may be redacted)
    ContentHash   string    `json:"content_hash"`   // Hash of original content
    Payload       Payload   `json:"payload"`        // Structured metadata

    // PII Handling
    PIIDetected   bool      `json:"pii_detected"`   // Was PII found?
    PIITypes      []string  `json:"pii_types"`      // Types of PII found
    RedactionCount int      `json:"redaction_count"` // Number of redactions

    // Analysis (populated async)
    Intent        string    `json:"intent,omitempty"`       // Detected intent
    Sensitivity   string    `json:"sensitivity,omitempty"`  // low|medium|high|critical
    Topics        []string  `json:"topics,omitempty"`       // Relevant topics
    RiskScore     float64   `json:"risk_score,omitempty"`   // 0.0-1.0

    // Metrics
    TokensInput   int       `json:"tokens_input,omitempty"`  // Input tokens
    TokensOutput  int       `json:"tokens_output,omitempty"` // Output tokens
    CostEstimate  float64   `json:"cost_estimate,omitempty"` // USD
    LatencyMs     int       `json:"latency_ms,omitempty"`    // Response time

    // Timestamps
    Timestamp     time.Time `json:"timestamp"`      // Event time
    ProcessedAt   time.Time `json:"processed_at"`   // Processing time

    // Audit
    ClientVersion string    `json:"client_version"` // CLI version
    ClientIP      string    `json:"client_ip"`      // Redacted/hashed
}

type Payload struct {
    // Request details (for api_request)
    Method      string            `json:"method,omitempty"`
    URL         string            `json:"url,omitempty"`
    Headers     map[string]string `json:"headers,omitempty"`
    Body        string            `json:"body,omitempty"`       // May be truncated
    BodyHash    string            `json:"body_hash,omitempty"`  // Full body hash

    // Response details (for api_response)
    StatusCode  int               `json:"status_code,omitempty"`

    // Tool details (for tool_use)
    ToolName    string            `json:"tool_name,omitempty"`
    ToolInput   map[string]any    `json:"tool_input,omitempty"`
    ToolOutput  string            `json:"tool_output,omitempty"`

    // Error details (for errors)
    ErrorCode   string            `json:"error_code,omitempty"`
    ErrorMessage string           `json:"error_message,omitempty"`
    StackTrace  string            `json:"stack_trace,omitempty"`

    // Session details
    Provider    string            `json:"provider,omitempty"`   // anthropic|openai|google
    Model       string            `json:"model,omitempty"`      // claude-3-opus, etc.

    // Custom metadata
    Custom      map[string]any    `json:"custom,omitempty"`
}
```

### New API Endpoints

```yaml
# Log Ingestion
POST /api/v1/logs
POST /api/v1/logs/batch

# Log Query
GET  /api/v1/logs
GET  /api/v1/logs/{id}
GET  /api/v1/logs/sessions/{session_id}
POST /api/v1/logs/search          # Advanced search

# Log Streaming
GET  /api/v1/logs/stream          # WebSocket (existing)

# Analytics
GET  /api/v1/logs/analytics/usage
GET  /api/v1/logs/analytics/costs
GET  /api/v1/logs/analytics/trends

# Compliance
GET  /api/v1/logs/compliance/report
GET  /api/v1/logs/compliance/audit-trail
POST /api/v1/logs/compliance/export

# Configuration
GET  /api/v1/logs/config/pii-rules
PUT  /api/v1/logs/config/pii-rules
GET  /api/v1/logs/config/retention
PUT  /api/v1/logs/config/retention
```

### Search API

```yaml
POST /api/v1/logs/search
Content-Type: application/json

{
  "query": {
    "text": "database query",           # Full-text search
    "event_class": ["TOOL_USE"],        # Filter by class
    "event_type": ["file_read", "command_execute"],
    "sensitivity": ["high", "critical"],
    "pii_detected": true,
    "date_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-01-31T23:59:59Z"
    },
    "employee_ids": ["uuid1", "uuid2"],
    "agent_ids": ["uuid3"],
    "session_ids": ["uuid4"]
  },
  "sort": {
    "field": "timestamp",
    "order": "desc"
  },
  "pagination": {
    "page": 1,
    "per_page": 50
  },
  "include": ["payload", "analysis"]    # Fields to include
}
```

---

## Security Considerations

### Threat Model

| Threat | Impact | Mitigation |
|--------|--------|------------|
| Log injection | Data corruption, XSS | Input sanitization, schema validation |
| PII leakage | Privacy breach, fines | Multi-layer redaction, encryption |
| Unauthorized access | Data breach | RBAC, audit logging, encryption |
| Log tampering | Compliance failure | Immutable storage, checksums |
| Denial of service | Availability loss | Rate limiting, quotas |
| Insider threat | Data exfiltration | Audit trails, access monitoring |

### Security Controls

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         SECURITY LAYERS                                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │  TRANSPORT SECURITY                                                 │ │
│  │  • TLS 1.3 for all connections                                     │ │
│  │  • Certificate pinning (optional)                                  │ │
│  │  • mTLS for service-to-service (future)                           │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │  AUTHENTICATION & AUTHORIZATION                                     │ │
│  │  • JWT tokens with short expiry                                    │ │
│  │  • Role-based access control (RBAC)                                │ │
│  │  • Organization-scoped data isolation                              │ │
│  │  • Audit log viewer role                                           │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │  DATA PROTECTION                                                    │ │
│  │  • Encryption at rest (AES-256)                                    │ │
│  │  • PII redaction before storage                                    │ │
│  │  • Field-level encryption for sensitive data                       │ │
│  │  • Secure key management (KMS)                                     │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │  INTEGRITY                                                          │ │
│  │  • Immutable log storage                                           │ │
│  │  • Content hashing (SHA-256)                                       │ │
│  │  • Chain of custody tracking                                       │ │
│  │  • Tamper detection                                                │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │  MONITORING & RESPONSE                                              │ │
│  │  • Real-time security alerts                                       │ │
│  │  • Anomaly detection                                               │ │
│  │  • Incident response playbooks                                     │ │
│  │  • Regular security audits                                         │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Access Control Model

```go
type LogAccessRole string

const (
    RoleAdmin       LogAccessRole = "admin"       // Full access
    RoleAuditor     LogAccessRole = "auditor"     // Read-only, all logs
    RoleManager     LogAccessRole = "manager"     // Team logs only
    RoleCompliance  LogAccessRole = "compliance"  // Compliance reports
    RoleSecurity    LogAccessRole = "security"    // Security events
    RoleAnalytics   LogAccessRole = "analytics"   // Aggregated only
)

type LogPermission struct {
    Role            LogAccessRole
    CanViewRawLogs  bool
    CanViewPII      bool      // Can see unredacted data
    CanExport       bool
    CanConfigurePII bool
    ScopeFilter     string    // e.g., "team_id = X" or "org_id = Y"
}

var RolePermissions = map[LogAccessRole]LogPermission{
    RoleAdmin: {
        CanViewRawLogs:  true,
        CanViewPII:      false,  // Even admins can't see PII by default
        CanExport:       true,
        CanConfigurePII: true,
        ScopeFilter:     "org_id = :org_id",
    },
    RoleAuditor: {
        CanViewRawLogs:  true,
        CanViewPII:      true,   // For compliance investigations
        CanExport:       true,
        CanConfigurePII: false,
        ScopeFilter:     "org_id = :org_id",
    },
    RoleManager: {
        CanViewRawLogs:  true,
        CanViewPII:      false,
        CanExport:       false,
        CanConfigurePII: false,
        ScopeFilter:     "team_id IN (:managed_teams)",
    },
    // ... etc
}
```

---

## Future Extensibility

### Plugin Architecture

```go
// Classifier plugin interface
type ClassifierPlugin interface {
    Name() string
    Version() string
    Classify(entry *LogEntry) (*Classification, error)
    Priority() int  // Execution order
}

// PII detector plugin interface
type PIIDetectorPlugin interface {
    Name() string
    Patterns() []PIIPattern
    Detect(content string) []PIIMatch
    Redact(content string, matches []PIIMatch) string
}

// Export plugin interface
type ExportPlugin interface {
    Name() string
    Format() string  // json, csv, cef, leef, etc.
    Export(logs []LogEntry, config ExportConfig) ([]byte, error)
}

// Alert plugin interface
type AlertPlugin interface {
    Name() string
    Channels() []string  // email, slack, webhook, etc.
    ShouldAlert(entry *LogEntry) bool
    Send(alert Alert) error
}
```

### Integration Points

```yaml
integrations:
  siem:
    - name: splunk
      type: hec  # HTTP Event Collector
      config:
        endpoint: https://splunk.company.com:8088
        token: ${SPLUNK_HEC_TOKEN}
        index: ubik_logs

    - name: datadog
      type: api
      config:
        api_key: ${DATADOG_API_KEY}
        site: datadoghq.com

  ticketing:
    - name: jira
      type: webhook
      config:
        url: https://company.atlassian.net/rest/api/3
        auth: basic
        project: SEC
        trigger_on: [security_alert, policy_violation]

  identity:
    - name: okta
      type: scim
      config:
        domain: company.okta.com
        correlation: employee_email

  webhooks:
    - name: custom_alerts
      url: https://alerts.company.com/ubik
      events: [high_sensitivity, pii_detected, anomaly]
      auth: bearer
      retry: 3
```

### Extensibility Roadmap

```
MVP (Now)
├── Core classification engine
├── Plugin interface definitions
└── Basic PII patterns

v1.1
├── Custom PII patterns via API
├── Webhook integrations
└── CSV/JSON export

v1.2
├── AI classifier plugin
├── Splunk integration
└── Advanced search

v2.0
├── Plugin marketplace
├── Custom classifier upload
├── Multi-region support
└── Real-time ML pipeline
```

---

## Appendix

### A. Event Type Reference

| Event Type | Class | Description | Example |
|------------|-------|-------------|---------|
| `session_start` | SYSTEM | New session initiated | User started ubik sync |
| `session_end` | SYSTEM | Session terminated | User exited or timeout |
| `api_request` | SYSTEM | Outbound API call | POST /v1/messages |
| `api_response` | SYSTEM | API response received | 200 OK with tokens |
| `user_prompt` | INTERACTION | User input to AI | "Fix the bug in auth.go" |
| `ai_response` | INTERACTION | AI output to user | Generated code |
| `file_read` | TOOL_USE | File read operation | Read auth.go |
| `file_write` | TOOL_USE | File write/edit | Edit auth.go |
| `command_execute` | TOOL_USE | Shell command | Run `go test` |
| `web_fetch` | TOOL_USE | HTTP fetch | Fetch documentation |
| `api_error` | ERROR | Provider error | Rate limit exceeded |
| `tool_error` | ERROR | Tool failure | Permission denied |

### B. PII Pattern Library

See [PII Patterns Documentation](./PII_PATTERNS.md) for complete pattern library.

### C. Compliance Checklist

See [Compliance Checklist](./COMPLIANCE_CHECKLIST.md) for detailed requirements per standard.

### D. Performance Benchmarks

| Operation | Target Latency | Throughput |
|-----------|---------------|------------|
| Log ingestion | <50ms p99 | 10,000/sec |
| Classification (rules) | <5ms | 50,000/sec |
| Classification (AI) | <100ms | 1,000/sec |
| PII detection | <10ms | 20,000/sec |
| Search query | <500ms p99 | 100/sec |
| Export (1M logs) | <60s | N/A |

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-12-10 | Engineering | Initial draft |

---

## Open Questions

1. **AI Model Selection:** Should we use Claude Haiku (best quality) or GPT-4o-mini (lowest cost) for classification?

2. **PII Storage:** Should we store original content in encrypted form for legal holds, or only store redacted versions?

3. **Real-time vs Batch:** For AI classification, should we process in real-time (higher cost, lower latency) or batch (lower cost, delayed insights)?

4. **On-Premise Option:** Do customers need on-premise deployment for the classification model (data sovereignty)?

5. **Retention Defaults:** What should be the default retention period? 90 days? 1 year?

---

*This is a living document. Please submit feedback and suggestions via GitHub issues.*
