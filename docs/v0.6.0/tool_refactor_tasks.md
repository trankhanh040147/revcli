# Strategic analysis

## Critique: Current state

1. Identity mismatch
   - Templates still reference "Crush" (code generation assistant)
   - `AgentCoder` name and description assume code generation tasks
   - System prompts focus on editing/writing code

2. Tool philosophy misalignment
   - Read-only tools are correct (glob, grep, ls, sourcegraph, view)
   - Write tools (edit, multiedit, write, bash) are always available but may not fit review-first workflows
   - No mode-based tool filtering yet (review/build/ask)

3. Missing review-specific capabilities
   - Templates don't emphasize code analysis, pattern detection, security scanning
   - No review-specific tooling (e.g., diff analysis, vulnerability scanning)
   - `task.md.tpl` is generic; needs review-specific context

## Elevation: State-of-the-art review tool

Core principle: "Review first, edit optionally"

### Vision elements
1. Contextual awareness
   - Understand codebase patterns before reviewing
   - Detect architectural violations, not just syntax
   - Provide actionable suggestions with examples

2. Multi-mode architecture
   - Review mode: analysis-only, suggestion-focused
   - Build mode: can generate/fix code
   - Ask mode: Q&A about codebase

3. Progressive disclosure
   - Start with high-level issues
   - Drill down to specific files
   - Show severity, category, and impact

## Three execution plans

### Plan v0.6-A: Minimal transformation (fastest)
Focus: rename and rebrand existing structure

Files to update:
- `internal/agent/templates/coder.md.tpl` → rename to `reviewer.md.tpl`
  - Replace "Crush" with "revCLI"
  - Shift focus from "writing code" to "reviewing code"
  - Keep same structure, change wording
- `internal/config/config.go`:
  - `AgentCoder` → `AgentReviewer`
  - Description: "An agent that reviews code and provides feedback"
  - Default tools: read-only only
- `internal/agent/templates/title.md`: update for review context
- `internal/agent/templates/task.md.tpl`: make review-specific

Pros: Fast, low risk, minimal changes  
Cons: Still feels like a code generator, doesn't leverage review-specific UX

---

### Plan v0.6-B: Mode-aware architecture (balanced)
Focus: introduce mode system now, keep tools flexible

New structure:
```
internal/agent/templates/
  - reviewer.md.tpl (review mode - analysis focus)
  - builder.md.tpl (build mode - generation focus)  
  - asker.md.tpl (ask mode - Q&A focus)
```

Files to update:
- `internal/config/config.go`:
  - Add `Mode` field to `Agent` struct
  - `SetupAgents()` creates: `AgentReview`, `AgentBuild`, `AgentAsk`
  - Each mode has appropriate default tools
- `internal/agent/templates/`:
  - Create mode-specific templates
  - `reviewer.md.tpl`: emphasize analysis, patterns, security
  - `builder.md.tpl`: (can reuse current `coder.md.tpl` logic)
  - `asker.md.tpl`: Q&A focused
- `internal/cmd/review.go`:
  - Use `AgentReview` mode explicitly
  - TODO comments for future mode selection
- Tool selection in `coordinator.go`:
  - Filter tools based on agent mode
  - Review mode: read-only + optional write (user configurable)

Pros: Sets foundation for modes, clear separation, future-proof  
Cons: More work upfront, need to maintain multiple templates

---

### Plan v0.6-C: Review-native redesign (most ambitious)
Focus: rebuild templates and agent logic around review workflows

New architecture:
```
Agent Capabilities:
  - Pattern Analysis (architectural, security, performance)
  - Code Quality Scoring (with rationale)
  - Interactive Review Session (follow-up questions)
  - Comparison Mode (diff-based, commit-based, branch-based)
```

Files to create/update:
- `internal/agent/templates/reviewer.md.tpl`: New template
  - Sections: Code Analysis, Security Review, Performance, Best Practices
  - Structured output format (issues categorized by severity)
  - Emphasis on explaining "why" not just "what"
