---
sidebar_position: 3
---

# Activity Logging

The CLI captures and uploads activity logs to the Arfa platform.

## Flow

```
┌────────────┐     ┌────────────┐     ┌────────────┐
│   Event    │────▶│   Buffer   │────▶│   Upload   │
│  Capture   │     │ (100 items)│     │   Batch    │
└────────────┘     └────────────┘     └─────┬──────┘
                                            │
                                    ┌───────▼───────┐
                                    │   Success?    │
                                    └───────┬───────┘
                                      yes │   │ no
                                          │   │
                                    ┌─────▼─┐ │
                                    │ Done  │ │
                                    └───────┘ │
                                              │
                                    ┌─────────▼─────────┐
                                    │    Disk Queue     │
                                    │ ~/.arfa/log_queue │
                                    └───────────────────┘
```

## Event Types

| Event Type | Category | Description |
|------------|----------|-------------|
| `session_start` | session | CLI session began |
| `session_end` | session | CLI session ended |
| `tool_call` | classified | AI invoked a tool |
| `tool_result` | classified | Tool execution result |
| `permission_denied` | security | Tool call blocked |

## Batching

```go
type LoggerConfig struct {
    BatchSize     int           // 100 entries
    BatchInterval time.Duration // 5 seconds
    MaxRetries    int           // 5 attempts
}
```

Flush triggers:
1. Buffer reaches 100 entries
2. 5 second timer fires
3. Session ends

## Offline Queue

When API is unreachable:

```
~/.arfa/log_queue/
├── logs_1703347200_001.json
├── logs_1703347210_002.json
└── logs_1703347220_003.json
```

Background worker retries every 10 seconds.

## Troubleshooting

### Logs Not Appearing

1. Check disk queue:
   ```bash
   ls -la ~/.arfa/log_queue/
   ```

2. View queued content:
   ```bash
   head ~/.arfa/log_queue/*.json | jq .
   ```

3. Clear after fixing issues:
   ```bash
   rm ~/.arfa/log_queue/*.json
   ```
