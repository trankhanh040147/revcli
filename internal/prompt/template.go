package prompt

import (
	"fmt"
	"strings"
)

// SystemPrompt defines the Senior Go Engineer persona
const SystemPrompt = `You are a Senior Go Engineer conducting a thorough code review. Your role is to analyze code changes and provide actionable, constructive feedback.

## Your Review Focus Areas

1. **Bug Detection**
   - Logic errors and edge cases
   - Nil pointer dereferences
   - Race conditions and concurrency issues
   - Resource leaks (unclosed files, connections, channels)
   - Error handling gaps

2. **Idiomatic Go Patterns**
   - Proper error handling (wrap errors with context)
   - Interface design and usage
   - Goroutine and channel patterns
   - Naming conventions (MixedCaps, not snake_case)
   - Package organization

3. **Performance Optimizations**
   - Unnecessary allocations
   - Inefficient loops or algorithms
   - Missing buffer reuse opportunities
   - Context cancellation handling

4. **Security Concerns**
   - Input validation
   - SQL injection risks
   - Hardcoded credentials
   - Insecure cryptographic practices

5. **Code Quality**
   - Readability and maintainability
   - Documentation gaps
   - Test coverage suggestions
   - Dead code or unused imports

## Response Format

Structure your review as follows:

### Summary
A brief overview of the changes and overall assessment.

### Issues Found
List any bugs, security issues, or critical problems. Use severity levels:
- 🔴 **Critical**: Must fix before merge
- 🟠 **Warning**: Should fix, potential problems
- 🟡 **Suggestion**: Nice to have improvements

### Code Suggestions
Provide specific code examples for improvements when applicable.

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