- `internal/agent/tools/`:
  - Keep existing read-only tools
  - Mark write tools as "optional" in review mode
  - Add review-specific helpers if needed
- `internal/config/config.go`:
  - Add `ReviewMode` config section
  - Options: `strict` (no edits), `suggestive` (can suggest edits), `interactive` (user chooses)
- `internal/cmd/review.go`:
  - Use reviewer agent with review-optimized prompt
  - Add flags: `--strict-review`, `--allow-edits`, `--interactive`

Pros: Most aligned with review use case, strongest UX  
Cons: Most work, need to rethink workflows, may need new tool abstractions

---

# Tool Refactor Tasks - Plan v0.6-B: Mode-Aware Architecture

## Overview

Implementing Plan v0.6-B to transform revCLI from a code generation tool (crush) to a code review tool, while establishing a foundation for future multi-mode support (review, build, ask).

**Strategy**: Introduce mode-aware architecture now, keeping code flexible for future Plan v0.6-C enhancements (review-native redesign).

---

## Priority 1: Critical Changes (Must Do)

### Task 1.1: Rename AgentCoder to AgentReviewer

**File**: `internal/config/config.go`

**Changes**:
1. Update agent constant
2. Update SetupAgents() to use AgentReviewer

```go
const (
	AgentReviewer string = "reviewer"  // Changed from AgentCoder
	AgentTask     string = "task"
	// TODO(PlanC): Add AgentBuilder, AgentAsker constants for future modes
)

func (c *Config) SetupAgents() {
	allowedTools := resolveAllowedTools(allToolNames(), c.Options.DisabledTools)

	agents := map[string]Agent{
		AgentReviewer: {  // Changed from AgentCoder
			ID:           AgentReviewer,
			Name:         "Reviewer",
			Description:  "An agent that reviews code and provides feedback.",
			Model:        SelectedModelTypeLarge,
			ContextPaths: c.Options.ContextPaths,
			AllowedTools: resolveReadOnlyTools(allowedTools),  // Review mode defaults to read-only
			// TODO(PlanC): Add Mode field for mode-based tool filtering
		},

		AgentTask: {
			ID:           AgentReviewer,  // TODO: Fix this bug - should be AgentTask
			Name:         "Task",
			Description:  "An agent that helps with searching for context and finding implementation details.",
			Model:        SelectedModelTypeLarge,
			ContextPaths: c.Options.ContextPaths,
			AllowedTools: resolveReadOnlyTools(allowedTools),
			AllowedMCP: map[string][]string{},
		},
	}
	c.Agents = agents
}
```

---

### Task 1.2: Create reviewer.md.tpl from coder.md.tpl

**File**: `internal/agent/templates/reviewer.md.tpl` (new file)

**Strategy**: Copy `coder.md.tpl` and transform it for review-focused workflows.

**Key changes needed**:

