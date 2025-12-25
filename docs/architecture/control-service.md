# Control Service Architecture

A robust, extensible architecture for intercepting, analyzing, and controlling LLM API traffic.

## Design Principles

1. **Pipeline Pattern** - Requests/responses flow through a chain of handlers
2. **Single Responsibility** - Each handler does one thing well
3. **Open/Closed** - Add new handlers without modifying existing ones
4. **Fail-Safe** - If a handler fails, traffic continues (configurable)

---

## Architecture Overview

```mermaid
flowchart TB
    subgraph CLI["CLI Process"]
        IC[Interactive Command]

        subgraph CS["Control Service"]
            direction TB

            subgraph Core["Core Layer"]
                Proxy[MITM Proxy<br/>goproxy]
                Router[Traffic Router]
                State[State Manager<br/>Sessions / Configs]
            end

            subgraph Pipeline["Handler Pipeline"]
                direction LR
                H1[Policy<br/>Handler]
                H2[Logger<br/>Handler]
                H3[PII<br/>Handler]
                H4[Analytics<br/>Handler]
                HN[Custom<br/>Handlers...]

                H1 --> H2 --> H3 --> H4 --> HN
            end

            subgraph Interfaces["Extension Interfaces"]
                IReq[RequestHandler]
                IRes[ResponseHandler]
                IBlock[BlockDecider]
                ITransform[Transformer]
            end
        end

        subgraph Agent["Agent Process"]
            A[Claude Code / Cursor / Windsurf]
        end
    end

    subgraph External["External"]
        LLM[LLM APIs]
        Platform[Ubik Platform API]
    end

    %% Traffic Flow
    IC --> State
    IC --> Proxy
    A -->|"1. Request"| Proxy
    Proxy --> Router
    Router -->|"2. Pipeline"| H1
    HN -->|"3. Forward/Block"| Router
    Router -->|"Allow"| LLM
    Router -.->|"Block"| A

    LLM -->|"4. Response"| Router
    Router -->|"5. Pipeline"| H1
    HN -->|"6. Return"| Router
    Router --> A

    %% Side Effects
    H2 -.->|"Async"| Platform
    State -.-> Pipeline
```

---

## Handler Pipeline Detail

```mermaid
flowchart LR
    subgraph Input
        REQ[Request/Response]
    end

    subgraph Pipeline["Handler Chain"]
        direction LR

        subgraph H1["Policy Handler"]
            P1[Check Rules]
            P2{Allow?}
            P1 --> P2
        end

        subgraph H2["Logger Handler"]
            L1[Capture Data]
            L2[Queue for Send]
            L1 --> L2
        end

        subgraph H3["PII Handler"]
            PII1[Detect PII]
            PII2[Redact/Flag]
            PII1 --> PII2
        end

        subgraph H4["Analytics Handler"]
            A1[Count Tokens]
            A2[Track Tools]
            A3[Update Metrics]
            A1 --> A2 --> A3
        end
    end

    subgraph Output
        OUT[Continue/Block]
    end

    REQ --> H1
    P2 -->|"Yes"| H2
    P2 -->|"No"| BLOCK[Block Response]
    H2 --> H3
    H3 --> H4
    H4 --> OUT

    BLOCK --> Output
```

---

## Handler Interface Design

```mermaid
classDiagram
    class Handler {
        <<interface>>
        +Name() string
        +Priority() int
        +HandleRequest(ctx, req) Result
        +HandleResponse(ctx, res) Result
    }

    class Result {
        +Action: Continue|Block|Transform
        +Data: any
        +Error: error
    }

    class HandlerContext {
        +SessionID: string
        +AgentID: string
        +Config: map
        +State: StateManager
        +Logger: Logger
    }

    class PolicyHandler {
        +rules: []Rule
        +HandleRequest() Result
        +HandleResponse() Result
    }

    class LoggerHandler {
        +buffer: []LogEntry
        +client: APIClient
        +HandleRequest() Result
        +HandleResponse() Result
    }

    class PIIHandler {
        +patterns: []Regex
        +redactor: Redactor
        +HandleRequest() Result
        +HandleResponse() Result
    }

    class AnalyticsHandler {
        +tokenCounter: TokenCounter
        +metrics: Metrics
        +HandleRequest() Result
        +HandleResponse() Result
    }

    Handler <|.. PolicyHandler
    Handler <|.. LoggerHandler
    Handler <|.. PIIHandler
    Handler <|.. AnalyticsHandler
    Handler --> Result
    Handler --> HandlerContext
```

---

## State Management

```mermaid
flowchart TB
    subgraph StateManager["State Manager"]
        Sessions[(Active Sessions)]
        Configs[(Handler Configs)]
        Policies[(Policy Rules)]
        Metrics[(Runtime Metrics)]
    end

    subgraph Sources["Configuration Sources"]
        Local[Local Config<br/>~/.ubik/config.json]
        Remote[Platform API<br/>/api/v1/config]
        Env[Environment<br/>Variables]
    end

    subgraph Consumers["State Consumers"]
        Pipeline[Handler Pipeline]
        Proxy[Proxy Server]
        API[API Client]
    end

    Sources --> StateManager
    StateManager --> Consumers

    Remote -.->|"Sync"| Configs
    Remote -.->|"Sync"| Policies
```

