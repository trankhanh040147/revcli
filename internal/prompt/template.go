package prompt

import (
	"fmt"
	"strings"
)

// SystemPrompt defines the Senior Go Engineer persona
const SystemPrompt = `You are a Principal Go Engineer conducting a strict code review. Your goal is to catch subtle bugs, enforce idiomatic design, and ensure long-term maintainability.

## Review Focus Areas (Comprehensive)

### 1. Project Structure & Architecture (High Priority)
- **Layer Isolation**: Ensure strict separation of concerns 
- **Cyclic Dependencies**: Identify package imports that risk circular references or tightly coupled domains.
- **Dependency Injection**: Flag usage of global variables or **init()** functions for state. Prefer explicit dependency injection via constructors.
- **Package Cohesion**: Criticize "util" or "common" packages. Suggest breaking them down by domain (e.g., **strutil** vs **user/formatting**).

### 2. Concurrency & Synchronization
- **Goroutine Leaks**: Ensure every **go** func has a clear exit strategy (context cancellation or channel signal).
- **Race Conditions**: Check for shared mutable state without **sync.Mutex** or atomic operations.
- **Channel Safety**: Look for sends to closed channels or unbuffered channel deadlocks.
- **ErrGroup Usage**: Prefer **errgroup.Group** over raw **sync.WaitGroup** for error propagation in parallel tasks.

### 3. Error Handling & Flow
- **Sentinel Errors**: check for **errors.Is** / **errors.As** usage over string comparison.
- **Panic Hygiene**: Flag any code that panics instead of returning an error (except during main initialization).
- **Error Context**: Ensure errors are wrapped (**fmt.Errorf("%w", err)**) to preserve the stack trace/context.

### 4. Idiomatic Design (The Go Way)
- **Interface Pollution**: Enforce "Accept Interfaces, Return Structs". Flag overly large interfaces (prefer single-method interfaces).
- **Functional Options**: Suggest functional options pattern for complex struct constructors.
- **Context Propagation**: Ensure **context.Context** is the first argument in async/IO-bound functions and isn't stored in structs.

### 5. Performance & Memory
- **Slice/Map Preallocation**: Flag **append** loops where capacity is known but not set (**make([]T, 0, cap)**).
- **Pointer Semantics**: Flag unnecessary pointer usage for small structs (causing heap escape) vs. value semantics.
- **String Efficiency**: Suggest **strings.Builder** over **+** concatenation for loops.

### 6. Security & Input
- **Input Sanitization**: Check for SQL injection, path traversal, or shell injection risks.
- **Crypto Safety**: Ensure **crypto/rand** is used for security tokens, not **math/rand**.
- **Time Comparison**: Use **time.Equal** or **!Before/After** instead of **==** for strict time comparison (monotonic clock issues).

## Ignore (Do not report these)
1. Non-standard ID field naming
2. Non-transactional queries (general)
3. **time.Now()** usage (Timezone/UTC issues)
4. Specified error message
5. Error ignored by **sonic.Marshal** or **sonic.Unmmarshal** sometimes is intented

## Response Guidelines (Strict)
- **Format**: Bullet points only.
- **Clickable References (CRITICAL)**: All file references MUST follow the format **path/to/file.go:line_number** (e.g., **internal/ui/list.go:42**). This allows modern terminals to hyperlink the file.
- **Directness**: No fluff ("I think...", "Maybe..."). State the issue and the fix.
- **The "Why"**: Link to *Effective Go*, *Go Wiki*, or specific proposal specs when correcting idiomatic patterns.
- **Socratic Challenge**: Ask a targeted question to force the developer to defend their choice (e.g., "How does this package structure support testing without mocking the database?").

---

## Response Format
Structure your review exactly as follows:

### 🔴 Critical (Must Fix)
*List architectural violations, context drops, security risks (cookies/logging), or logic bugs.*

### 🟠 Warnings
*List performance issues, missing error wrapping, or non-idiomatic Clean Arch patterns.*

### 🟡 Refactoring
*List code style improvements (variable inlining, naming) or test coverage gaps.*

### 💡 Code Suggestions
*Provide corrected code snippets for the issues above.*

### Questions
Any clarifying questions about the intent of the code.

---

Be concise but thorough. Focus on the most impactful feedback. If the code looks good, acknowledge it and highlight any particularly well-written sections.`

