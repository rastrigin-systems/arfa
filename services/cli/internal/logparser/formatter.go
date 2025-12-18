package logparser

import (
	"fmt"
	"strings"
	"time"

	"github.com/rastrigin-systems/ubik-enterprise/pkg/types"
)

// Formatter formats classified log entries for display
type Formatter struct {
	// UseEmoji enables emoji prefixes (disable for non-unicode terminals)
	UseEmoji bool
	// ShowTimestamp includes timestamp in output
	ShowTimestamp bool
	// MaxContentLength truncates content longer than this (0 = no limit)
	MaxContentLength int
	// IndentToolInput indents tool input for readability
	IndentToolInput bool
}

// DefaultFormatter returns a formatter with sensible defaults
func DefaultFormatter() *Formatter {
	return &Formatter{
		UseEmoji:         true,
		ShowTimestamp:    true,
		MaxContentLength: 500,
		IndentToolInput:  true,
	}
}

// Format formats a single log entry for display
func (f *Formatter) Format(entry types.ClassifiedLogEntry) string {
	var sb strings.Builder

	// Timestamp
	if f.ShowTimestamp {
		ts := entry.Timestamp
		if ts.IsZero() {
			ts = time.Now()
		}
		sb.WriteString(fmt.Sprintf("[%s] ", ts.Format("15:04:05")))
	}

	// Entry type with optional emoji
	sb.WriteString(f.formatEntryType(entry.EntryType))

	// Tool name for tool-related entries
	if entry.ToolName != "" {
		sb.WriteString(fmt.Sprintf(": %s", entry.ToolName))
	}

	sb.WriteString("\n")

	// Content based on entry type
	switch entry.EntryType {
	case types.LogTypeUserPrompt, types.LogTypeAIText:
		content := f.truncate(entry.Content)
		sb.WriteString(content)
		sb.WriteString("\n")

	case types.LogTypeToolCall:
		if f.IndentToolInput && len(entry.ToolInput) > 0 {
			sb.WriteString(f.formatToolInput(entry.ToolInput))
		}

	case types.LogTypeToolResult:
		output := f.truncate(entry.ToolOutput)
		if output == "" {
			sb.WriteString("(empty result)\n")
		} else {
			sb.WriteString(output)
			sb.WriteString("\n")
		}

	case types.LogTypeError:
		if entry.ErrorCode != "" {
			sb.WriteString(fmt.Sprintf("[%s] ", entry.ErrorCode))
		}
		sb.WriteString(entry.ErrorMessage)
		sb.WriteString("\n")

	case types.LogTypeSessionStart:
		sb.WriteString(fmt.Sprintf("Session started: %s\n", entry.SessionID))

	case types.LogTypeSessionEnd:
		sb.WriteString(fmt.Sprintf("Session ended: %s\n", entry.SessionID))
	}

	return sb.String()
}

// FormatSession formats all entries in a session with a header and summary
func (f *Formatter) FormatSession(sessionID string, entries []types.ClassifiedLogEntry) string {
	var sb strings.Builder

	// Header
	sb.WriteString("┌─────────────────────────────────────────────────────────────────────────────┐\n")
	sb.WriteString(fmt.Sprintf("│ SESSION: %-67s │\n", truncateString(sessionID, 67)))

	// Find metadata from first entry
	if len(entries) > 0 {
		first := entries[0]
		if first.Model != "" {
			sb.WriteString(fmt.Sprintf("│ MODEL: %-69s │\n", first.Model))
		}
		if first.Provider != "" {
			sb.WriteString(fmt.Sprintf("│ PROVIDER: %-66s │\n", first.Provider))
		}
	}

	sb.WriteString("├─────────────────────────────────────────────────────────────────────────────┤\n")
	sb.WriteString("│                                                                             │\n")

	// Format each entry
	for _, entry := range entries {
		formatted := f.Format(entry)
		// Indent each line
		for _, line := range strings.Split(formatted, "\n") {
			if line != "" {
				sb.WriteString(fmt.Sprintf("│ %-75s │\n", truncateString(line, 75)))
			}
		}
		sb.WriteString("│                                                                             │\n")
	}

	// Summary
	summary := f.calculateSummary(sessionID, entries)
	sb.WriteString("├─────────────────────────────────────────────────────────────────────────────┤\n")
	sb.WriteString("│ SESSION SUMMARY                                                             │\n")
	sb.WriteString(fmt.Sprintf("│ Tokens: %d input / %d output%-42s │\n",
		summary.TokensInput, summary.TokensOutput, ""))
	sb.WriteString(fmt.Sprintf("│ Tool Calls: %-64d │\n", summary.ToolCalls))
	if summary.CostEstimate > 0 {
		sb.WriteString(fmt.Sprintf("│ Cost Estimate: $%-60.4f │\n", summary.CostEstimate))
	}
	sb.WriteString("└─────────────────────────────────────────────────────────────────────────────┘\n")

	return sb.String()
}

