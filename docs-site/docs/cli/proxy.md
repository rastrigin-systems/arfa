---
sidebar_position: 2
---

# Transparent HTTPS Proxy

The CLI includes a MITM proxy for intercepting LLM API traffic.

## How It Works

```
┌─────────────┐     ┌─────────────────┐     ┌─────────────┐
│ AI Client   │────▶│ Arfa Proxy      │────▶│ LLM APIs    │
│ (Claude,    │     │ (localhost:8082)│     │ (Anthropic) │
│  Cursor)    │     └────────┬────────┘     └─────────────┘
└─────────────┘              │
                     Capture, Analyze, Enforce
```

## Setup

### 1. CA Certificate

On first run, the CLI generates a self-signed CA:

```
~/.arfa/certs/
├── arfa-ca.pem      # CA certificate
└── arfa-ca-key.pem  # CA private key
```

### 2. Client Configuration

```bash
export HTTPS_PROXY=http://localhost:8082
export HTTP_PROXY=http://localhost:8082
export NODE_EXTRA_CA_CERTS=~/.arfa/certs/arfa-ca.pem
```

Or use the helper:

```bash
eval $(arfa proxy env)
```

## Supported Clients

| Client | Detection | Status |
|--------|-----------|--------|
| Claude Code | User-Agent: `claude-code/*` | ✅ Full support |
| Cursor | User-Agent: `cursor/*` | ✅ Full support |
| Windsurf | User-Agent: `windsurf/*` | ✅ Full support |
| Aider | User-Agent: `aider/*` | ✅ Full support |

## SSE Stream Parsing

LLM APIs use Server-Sent Events for streaming:

```
data: {"type":"content_block_delta","delta":{"type":"tool_use",...}}
data: {"type":"message_stop"}
```

The proxy parses these in real-time to extract tool calls, token usage, and errors.

## Security

- CA private key stored with 0600 permissions
- Only inspects traffic to known LLM endpoints
- Non-LLM traffic passed through unchanged
