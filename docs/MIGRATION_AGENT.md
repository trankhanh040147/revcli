# Migration Guide: From Gemini SDK to Crush Agent Package

This guide explains how to migrate from using the direct Gemini SDK (`internal/gemini/`) to using the Crush agent package (`agent/`) from [charmbracelet/crush](https://github.com/trankhanh040147/revcli).

## Overview

The migration moves from:
- **Direct SDK usage**: Manual client management, history tracking, and API calls using `google.golang.org/genai`
- **Agent-based architecture**: Session-based orchestration, tool support, automatic summarization, and multi-provider support using `charm.land/fantasy`

## Key Architectural Differences

### Old Approach (Gemini SDK)

```go
// Direct client creation
client, err := gemini.NewClient(ctx, apiKey, model, cfg)
client.StartChat(systemPrompt)
response, err := client.SendMessage(ctx, message, webSearchEnabled)
```

**Characteristics:**
- Manual history management (`[]*genai.Content`)
- Direct API calls to Gemini
- Simple request/response pattern
- Session persistence via `SaveSession()`/`LoadSession()`
- Single provider (Gemini only)
- No tool support

### New Approach (Agent Package)

```go
// Coordinator-based orchestration
coordinator, err := agent.NewCoordinator(ctx, cfg, sessions, messages, permissions, history, lspClients)
result, err := coordinator.Run(ctx, sessionID, prompt, attachments...)
```

**Characteristics:**
- Session-based architecture with persistent storage
- Multi-provider support (Gemini, OpenAI, Anthropic, etc.)
- Built-in tool system (`fantasy.AgentTool`)
- Automatic summarization and token management
- Message queuing and cancellation support
- Attachments support (text, images)

## Step-by-Step Migration

### 1. Replace Client Initialization

**Before:**
```go
// cmd/review_helpers.go
func initializeClient(ctx context.Context, apiKey, model string) (*gemini.Client, error) {
    cfg, err := preset.LoadConfig()
    if err != nil {
        log.Printf("warn: failed to load configuration: %v", err)
        cfg = nil
    }
    client, err := gemini.NewClient(ctx, apiKey, model, cfg)
    if err != nil {
        return nil, fmt.Errorf("failed to create Gemini client: %w", err)
    }
    return client, nil
}
```

**After:**
```go
// Initialize Coordinator with required services
func initializeCoordinator(ctx context.Context, cfg *config.Config) (agent.Coordinator, error) {
    // These services need to be initialized (see Service Setup section)
    sessions := session.NewService(queries)
    messages := message.NewService(queries)
    permissions := permission.NewPermissionService(workingDir, skipRequests, allowedPaths)
    history := history.NewService(queries, dbConn)
    lspClients := csync.NewMap[string, *lsp.Client]()
    
    coordinator, err := agent.NewCoordinator(
        ctx,
        cfg,
        sessions,
        messages,
        permissions,
        history,
        lspClients,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create coordinator: %w", err)
    }
    return coordinator, nil
}
```

### 2. Update Configuration

The agent package uses Crush's configuration format. You'll need to adapt your configuration structure:

**Before (preset-based):**
```yaml
# ~/.config/revcli/config.yaml
gemini:
  model_params:
    temperature: 0.3
    top_p: 0.95
    top_k: 40
  safety_settings:
    threshold: "HIGH"
```

**After (Crush config format):**
```yaml
# ~/.config/revcli/config.yaml
providers:
  gemini:
    api_key: $GEMINI_API_KEY
    type: google

models:
  review-model:
    provider: gemini
    model: gemini-2.5-pro
    temperature: 0.3
    top_p: 0.95
    top_k: 40
    max_tokens: 8000

agents:
  review:
    model: review-model
    allowed_tools:
      - review_analyze
      - review_suggest
      - view
      - grep
```

### 3. Replace API Calls

**Before:**
```go
// Start chat session
client.StartChat(systemPrompt)

// Send message
response, err := client.SendMessage(ctx, userPrompt, webSearchEnabled)

// Stream message
err := client.SendMessageStream(ctx, userPrompt, func(chunk string) {
    // Handle chunk
}, webSearchEnabled)
```

**After:**
```go
// Create or get session
session, err := sessions.Create(ctx, "review-session-id", "Code Review Session")
if err != nil {
    return err
}

// Run review with coordinator
result, err := coordinator.Run(ctx, session.ID, userPrompt, attachments...)
if err != nil {
    return err
}

// Access response
responseText := result.Response.Content.Text()

// For streaming, you'll need to use message service to poll or
// implement a custom streaming handler
```

### 4. Session Management

**Before:**
```go
// Manual session persistence
client.SaveSession("my-session")
client.LoadSession("my-session")
```

**After:**
```go
// Sessions are automatically managed by session.Service
session, err := sessions.Get(ctx, sessionID)
if err != nil {
    // Session doesn't exist, create it
    session, err = sessions.Create(ctx, sessionID, "Session Title")
}

// Messages are automatically stored via message.Service
// No manual persistence needed
```

### 5. Update UI Integration

**Before:**
```go
// internal/ui/model.go
type Model struct {
    client      *gemini.Client
    flashClient *gemini.Client
    // ...
}

func streamReviewCmd(ctx context.Context, client *gemini.Client, userPrompt string, webSearchEnabled bool) tea.Cmd {
    return tea.Cmd(func() tea.Msg {
        var fullResponse string
        err := client.SendMessageStream(ctx, userPrompt, func(chunk string) {
            fullResponse += chunk
        }, webSearchEnabled)
        return ReviewResponseMsg{Text: fullResponse, Err: err}
    })
}
```

**After:**
```go
// internal/ui/model.go
type Model struct {
    coordinator agent.Coordinator
    sessionID   string
    // ...
}

func streamReviewCmd(ctx context.Context, coordinator agent.Coordinator, sessionID, userPrompt string) tea.Cmd {
    return tea.Cmd(func() tea.Msg {
        // For streaming, you may need to implement polling or use
        // the message service directly
        result, err := coordinator.Run(ctx, sessionID, userPrompt)
        if err != nil {
            return ReviewResponseMsg{Err: err}
        }
        return ReviewResponseMsg{Text: result.Response.Content.Text()}
    })
}
```

## Service Setup

The Coordinator requires several services. Here's how to set them up:

```go
import (
    "github.com/trankhanh040147/revcli/internal/session"
    "github.com/trankhanh040147/revcli/internal/message"
    "github.com/trankhanh040147/revcli/internal/permission"
    "github.com/trankhanh040147/revcli/internal/history"
    "github.com/trankhanh040147/revcli/internal/lsp"
    "github.com/trankhanh040147/revcli/internal/csync"
)

func setupServices(ctx context.Context, dataDir string) (
    session.Service,
    message.Service,
    permission.Service,
    history.Service,
    *csync.Map[string, *lsp.Client],
    error,
) {
    // Database connection (Crush uses SQLite)
    conn, err := db.Connect(ctx, dataDir)
    if err != nil {
        return nil, nil, nil, nil, nil, err
    }
    
    queries := db.New(conn)
    
    // Services
    sessions := session.NewService(queries)
    messages := message.NewService(queries)
    permissions := permission.NewPermissionService(
        workingDir,
        skipRequests, // bool: skip permission prompts
        []string{},   // allowed paths
    )
    history := history.NewService(queries, conn)
    lspClients := csync.NewMap[string, *lsp.Client]()
    
    return sessions, messages, permissions, history, lspClients, nil
}
```

## Creating Review Tools

One of the major advantages of the agent package is the ability to create custom tools. Here's how to create review-specific tools:

### Example: Review Analysis Tool

```go
// agent/tools/review_analyze_tool.go
package tools

import (
    "context"
    "fmt"
    
    "charm.land/fantasy"
    "github.com/trankhanh040147/revcli/internal/permission"
)

const ReviewAnalyzeToolName = "review_analyze"

type ReviewAnalyzeParams struct {
    FilePath    string `json:"file_path" description:"Path to the file to analyze"`
    CodeSnippet string `json:"code_snippet" description:"Code snippet to review"`
    FocusArea   string `json:"focus_area" description:"What to focus on (security, performance, style, etc.)"`
}

func NewReviewAnalyzeTool(permissions permission.Service, workingDir string) fantasy.AgentTool {
    return fantasy.NewTool(
        ReviewAnalyzeToolName,
        `Analyze code for issues, bugs, and improvements. 
        Focus on the specified area (security, performance, style, logic, etc.)
        and provide actionable feedback.`,
        ReviewAnalyzeParams{},
        func(ctx context.Context, params ReviewAnalyzeParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
            // Request permission
            p := permissions.Request(permission.CreatePermissionRequest{
                SessionID:   tools.GetSessionFromContext(ctx),
                Path:        workingDir,
                ToolCallID:  call.ID,
                ToolName:    ReviewAnalyzeToolName,
                Action:      "analyze",
                Description: fmt.Sprintf("Analyze %s for %s issues", params.FilePath, params.FocusArea),
                Params:      params,
            })
            if !p {
                return fantasy.ToolResponse{}, permission.ErrorPermissionDenied
            }
            
            // Perform analysis (this could call an external service or use the agent)
            analysis := performCodeAnalysis(params.FilePath, params.CodeSnippet, params.FocusArea)
            
            return fantasy.NewTextResponse(analysis), nil
        },
    )
}
```

### Example: Review Suggestion Tool

```go
// agent/tools/review_suggest_tool.go
package tools

import (
    "context"
    
    "charm.land/fantasy"
    "github.com/trankhanh040147/revcli/internal/permission"
)

const ReviewSuggestToolName = "review_suggest"

type ReviewSuggestParams struct {
    FilePath     string `json:"file_path" description:"Path to the file"`
    LineNumber   int    `json:"line_number" description:"Line number to suggest improvement"`
    Suggestion   string `json:"suggestion" description:"The improvement suggestion"`
    Severity     string `json:"severity" description:"Severity level (low, medium, high, critical)"`
}

func NewReviewSuggestTool(permissions permission.Service, workingDir string) fantasy.AgentTool {
    return fantasy.NewTool(
        ReviewSuggestToolName,
        `Generate code improvement suggestions for specific lines.
        Returns structured suggestions with severity levels and explanations.`,
        ReviewSuggestParams{},
        func(ctx context.Context, params ReviewSuggestParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
            p := permissions.Request(permission.CreatePermissionRequest{
                SessionID:   tools.GetSessionFromContext(ctx),
                Path:        workingDir,
                ToolCallID:  call.ID,
                ToolName:    ReviewSuggestToolName,
                Action:      "suggest",
                Description: fmt.Sprintf("Suggest improvement for %s:%d", params.FilePath, params.LineNumber),
                Params:      params,
            })
            if !p {
                return fantasy.ToolResponse{}, permission.ErrorPermissionDenied
            }
            
            suggestion := formatSuggestion(params)
            return fantasy.NewTextResponse(suggestion), nil
        },
    )
}
```

### Registering Tools in Coordinator

Update the `buildTools` method in `agent/coordinator.go`:

```go
func (c *coordinator) buildTools(ctx context.Context, agent config.Agent) ([]fantasy.AgentTool, error) {
    var allTools []fantasy.AgentTool
    
    // Add review-specific tools
    if slices.Contains(agent.AllowedTools, tools.ReviewAnalyzeToolName) {
        allTools = append(allTools, tools.NewReviewAnalyzeTool(c.permissions, c.cfg.WorkingDir()))
    }
    
    if slices.Contains(agent.AllowedTools, tools.ReviewSuggestToolName) {
        allTools = append(allTools, tools.NewReviewSuggestTool(c.permissions, c.cfg.WorkingDir()))
    }
    
    // Add existing tools...
    allTools = append(allTools,
        tools.NewViewTool(c.lspClients, c.permissions, c.cfg.WorkingDir(), c.cfg.Options.SkillsPaths...),
        tools.NewGrepTool(c.cfg.WorkingDir()),
        // ... other tools
    )
    
    // Filter by allowed tools
    var filteredTools []fantasy.AgentTool
    for _, tool := range allTools {
        if slices.Contains(agent.AllowedTools, tool.Info().Name) {
            filteredTools = append(filteredTools, tool)
        }
    }
    
    return filteredTools, nil
}
```

## New Features Enabled

### 1. Multi-Provider Support

Switch between providers without code changes:

```go
// Configuration supports multiple providers
providers:
  gemini:
    api_key: $GEMINI_API_KEY
    type: google
  openai:
    api_key: $OPENAI_API_KEY
    type: openai
  anthropic:
    api_key: $ANTHROPIC_API_KEY
    type: anthropic

models:
  review-model:
    provider: gemini  # or openai, anthropic, etc.
    model: gemini-2.5-pro
```

### 2. Automatic Summarization

Long conversations are automatically summarized to save tokens:

```go
// Automatic summarization is enabled by default
// Disable via config:
agents:
  review:
    disable_auto_summarize: false
```

### 3. Message Queuing

Multiple prompts can be queued per session:

```go
// Queue status
if coordinator.IsSessionBusy(sessionID) {
    queuedCount := coordinator.QueuedPrompts(sessionID)
    fmt.Printf("Session busy, %d prompts queued\n", queuedCount)
}

// Cancel operations
coordinator.Cancel(sessionID)      // Cancel specific session
coordinator.CancelAll()            // Cancel all sessions
coordinator.ClearQueue(sessionID)  // Clear queue
```

### 4. Attachments Support

Send file contents and images as attachments:

```go
import "github.com/trankhanh040147/revcli/internal/message"

attachments := []message.Attachment{
    message.NewTextAttachment("file.go", fileContent),
    // Images supported if model supports it
}

result, err := coordinator.Run(ctx, sessionID, prompt, attachments...)
```

### 5. Permission System

Built-in permission system for tool usage:

```go
// Tools request permission before execution
// Users can approve/deny via UI or auto-approve in config
permissions := permission.NewPermissionService(
    workingDir,
    false, // skipRequests: require user approval
    []string{}, // allowedPaths: paths that don't need approval
)
```

### 6. Sub-Agents

Create specialized agents for specific tasks (like the `agent` tool):

```go
// Create a sub-agent for focused analysis
subAgent := agent.NewSessionAgent(agent.SessionAgentOptions{
    LargeModel: smallModel,  // Use smaller model for sub-tasks
    SmallModel: smallModel,
    SystemPrompt: reviewAnalysisPrompt,
    IsSubAgent: true,
    Sessions: sessions,
    Messages: messages,
    Tools: []fantasy.AgentTool{
        tools.NewReviewAnalyzeTool(permissions, workingDir),
        tools.NewViewTool(lspClients, permissions, workingDir),
        tools.NewGrepTool(workingDir),
    },
})
```

## Complete Migration Example

Here's a complete example of migrating the review command:

```go
// cmd/review.go (migrated)

func runReview(cmd *cobra.Command, args []string) error {
    ctx := context.Background()
    
    // Load Crush config
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    // Setup services
    dataDir := filepath.Join(os.UserConfigDir(), "revcli", "data")
    sessions, messages, permissions, history, lspClients, err := setupServices(ctx, dataDir)
    if err != nil {
        return fmt.Errorf("failed to setup services: %w", err)
    }
    
    // Initialize coordinator
    coordinator, err := agent.NewCoordinator(
        ctx,
        cfg,
        sessions,
        messages,
        permissions,
        history,
        lspClients,
    )
    if err != nil {
        return fmt.Errorf("failed to create coordinator: %w", err)
    }
    
    // Build review context (unchanged)
    builder := appcontext.NewBuilder(staged, force, baseBranch)
    reviewCtx, err := buildReviewContext(builder, intent)
    if err != nil {
        return err
    }
    
    // Create session
    sessionID := generateSessionID()
    session, err := sessions.Create(ctx, sessionID, "Code Review")
    if err != nil {
        return fmt.Errorf("failed to create session: %w", err)
    }
    
    // Build prompt with review context
    prompt := buildReviewPrompt(reviewCtx, activePreset)
    
    // Create attachments from changed files
    attachments := buildAttachments(reviewCtx)
    
    // Run review
    if interactive {
        return runInteractiveReview(ctx, coordinator, session.ID, prompt, attachments)
    }
    
    // Non-interactive
    result, err := coordinator.Run(ctx, session.ID, prompt, attachments...)
    if err != nil {
        return err
    }
    
    fmt.Println(result.Response.Content.Text())
    return nil
}

func buildAttachments(reviewCtx *appcontext.ReviewContext) []message.Attachment {
    var attachments []message.Attachment
    for _, file := range reviewCtx.Files {
        content := file.Content
        attachments = append(attachments, message.NewTextAttachment(file.Path, content))
    }
    return attachments
}


## Migration Checklist

- [ ] Replace `gemini.Client` initialization with `agent.Coordinator`
- [ ] Set up required services (session, message, permission, history, lspClients)
- [ ] Update configuration format to Crush's YAML structure
- [ ] Replace `SendMessage()`/`SendMessageStream()` with `coordinator.Run()`
- [ ] Remove manual session persistence (`SaveSession`/`LoadSession`)
- [ ] Update UI to work with Coordinator and session IDs
- [ ] Create review-specific tools (optional but recommended)
- [ ] Update error handling for new error types
- [ ] Test with multiple providers (Gemini, OpenAI, etc.)
- [ ] Update tests to use agent package mocks

## Benefits Summary

1. **Multi-provider support**: Switch between AI providers easily
2. **Tool ecosystem**: Extend functionality with custom tools
3. **Better session management**: Automatic persistence and history
4. **Scalability**: Queuing, cancellation, and resource management
5. **Extensibility**: Sub-agents, custom tools, permission system
6. **Production-ready**: Battle-tested in Crush with active maintenance

## Additional Resources

- [Crush GitHub Repository](https://github.com/trankhanh040147/revcli)
- [Fantasy Package Documentation](https://pkg.go.dev/charm.land/fantasy)
- [Crush Configuration Schema](https://github.com/trankhanh040147/revcli.json)
- Agent package source: `agent/` directory in this codebase
```

This migration guide covers:

1. **Overview** - High-level differences
2. **Step-by-step migration** - Code changes needed
3. **Configuration migration** - YAML format changes
4. **Service setup** - Required dependencies
5. **Tool creation** - Examples for review-specific tools
6. **New features** - Enabled capabilities
7. **Complete example** - Full migration walkthrough
8. **Checklist** - Migration tracking

The guide emphasizes creating custom tools for review functionality, which is a major advantage of the agent package architecture.