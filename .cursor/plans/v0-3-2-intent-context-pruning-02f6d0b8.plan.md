<!-- 02f6d0b8-ff62-4ad9-849b-5ad26f30556d 4f718588-b9d4-4db3-bf2b-4c70871791fb -->
# v0.3.2 - Intent-Driven Review & Context Pruning

## Overview

Add Intent-Driven Review (pre-review form) and Context Pruning (file list + summarize & prune) features to enable focused reviews and token optimization.

## Architecture Changes

### Data Structures

**New: `internal/context/intent.go`**

- `Intent` struct: `CustomInstruction`, `FocusAreas []string`, `NegativeConstraints []string`
- Focus area mapping: `security` → `security` preset, `performance` → `performance` preset, etc.
- Helper: `BuildSystemPromptWithIntent(basePrompt, intent, presets)`

**Modified: `internal/context/builder.go`**

- `ReviewContext` adds: `Intent *Intent`, `PrunedFiles map[string]string` (file path → summary)
- `Builder` adds: `intent *Intent` field
- `Build()` uses pruned files if available

**Modified: `internal/prompt/template.go`**

- `BuildReviewPrompt()` checks `PrunedFiles` map, uses summaries instead of full content
- Add `BuildReviewPromptWithPruning(rawDiff, fileContents, prunedFiles)`

### Intent-Driven Review (Pre-Review Form)

**New: `internal/ui/intent_form.go`**

- `CollectIntent(interactive bool) (*Intent, error)` - uses `huh.NewForm`:
  - Custom instruction (optional textarea)
  - Focus areas (multiple checkboxes: Security, Performance, Logic, Style, Typo, Naming)
  - Negative constraints (optional textarea: "Ignore X, Y, Z")
- Returns `*Intent` or `nil` if skipped (non-interactive)

**Modified: `cmd/review.go`**

- Before `builder.Build()`, call `ui.CollectIntent(interactive)` if interactive
- Pass intent to `Builder` or store in `ReviewContext`
- Map focus areas to preset prompts, merge into system prompt
- If intent has focus areas, auto-inject matching preset rules

**Modified: `internal/context/builder.go`**

- `NewBuilder()` accepts optional `intent *Intent`
- `BuildSystemPromptWithIntent()` merges base prompt + focus area presets + custom instruction + negative constraints

### Context Pruning (File List + Summarize & Prune)

**New: `internal/ui/file_list.go`**

- `FileListModel` using `bubbles/list.Model`
- Items: `{Path, Size, Pruned bool}` from `ReviewContext.FileContents`
- Navigation: `j/k`, `Enter` to select, `i` to prune
- Display: file path, size, pruned indicator

**New: `internal/ui/prune.go`**

- `PruneFile(client *gemini.Client, filePath, content string) (summary string, err error)`
- Uses Gemini Flash model (`gemini-2.5-flash`) for cheap summarization
- Prompt: "Summarize this code block in one sentence: [code]"
- Returns summary string

**Modified: `internal/ui/model.go`**

- Add `StateFileList` to `State` enum
- Add `fileList list.Model` to `Model` struct
- Add `prunedFiles map[string]string` to track pruned summaries

**Modified: `internal/ui/update.go`**

- `updateKeyMsgReviewing()`: Add `i` keybinding to enter file list mode
- `updateKeyMsgFileList()`: Handle file list navigation, `i` to prune selected file
- On prune: call `PruneFile()`, update `Model.prunedFiles`, update `ReviewContext.PrunedFiles`
- `i` in file list: trigger prune action, show loading, update file list item

**Modified: `internal/ui/view.go`**

- `viewFileList()`: Render file list with `bubbles/list` component
- Show pruned indicator (✓) next to pruned files
- Footer: "j/k: navigate • i: prune • Enter: view • Esc: back"

**Modified: `internal/ui/keys.go`**

- Add `FileList` keymap with `Prune`, `Select`, `Back`

### Negative Prompting

**Modified: `internal/ui/intent_form.go`**

- Add "Negative Constraints" textarea field (optional)
- User can specify: "Don't review variable names", "Ignore style issues", etc.

**Modified: `internal/ui/model.go`**

- Add `i` keybinding in reviewing mode to add negative constraint dynamically
- Show prompt: "What should be ignored?" → append to `ReviewContext.Intent.NegativeConstraints`
- Rebuild system prompt with updated constraints

**Modified: `internal/context/builder.go`**

- `GetSystemPromptWithIntent()` appends negative constraints as: "User explicitly stated to ignore: [constraints]"

## Integration Flow

```
1. User runs `revcli review`
2. If interactive: Show intent form (huh)
3. Build ReviewContext with intent
4. Start TUI (StateLoading → StateReviewing)
5. User presses `i` → Enter file list (StateFileList)
6. User selects file, presses `i` → Prune file (call Gemini Flash)
7. Update ReviewContext.PrunedFiles
8. Future chat requests use pruned summaries
```

## File Changes

**New Files:**

- `internal/context/intent.go` - Intent data structure and helpers
- `internal/ui/intent_form.go` - Pre-review form using `huh`
- `internal/ui/file_list.go` - File list component using `bubbles/list`
- `internal/ui/prune.go` - Pruning logic with Gemini Flash

**Modified Files:**

- `cmd/review.go` - Integrate intent form before context building
- `internal/context/builder.go` - Add intent support, pruned files
- `internal/context/context.go` (if exists) or `builder.go` - Add Intent/PrunedFiles to ReviewContext
- `internal/prompt/template.go` - Use pruned files in prompt building
- `internal/ui/model.go` - Add file list state and model
- `internal/ui/update.go` - Handle file list navigation and pruning
- `internal/ui/view.go` - Render file list view
- `internal/ui/keys.go` - Add file list keybindings
- `internal/ui/help.go` - Update help text with new keybindings
- `go.mod` - Add `github.com/charmbracelet/huh` dependency

## Dependencies

- Add `github.com/charmbracelet/huh v0.x.x` (forms)
- `bubbles/list` already available via `github.com/charmbracelet/bubbles`

## Testing Considerations

- Test intent form in interactive mode (should show), non-interactive (should skip)
- Test focus area mapping to presets
- Test file list navigation and pruning
- Test that pruned files use summaries in subsequent prompts
- Test negative constraints in both form and TUI action
- Verify token savings after pruning

## Rules Updates

Update `.cursor/rules/rules.mdc`:

- Add pattern for `huh.NewForm` usage (avoid manual state management)
- Add pattern for `bubbles/list` integration (wrap, don't rewrite)
- Document intent-driven review flow

### To-dos

- [ ] Add `huh` dependency to go.mod and run go mod tidy
- [ ] Create `internal/context/intent.go` with Intent struct and focus area mapping
- [ ] Create `internal/ui/intent_form.go` with huh form for collecting intent
- [ ] Add Intent and PrunedFiles fields to ReviewContext in builder.go
- [ ] Integrate intent form into cmd/review.go before context building
- [ ] Update GetSystemPromptWithPreset to support intent (focus areas + negative constraints)
- [ ] Create `internal/ui/file_list.go` with bubbles/list for file navigation
- [ ] Create `internal/ui/prune.go` with Gemini Flash summarization
- [ ] Add StateFileList to Model, integrate file list component
- [ ] Implement `i` keybinding in reviewing mode to enter file list, `i` in file list to prune
- [ ] Update BuildReviewPrompt to use PrunedFiles map when available
- [ ] Update help.go with new keybindings (i for file list, i for prune)
- [ ] Test intent form in interactive/non-interactive modes
- [ ] Test file list navigation and pruning functionality