// formatEntryType returns the display string for an entry type
func (f *Formatter) formatEntryType(t types.LogEntryType) string {
	if f.UseEmoji {
		switch t {
		case types.LogTypeUserPrompt:
			return "💬 USER_PROMPT"
		case types.LogTypeAIText:
			return "🤖 AI_TEXT"
		case types.LogTypeToolCall:
			return "🔧 TOOL_CALL"
		case types.LogTypeToolResult:
			return "📤 TOOL_RESULT"
		case types.LogTypeError:
			return "❌ ERROR"
		case types.LogTypeSessionStart:
			return "▶️ SESSION_START"
		case types.LogTypeSessionEnd:
			return "⏹️ SESSION_END"
		default:
			return string(t)
		}
	}

	// Without emoji
	switch t {
	case types.LogTypeUserPrompt:
		return "[USER_PROMPT]"
	case types.LogTypeAIText:
		return "[AI_TEXT]"
	case types.LogTypeToolCall:
		return "[TOOL_CALL]"
	case types.LogTypeToolResult:
		return "[TOOL_RESULT]"
	case types.LogTypeError:
		return "[ERROR]"
	case types.LogTypeSessionStart:
		return "[SESSION_START]"
	case types.LogTypeSessionEnd:
		return "[SESSION_END]"
	default:
		return fmt.Sprintf("[%s]", t)
	}
}

// formatToolInput formats tool input parameters
func (f *Formatter) formatToolInput(input map[string]any) string {
	var sb strings.Builder
	for key, value := range input {
		strValue := fmt.Sprintf("%v", value)
		strValue = f.truncate(strValue)
		sb.WriteString(fmt.Sprintf("  %s: %s\n", key, strValue))
	}
	return sb.String()
}

// truncate truncates a string if it exceeds MaxContentLength
func (f *Formatter) truncate(s string) string {
	if f.MaxContentLength <= 0 {
		return s
	}
	return truncateString(s, f.MaxContentLength)
}

// calculateSummary computes aggregate statistics for a session
func (f *Formatter) calculateSummary(sessionID string, entries []types.ClassifiedLogEntry) types.SessionSummary {
	summary := types.SessionSummary{
		SessionID:   sessionID,
		ToolsByName: make(map[string]int),
	}

	for _, entry := range entries {
		// Track max tokens (they're cumulative in responses)
		if entry.TokensInput > summary.TokensInput {
			summary.TokensInput = entry.TokensInput
		}
		if entry.TokensOutput > summary.TokensOutput {
			summary.TokensOutput = entry.TokensOutput
		}

		// Count tool calls
		if entry.EntryType == types.LogTypeToolCall {
			summary.ToolCalls++
			summary.ToolsByName[entry.ToolName]++
		}

		// Count errors
		if entry.EntryType == types.LogTypeError {
			summary.Errors++
		}

		// Track model/provider
		if entry.Model != "" && summary.Model == "" {
			summary.Model = entry.Model
		}
		if entry.Provider != "" {
			summary.Provider = entry.Provider
		}

		// Track timestamps
		if !entry.Timestamp.IsZero() {
			if summary.StartTime.IsZero() || entry.Timestamp.Before(summary.StartTime) {
				summary.StartTime = entry.Timestamp
			}
			if summary.EndTime == nil || entry.Timestamp.After(*summary.EndTime) {
				t := entry.Timestamp
				summary.EndTime = &t
			}
		}
	}

	// Calculate duration
	if !summary.StartTime.IsZero() && summary.EndTime != nil {
		summary.Duration = summary.EndTime.Sub(summary.StartTime)
	}

	// Estimate cost (rough estimates based on Claude pricing)
	summary.CostEstimate = estimateCost(summary.Model, summary.TokensInput, summary.TokensOutput)

	return summary
}

// truncateString truncates a string to maxLen, adding "..." if truncated
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// estimateCost returns estimated cost in USD based on model and tokens
func estimateCost(model string, inputTokens, outputTokens int) float64 {
	// Pricing per 1M tokens (as of 2024)
	var inputPrice, outputPrice float64

	switch {
	case strings.Contains(model, "opus"):
		inputPrice = 15.0  // $15 per 1M input
		outputPrice = 75.0 // $75 per 1M output
	case strings.Contains(model, "sonnet"):
		inputPrice = 3.0   // $3 per 1M input
		outputPrice = 15.0 // $15 per 1M output
	case strings.Contains(model, "haiku"):
		inputPrice = 0.25  // $0.25 per 1M input
		outputPrice = 1.25 // $1.25 per 1M output
	default:
		// Default to sonnet pricing
		inputPrice = 3.0
		outputPrice = 15.0
	}

	inputCost := (float64(inputTokens) / 1_000_000) * inputPrice
	outputCost := (float64(outputTokens) / 1_000_000) * outputPrice

	return inputCost + outputCost
}
