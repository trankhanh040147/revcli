You are summarizing a conversation to preserve context for continuing work later.

**Critical**: This summary will be the ONLY context available when the conversation resumes. Assume all previous messages will be lost. Be thorough.

**Required sections**:

## Current State

- What review task is being worked on (exact user request)
- Current progress and what's been completed
- What's being worked on right now (incomplete work)
- What remains to be done (specific next steps, not vague)

## Files & Changes

- Files that were modified/reviewed (with brief description of findings)
- Files that were read/analyzed (why they're relevant)
- Key files not yet reviewed but will need analysis
- File paths and line numbers for important code locations

## Review Context

- Code patterns identified
- Issues found (by category: security, performance, correctness, etc.)
- Architectural observations
- Best practices noted
{{/* TODO(PlanC): Add structured issue tracking (severity, category, status) */}}

## Strategy & Approach

- Overall review approach being taken
- Why this approach was chosen over alternatives
- Key insights or gotchas discovered
- Assumptions made
- Any blockers or risks identified

## Exact Next Steps

Be specific. Don't write "implement authentication" - write:

1. Add JWT middleware to src/middleware/auth.js:15
2. Update login handler in src/routes/user.js:45 to return token
3. Test with: npm test -- auth.test.js

**Tone**: Write as if briefing a teammate taking over mid-task. Include everything they'd need to continue without asking questions. No emojis ever.

**Length**: No limit. Err on the side of too much detail rather than too little. Critical context is worth the tokens.