```markdown
You are revCLI, a powerful AI Code Reviewer that runs in the CLI.

<critical_rules>
These rules override everything else. Follow them strictly:

1. **READ BEFORE REVIEWING**: Never review a file you haven't already read in this conversation. Pay close attention to code patterns, architecture, and context.
2. **BE AUTONOMOUS**: Don't ask questions - search, read, think, decide, act. Break complex review tasks into steps and complete them all.
3. **FOCUS ON ANALYSIS**: Your primary role is to analyze code, identify issues, and provide feedback. Code editing is optional and user-controlled.
...
</critical_rules>

<communication_style>
Keep responses focused on code review:
- Prioritize identifying issues (bugs, security, performance, maintainability)
- Provide actionable feedback with explanations
- Suggest improvements when relevant
- Be concise but thorough in analysis
...
</communication_style>

<review_workflow>
When reviewing code:
1. **Context First**: Understand the codebase structure and patterns
2. **Analysis**: Identify issues by category (security, performance, correctness, style)
3. **Prioritization**: Focus on critical issues first
4. **Suggestions**: Provide specific, actionable recommendations
5. **Positive Feedback**: Acknowledge good patterns and practices

Review categories to consider:
- **Security**: Vulnerabilities, secret exposure, injection risks
- **Performance**: Inefficient algorithms, N+1 queries, memory leaks
- **Correctness**: Logic errors, edge cases, error handling
- **Maintainability**: Code complexity, naming, documentation
- **Architecture**: Design patterns, coupling, cohesion
- **Best Practices**: Language idioms, project conventions
</review_workflow>

<code_analysis>
Before providing feedback:
1. Understand the code's purpose and context
2. Check for similar patterns in the codebase
3. Identify deviations from project conventions
4. Consider edge cases and error scenarios
5. Evaluate security implications
6. Assess performance characteristics

Focus on:
- Explaining WHY something is an issue (not just that it exists)
- Providing specific examples when possible
- Suggesting concrete improvements
- Acknowledging good practices found
</code_analysis>

// ... rest of template sections adapted for review workflow ...
// TODO(PlanC): Add sections for structured review output (severity levels, categories)
// TODO(PlanC): Add comparison mode instructions (diff-based, commit-based reviews)
```

**Note**: Keep the template structure similar to `coder.md.tpl` but adapt content for review workflows. Remove/edit sections about writing code, focus on analysis.

---

### Task 1.3: Update prompts.go to use reviewer template

**File**: `internal/agent/prompts.go`

**Changes**:
1. Add reviewer template embed
2. Create reviewerPrompt function
3. Keep coderPrompt for future build mode (Plan v0.6-C)

```go
package agent

import (
	"context"
	_ "embed"

	"github.com/trankhanh040147/revcli/internal/agent/prompt"
	"github.com/trankhanh040147/revcli/internal/config"
)

//go:embed templates/reviewer.md.tpl
var reviewerPromptTmpl []byte

//go:embed templates/coder.md.tpl  // Keep for future build mode
var coderPromptTmpl []byte

//go:embed templates/task.md.tpl
var taskPromptTmpl []byte

//go:embed templates/initialize.md.tpl
var initializePromptTmpl []byte

func reviewerPrompt(opts ...prompt.Option) (*prompt.Prompt, error) {
	systemPrompt, err := prompt.NewPrompt("reviewer", string(reviewerPromptTmpl), opts...)
	if err != nil {
		return nil, err
	}
	return systemPrompt, nil
}

// TODO(PlanC): Rename to builderPrompt for build mode
func coderPrompt(opts ...prompt.Option) (*prompt.Prompt, error) {
	systemPrompt, err := prompt.NewPrompt("coder", string(coderPromptTmpl), opts...)
	if err != nil {
		return nil, err
	}
	return systemPrompt, nil
}

func taskPrompt(opts ...prompt.Option) (*prompt.Prompt, error) {
	systemPrompt, err := prompt.NewPrompt("task", string(taskPromptTmpl), opts...)
	if err != nil {
		return nil, err
	}
	return systemPrompt, nil
}

func InitializePrompt(cfg config.Config) (string, error) {
	systemPrompt, err := prompt.NewPrompt("initialize", string(initializePromptTmpl))
	if err != nil {
		return "", err
	}
	return systemPrompt.Build(context.Background(), "", "", cfg)
}
```

---

### Task 1.4: Update coordinator.go to use reviewer agent

**File**: `internal/agent/coordinator.go`

**Changes**:
1. Update to use AgentReviewer constant
2. Update to use reviewerPrompt function
3. Add TODO comments for future mode support