---

## Request Lifecycle

```mermaid
sequenceDiagram
    participant Agent
    participant Proxy
    participant Router
    participant Policy as Policy Handler
    participant Logger as Logger Handler
    participant PII as PII Handler
    participant LLM as LLM API
    participant Platform

    Agent->>Proxy: HTTPS Request
    Proxy->>Router: Intercept

    Router->>Policy: HandleRequest

    alt Blocked by Policy
        Policy-->>Router: Block(reason)
        Router-->>Agent: 403 Forbidden
    else Allowed
        Policy->>Logger: HandleRequest
        Logger->>Logger: Queue log entry
        Logger->>PII: HandleRequest
        PII->>PII: Scan & flag PII
        PII->>Router: Continue

        Router->>LLM: Forward Request
        LLM->>Router: Response

        Router->>Policy: HandleResponse
        Policy->>Logger: HandleResponse
        Logger->>Logger: Queue log entry
        Logger->>PII: HandleResponse
        PII->>PII: Redact PII if configured
        PII->>Router: Continue

        Router->>Agent: Response
    end

    Logger-->>Platform: Batch send (async)
```

---

## Extensibility: Adding New Handlers

```mermaid
flowchart TB
    subgraph NewHandler["Creating a New Handler"]
        Step1["1. Implement Handler Interface"]
        Step2["2. Register with Pipeline"]
        Step3["3. Configure Priority"]
        Step4["4. Deploy"]

        Step1 --> Step2 --> Step3 --> Step4
    end

    subgraph Example["Example: Rate Limit Handler"]
        Code["type RateLimitHandler struct {
    limits map[string]int
    windows map[string]time.Time
}

func (h *RateLimitHandler) HandleRequest(
    ctx HandlerContext,
    req *http.Request,
) Result {
    if h.isRateLimited(ctx.SessionID) {
        return Result{
            Action: Block,
            Data: 'Rate limit exceeded',
        }
    }
    h.increment(ctx.SessionID)
    return Result{Action: Continue}
}"]
    end

    subgraph Registry["Handler Registry"]
        R1[PolicyHandler - Priority: 100]
        R2[RateLimitHandler - Priority: 90]
        R3[LoggerHandler - Priority: 50]
        R4[PIIHandler - Priority: 40]
        R5[AnalyticsHandler - Priority: 10]
    end

    NewHandler --> Registry
```

---

## Future Extensions

```mermaid
mindmap
    root((Control Service))
        Policy
            Block by model
            Block by cost
            Block by content
            Allow/deny lists
        Logging
            Structured logs
            Real-time streaming
            Retention policies
        PII Detection
            Email/phone
            SSN/credit cards
            Custom patterns
            Auto-redaction
        Analytics
            Token usage
            Cost tracking
            Latency metrics
            Error rates
        Security
            Prompt injection detection
            Data exfiltration alerts
            Anomaly detection
        Compliance
            Audit trails
            Data residency
            GDPR/CCPA
        Integrations
            Webhooks
            SIEM export
            Custom plugins
```

---

## Directory Structure

```
services/cli/internal/control/
├── control.go              # Main Control Service
├── handler.go              # Handler interface
├── pipeline.go             # Pipeline orchestration
├── state.go                # State manager
├── router.go               # Traffic router
│
├── handlers/
│   ├── policy.go           # Policy enforcement
│   ├── logger.go           # Logging handler
│   ├── pii.go              # PII detection
│   ├── analytics.go        # Analytics collection
│   └── ratelimit.go        # Rate limiting
│
├── proxy/
│   └── proxy.go            # MITM proxy (existing)
│
└── config/
    └── config.go           # Configuration loading
```

---

## Key Benefits

| Aspect | Benefit |
|--------|---------|
| **Extensibility** | Add handlers without touching existing code |
| **Testability** | Each handler is independently testable |
| **Flexibility** | Enable/disable handlers via config |
| **Performance** | Async logging, non-blocking pipeline |
| **Maintainability** | Single responsibility per handler |
| **Future-proof** | Easy to add PII, rate limiting, etc. |

---

## Migration Path

```mermaid
flowchart LR
    subgraph Current["Current State"]
        C1[Proxy]
        C2[Logger]
        C3[Parser]
    end

    subgraph Phase1["Phase 1: Refactor"]
        P1[Control Service shell]
        P2[Move proxy inside]
        P3[Logger as handler]
    end

    subgraph Phase2["Phase 2: Pipeline"]
        P4[Add Handler interface]
        P5[Policy handler]
        P6[Pipeline orchestration]
    end

    subgraph Phase3["Phase 3: Extend"]
        P7[PII handler]
        P8[Analytics handler]
        P9[Custom handlers]
    end

    Current --> Phase1 --> Phase2 --> Phase3
```