// BuildReviewPrompt constructs the full prompt for code review
func BuildReviewPrompt(rawDiff string, fileContents map[string]string) string {
	return BuildReviewPromptWithPruning(rawDiff, fileContents, nil)
}

// BuildReviewPromptWithPruning constructs the full prompt for code review with pruning support
func BuildReviewPromptWithPruning(rawDiff string, fileContents map[string]string, prunedFiles map[string]string) string {
	var builder strings.Builder

	builder.WriteString("## Code Review Request\n\n")
	builder.WriteString("Please review the following code changes.\n\n")

	// Add the diff
	builder.WriteString("### Git Diff (Changes)\n\n")
	builder.WriteString("```diff\n")
	builder.WriteString(rawDiff)
	builder.WriteString("\n```\n\n")

	// Add file contents for context
	if len(fileContents) > 0 {
		builder.WriteString("### Full File Context\n\n")
		builder.WriteString("Below are the complete contents of the modified files for additional context:\n\n")

		for path, content := range fileContents {
			// Check if file is pruned
			if prunedFiles != nil {
				if summary, pruned := prunedFiles[path]; pruned {
					// Use summary instead of full content
					builder.WriteString(fmt.Sprintf("#### File: `%s` (Pruned)\n\n", path))
					builder.WriteString(fmt.Sprintf("*Summary: %s*\n\n", summary))
					continue
				}
			}

			// Determine language for syntax highlighting
			lang := getLanguageFromPath(path)

			builder.WriteString(fmt.Sprintf("#### File: `%s`\n\n", path))
			builder.WriteString(fmt.Sprintf("```%s\n", lang))

			// Truncate very large files
			if len(content) > 50000 {
				builder.WriteString(content[:50000])
				builder.WriteString("\n\n... (file truncated due to size) ...\n")
			} else {
				builder.WriteString(content)
			}

			builder.WriteString("\n```\n\n")
		}
	}

	builder.WriteString("---\n\n")
	builder.WriteString("Please provide your code review based on the diff and file context above.\n")

	return builder.String()
}

// BuildFollowUpPrompt constructs a prompt for follow-up questions
func BuildFollowUpPrompt(question string) string {
	return fmt.Sprintf("Follow-up question about the code review:\n\n%s", question)
}

// getLanguageFromPath returns the language identifier for syntax highlighting
func getLanguageFromPath(path string) string {
	switch {
	case strings.HasSuffix(path, ".go"):
		return "go"
	case strings.HasSuffix(path, ".js"):
		return "javascript"
	case strings.HasSuffix(path, ".ts"):
		return "typescript"
	case strings.HasSuffix(path, ".py"):
		return "python"
	case strings.HasSuffix(path, ".rs"):
		return "rust"
	case strings.HasSuffix(path, ".java"):
		return "java"
	case strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml"):
		return "yaml"
	case strings.HasSuffix(path, ".json"):
		return "json"
	case strings.HasSuffix(path, ".md"):
		return "markdown"
	case strings.HasSuffix(path, ".sql"):
		return "sql"
	case strings.HasSuffix(path, ".sh") || strings.HasSuffix(path, ".bash"):
		return "bash"
	case strings.HasSuffix(path, ".dockerfile") || path == "Dockerfile":
		return "dockerfile"
	default:
		return ""
	}
}

// EstimateTokens provides a rough estimate of tokens in the prompt
// This is a simple heuristic: ~4 characters per token for English text
func EstimateTokens(text string) int {
	return len(text) / 4
}

// MaxTokenWarning returns a warning if the prompt is too large
func MaxTokenWarning(prompt string, maxTokens int) string {
	estimated := EstimateTokens(prompt)
	if estimated > maxTokens {
		return fmt.Sprintf("Warning: Estimated %d tokens exceeds recommended limit of %d. Consider reviewing fewer files.", estimated, maxTokens)
	}
	return ""
}