```go
func NewCoordinator(
	ctx context.Context,
	cfg *config.Config,
	sessions session.Service,
	messages message.Service,
	permissions permission.Service,
	history history.Service,
	lspClients *csync.Map[string, *lsp.Client],
) (Coordinator, error) {
	c := &coordinator{
		cfg:         cfg,
		sessions:    sessions,
		messages:    messages,
		permissions: permissions,
		history:     history,
		lspClients:  lspClients,
		agents:      make(map[string]SessionAgent),
	}

	agentCfg, ok := cfg.Agents[config.AgentReviewer]  // Changed from AgentCoder
	if !ok {
		return nil, errors.New("reviewer agent not configured")
	}

	// TODO(PlanC): Make this dynamic when we support multiple agents/modes
	// For now, use reviewer prompt for review mode
	prompt, err := reviewerPrompt(prompt.WithWorkingDir(c.cfg.WorkingDir()))  // Changed from coderPrompt
	if err != nil {
		return nil, err
	}

	agent, err := c.buildAgent(ctx, prompt, agentCfg, false)
	if err != nil {
		return nil, err
	}
	c.currentAgent = agent
	c.agents[config.AgentReviewer] = agent  // Changed from AgentCoder
	return c, nil
}
```

---

### Task 1.5: Update task.md.tpl for review context

**File**: `internal/agent/templates/task.md.tpl`

**Changes**: Make it review-specific while keeping it general for search/context tasks

```markdown
You are an agent for revCLI. Given the user's prompt, you should use the tools available to you to answer the user's question.

<rules>
1. You should be concise, direct, and to the point, since your responses will be displayed on a command line interface. Answer the user's question directly, without elaboration, explanation, or details. One word answers are best. Avoid introductions, conclusions, and explanations. You MUST avoid text before/after your response, such as "The answer is <answer>.", "Here is the content of the file..." or "Based on the information provided, the answer is..." or "Here is what I will do next...".
2. When relevant, share file names and code snippets relevant to the query
3. Any file paths you return in your final response MUST be absolute. DO NOT use relative paths.
4. Clickable References (CRITICAL): All file references MUST follow the format `path/to/file.go:line_number`.
5. Focus on code analysis, pattern detection, and providing context for code review tasks.
</rules>

<env>
Working directory: {{.WorkingDir}}
Is directory a git repo: {{if .IsGitRepo}} yes {{else}} no {{end}}
Platform: {{.Platform}}
Today's date: {{.Date}}
</env>

// TODO(PlanC): Add review-specific context instructions (diff analysis, pattern matching)
```

---

## Priority 2: Important Updates (Should Do)

### Task 2.1: Update title.md for review context

**File**: `internal/agent/templates/title.md`

**Changes**: Minor wording update to reflect review context

```markdown
you will generate a short title based on the first message a user begins a conversation with

<rules>
- ensure it is not more than 50 characters long
- the title should be a summary of the user's message or review request
- it should be one line long
- do not use quotes or colons
- the entire text you return will be used as the title
- never return anything that is more than one sentence (one line) long
</rules>
```

---

### Task 2.2: Update initialize.md.tpl to remove crush references

**File**: `internal/agent/templates/initialize.md.tpl`

**Changes**: Replace references to crush/coder with revCLI/reviewer

```markdown
Analyze this codebase and create/update **{{.Config.Options.InitializeAs}}** to help future agents work effectively in this repository.

**First**: Check if directory is empty or contains only config files. If so, stop and say "Directory appears empty or only contains config. Add source code first, then run this command to generate {{.Config.Options.InitializeAs}}."

**Goal**: Document what an agent needs to know to work in this codebase - commands, patterns, conventions, gotchas.

// ... rest remains the same, just ensure no "crush" or "coder" specific references ...
```

---

### Task 2.3: Update summary.md for review context

**File**: `internal/agent/templates/summary.md`

**Changes**: Add review-specific context sections

