# Markdown
```md
# Response Format (Markdown)

**OUTPUT CONSTRAINT:** Return markdown format. No conversational text, condense, no fluff.

## Response Guidelines
- **UX First**: Be extremely concise. Bullet points only. No "fluff" or summaries.
- **Tone**: Direct, professional, constructive.

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

```

