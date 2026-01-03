# Migration Status: Gemini SDK to Crush Agent Package

## Completed

1. **Helper Functions** (`cmd/review_helpers.go`):
   - ✅ Removed `initializeClient`, `initializeAPIClient`, `initializeFlashClient`
   - ✅ Added `buildReviewPrompt()` - builds prompt from context and preset
   - ✅ Added `buildAttachments()` - converts review context files to message attachments

## Next Steps

### 1. Update `cmd/review.go`

**Changes needed:**
- Use `setupApp(cmd)` to get app instance
- Check if `app.AgentCoordinator != nil` (handle gracefully if nil)
- Create session via `app.Sessions.Create(ctx, title)`
- Use `buildReviewPrompt()` and `buildAttachments()` helpers
- Replace `ui.Run(client, flashClient, ...)` with `ui.Run(app, sessionID, ...)`
- Replace `ui.RunSimple(client, ...)` with `ui.RunSimple(app, sessionID, ...)`

**Note:** The coordinator may be nil if config is not set up. Need to handle this case.

### 2. Update UI Package

**Files to update:**
- `internal/ui/model.go` - Replace `client *gemini.Client` with `app *app.App` and `sessionID string`
- `internal/ui/model_review.go` - Update `streamReviewCmd` to use coordinator (may need message service subscriptions for streaming)
- `internal/ui/simple_run.go` - Replace client usage with coordinator
- `internal/ui/chat.go` - Replace `SendChatMessage` to use coordinator
- `internal/ui/prune.go` and `update_prune.go` - Keep direct model calls for now, or integrate with prune tool later

**Key Challenge:** Streaming is handled differently:
- **Old:** Direct streaming from `gemini.Client.SendMessageStream()`
- **New:** Coordinator returns result, streaming is via message service subscriptions (see `app.RunNonInteractive` for example)

### 3. Prune Tool (Optional/Deferred)

The prune tool was planned but requires rethinking:
- Tools in the agent package don't have direct access to models
- Pruning is currently a UI feature (user-initiated)
- Can defer this or implement as a separate helper that uses the small model directly

### 4. Configuration

The coordinator requires Crush-style config. The config adapter task is pending. The coordinator will be nil if `cfg.IsConfigured()` returns false.

## Implementation Notes

1. **Message Service Streaming:** For streaming in UI, subscribe to `app.Messages.Subscribe(ctx)` and filter by session ID (see `app.RunNonInteractive` example)

2. **Session Management:** Sessions are created via `app.Sessions.Create()`, which returns a session with an ID. Messages are automatically persisted.

3. **Error Handling:** Update error types - coordinator uses different error types (e.g., `agent.ErrEmptyPrompt`, `agent.ErrSessionMissing`)

4. **Dependencies:** Crush packages (message, session, etc.) need to be available. Linter errors are expected until packages are copied.

## Testing Checklist

- [ ] Test with coordinator initialized (config present)
- [ ] Test with coordinator nil (config missing) - should handle gracefully
- [ ] Test interactive mode with streaming
- [ ] Test non-interactive mode
- [ ] Test session persistence
- [ ] Test error handling
- [ ] Test with multiple file attachments