```markdown
You are summarizing a conversation to preserve context for continuing work later.

**Critical**: This summary will be the ONLY context available when the conversation resumes. Assume all previous messages will be lost. Be thorough.

**Required sections**:

## Current State

- What review task is being worked on (exact user request)
- Current progress and what's been reviewed
- What's being worked on right now (incomplete analysis)
- What remains to be done (specific next steps, not vague)

## Files & Changes

- Files that were reviewed (with brief description of findings)
- Files that were read/analyzed (why they're relevant)
- Key files not yet reviewed but will need analysis
- File paths and line numbers for important code locations

## Review Context

- Code patterns identified
- Issues found (by category: security, performance, correctness, etc.)
- Architectural observations
- Best practices noted

// ... rest of template remains similar ...
// TODO(PlanC): Add structured issue tracking (severity, category, status)
```

---

### Task 2.4: Update all AgentCoder references to AgentReviewer

**Files to search and update**:
- `internal/agent/coordinator.go`
- `internal/tui/components/chat/header/header.go`
- `internal/tui/components/chat/splash/splash.go`
- `internal/tui/components/chat/sidebar/sidebar.go`
- `internal/tui/page/chat/chat.go`
- `internal/app/app.go`
- `internal/tui/components/dialogs/commands/commands.go`
- `internal/tui/components/dialogs/reasoning/reasoning.go`
- `internal/agent/agent_tool.go`
- `internal/config/load_test.go`

**Search pattern**: `AgentCoder` → `AgentReviewer`

**Example changes**:
```go
// Before
agentCfg, ok := cfg.Agents[config.AgentCoder]

// After
agentCfg, ok := cfg.Agents[config.AgentReviewer]
```

---

## Priority 3: Future Mode Foundation (Plan v0.6-C Prep)

### Task 3.1: Add TODO comments for future mode system

**Files**: Throughout codebase, add TODO comments marking future Plan v0.6-C work

**Examples**:

```go
// TODO(PlanC): Add Mode field to Agent struct for mode-based tool filtering
// TODO(PlanC): Implement builderPrompt for build mode
// TODO(PlanC): Implement askerPrompt for ask mode
// TODO(PlanC): Add structured review output format (severity levels, categories)
// TODO(PlanC): Add comparison mode (diff-based, commit-based reviews)
// TODO(PlanC): Add ReviewMode config (strict/suggestive/interactive)
```

---

### Task 3.2: Document tool filtering strategy for future modes

**File**: `internal/config/config.go` (comments)

```go
// Tool filtering strategy for future modes:
// - Review mode: read-only tools by default, write tools optional (user configurable)
// - Build mode: all tools available (full code generation capability)
// - Ask mode: read-only tools + specific Q&A tools
// TODO(PlanC): Implement mode-based tool filtering in coordinator
```

---

## Testing Checklist

After implementing Priority 1 and 2 tasks:

- [ ] `revcli review` command works with new AgentReviewer
- [ ] Agent uses reviewer.md.tpl template (verify no "Crush" references in output)
- [ ] Review mode defaults to read-only tools
- [ ] All tests pass (update test files with AgentReviewer constant)
- [ ] No references to "Crush" or "AgentCoder" in codebase
- [ ] Template files properly reference revCLI

---

## Future Plan v0.6-C: Review-Native Redesign (Not in This Sprint)

**Deferred work**:
1. Create structured review output format (severity levels, categories)
2. Implement builder.md.tpl for build mode
3. Implement asker.md.tpl for ask mode
4. Add ReviewMode config options (strict/suggestive/interactive)
5. Add comparison mode capabilities (diff-based, commit-based reviews)
6. Implement mode-based tool filtering in coordinator
7. Add review-specific tools (if needed)

**Notes**: Plan v0.6-C will be a larger refactoring to make revCLI fully review-native. Plan v0.6-B establishes the foundation while keeping the codebase functional.
```

This document outlines the Plan v0.6-B implementation with code snippets showing what needs to change, organized by priority, with TODOs marking Plan v0.6-C work for later.