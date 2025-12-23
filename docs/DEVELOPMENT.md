# Development Roadmap

## Design Principles & Coding Standards

> **Reference:** All design principles, coding standards, and implementation guidelines are defined in [
`.cursor/rules/rules.mdc`](../.cursor/rules/rules.mdc).

### How To Apply These Rules

Automatically loads rules from the `.cursor/rules/` directory. The `rules.mdc` file includes `alwaysApply: true` in its
frontmatter, which ensures:

- **Automatic Application:** Rules are always active during coding sessions
- **Context Awareness:** Understands project-specific patterns (Vim navigation, TUI-first UX, Go conventions)
- **Consistency:** All code suggestions follow the defined principles without manual reminders

## Bug Fix Protocol

1. **Global Fix:** Search codebase (`rg`/`fd`) for similar patterns/implementations. Fix **all** occurrences, not just
   the reported one.
2. **Documentation:**
  - Update "Known Bugs" table (Status: Fixed).
  - Update coding standards in `.cursor/rules/rules.mdc` if the bug reflects a common anti-pattern.
3. **Testing:** Verify edge cases: Interactive, Piped (`|`), Redirected (`<`), and Non-interactive modes.

> **Reference:** Bug Fix Protocol are defined in [`.cursor/rules/rules.mdc`](../.cursor/rules/rules.mdc).

# v0.1 - MVP Release ✅

**Status:** Completed

**Features Implemented:**

- [x] Cobra CLI framework with `root` and `review` commands
- [x] Git diff extraction (`git diff` and `git diff --staged`)
- [x] File-scope context: reads full content of modified files
- [x] Gemini API client with streaming response support
- [x] Interactive TUI with Bubbletea
  - [x] State machine (Loading → Reviewing → Chatting)
  - [x] Markdown rendering with Glamour
  - [x] Follow-up chat mode
  - [x] Keyboard shortcuts (q: quit, Enter: chat, Esc: back)
- [x] Senior Go Engineer persona prompt
- [x] File filtering (vendor/, generated, tests, go.sum)
- [x] Secret detection (API keys, tokens, passwords, private keys)
- [x] Command flags: `--staged`, `--model`, `--force`, `--no-interactive`
- [x] Non-interactive mode for CI/scripts

---

# v0.2 - Enhanced Diff & Context ✅

**Status:** Completed

**Features Implemented:**

- [x] **Custom base branch/commit comparison**
  - `--base <branch>` - Compare against a branch (e.g., `main`, `develop`)
  - `--base <commit>` - Compare against a specific commit hash
  - MR-style diff using `git diff base...HEAD`
- [x] **Update default model** - Changed to `gemini-2.5-pro`
- [x] **Show context preview** - Display files/tokens being sent before review
  - File list with sizes
  - Total file count and size
  - Ignored files list
  - Token estimate
- [x] **Token usage display** - Show actual tokens used after review
  - Prompt tokens
  - Completion tokens
  - Total tokens

**Breaking Changes:**

- Default model changed: `gemini-2.5-flash` → `gemini-2.5-pro`

---

# v0.3.0 - Short Flags & Preset Management ✅

**Status:** Completed

**Features Implemented:**

### Short Flag Aliases ✅

- [x] Short aliases for all flags (`-s`, `-b`, `-m`, `-f`, `-i`, `-I`, `-k`, `-p`)
- [x] Version flag (`--version`, `-v`)

### Vim-Style Keybindings ✅

- [x] Navigation: `j/k`, `g/G`, `Ctrl+d/u/f/b`
- [x] Search: `/`, `n/N`, `Tab` toggle
- [x] Help overlay: `?` key

### Yank to Clipboard ✅

- [x] `y` - Yank entire review + chat history
- [x] `Y` - Yank only last response
- [x] `yb` - Yank code block
- [x] Visual feedback (toast notification)

### Review Presets ✅

- [x] `--preset <name>` / `-p` flag
- [x] Built-in presets: `quick`, `strict`, `security`, `performance`, `logic`, `style`, `typo`, `naming`
- [x] Custom presets in `~/.config/revcli/presets/*.yaml`
- [x] Default preset support via config
- [x] Preset replace mode (`--preset-replace` / `-R`)

### Preset Management Commands ✅

- [x] `preset list` - List all presets
- [x] `preset create` - Create custom preset
- [x] `preset edit` - Edit custom preset (external editor)
- [x] `preset delete` - Delete custom preset
- [x] `preset show` - Show preset details
- [x] `preset open` - Open preset file/directory
- [x] `preset path` - Show preset path
- [x] `preset default` - Set/show default preset
- [x] `preset system` - Manage system prompt (`show/edit/reset`)

---

# v0.3.1 - TUI Refactor & Code Block Removal ✅

**Status:** Completed

**Features:**

### TUI Refactoring

- [x] Replace `msg.String()` key comparisons with `key.Matches()` using centralized `KeyMap` structs
- [x] Decompose monolithic `Update` function into state-specific handlers (`updateKeyMsgReviewing`,
  `updateKeyMsgChatting`, etc.)
- [x] Decompose monolithic `View` function into state-specific renderers (`viewLoading`, `viewMain`, `viewError`)
- [x] Centralize yank chord state reset logic

### Code Block Feature Removal

- [x] Remove code block navigation (`[`, `]`) and `yb` yank functionality (deferred to v0.6)
- [x] Update all documentation to reflect removal
- [x] Add in-code comments explaining rationale for removal
- [x] `yy` now yanks entire review + chat history (no code-block navigation)
- [x] `Y` yanks only the last assistant response

### Documentation Updates

- [x] Update Coding Styles with TUI key-handling and feature-removal guidelines
- [x] Update help text and footer to remove code block references
- [x] Refactor `CalculateViewportHeight` to derive search/chat state from `State` enum instead of redundant boolean
  parameters
- [x] Add error logging for markdown rendering fallbacks to maintain visibility during development

---

# v0.3.2 - Context & Intent ✅

**Status**: Completed

## Features

#### 🎯 Intent-Driven Review (New "Prompt First") ✅

- [x] **Pre-Review Form (`huh`):** Before scanning, ask:
  - Custom instruction (e.g., "Focus on error handling").
  - Select Focus Areas (Security, Performance, Logic, Style, Typo, Naming).
  - Negative constraints (what to ignore).
- [x] **Smart Context:** If the user asks for "Security," automatically inject the `security` preset rules into the
  system prompt.
- [x] **Intent Integration:** Intent collected via `ui.CollectIntent()` in `cmd/review.go`, passed to
  `Builder.WithIntent()`, merged into system prompt via `BuildSystemPromptWithIntent()`.

#### 🧠 Context Pruning (Dynamic Ignore) ✅

- [x] **"Summarize & Prune" Action:** In the TUI, pressing `i` in reviewing mode:
  1. Enters file list view (`StateFileList`) using `bubbles/list`.
  2. User selects file and presses `i` to prune.
  3. Uses Gemini Flash model (`gemini-2.5-flash`) to summarize the code file.
  4. Replaces the actual code in the context window with summary in subsequent prompts.
  5. **Benefit:** Saves massive tokens for the _next_ turn of chat while keeping the "map" of the code.
- [x] **File List Navigation:** Vim-style navigation (`j/k`) through files, visual indicator (✓) for pruned files.
- [x] **Pruning Integration:** `PrunedFiles` map in `ReviewContext`, used by `BuildReviewPromptWithPruning()` in prompt
  template.
- [x] **Negative Prompting:** Negative constraints collected in intent form, added to system prompt as "User explicitly stated to ignore: [constraints]".

#### Gemini New Provider
- [x] Migrating Gemini provider
- [x] Toggle Enable Web Search on each request as checkbox, can be changed in follow-ups questions as well (default = true)

### Implementation Details

**New Files:**

- `internal/ui/intent_form.go` - Pre-review form using `huh.NewForm`
- `internal/context/intent.go` - Intent struct and `BuildSystemPromptWithIntent()` helper
- `internal/ui/file_list.go` - File list component using `bubbles/list`
- `internal/ui/prune.go` - `PruneFile()` function using Gemini Flash for summarization
- `internal/ui/update_filelist.go` - File list state update handlers
- `internal/ui/prune.go` - Pruning action handlers

**Modified Files:**

- `cmd/review.go` - Collects intent before building context
- `cmd/review_helpers.go` - Integrates intent with builder via `WithIntent()`
- `internal/context/builder.go` - Added `Intent` field and `WithIntent()` method, `PrunedFiles` in `ReviewContext`
- `internal/prompt/template.go` - Added `BuildReviewPromptWithPruning()` to use summaries
- `internal/ui/model.go` - Added `StateFileList` state and `fileList` model
- `internal/ui/update_reviewing.go` - Added `i` keybinding to enter file list model
- `internal/ui/view_model.go` - Added `viewFileList()` renderer

### Issues Found

- [x] IS01: Cannot ask follow-ups questions --> can not type
- [x] IS02: Can not quit (`q`) after get streaming error
- [x] IS03: Need to detect/bypass `FinishReasonSafety`
- [x] IS04: Can not cancel streaming

Here is the updated **v0.4.0** plan with the completed SDK migration removed.

# v0.4.0 - Responsive control
**Status:** Planned (Scalable Standard)

### 1. Interaction & Feedback

* [ ] **Async Pruning with Enhanced Feedback:** Implement `tea.Cmd` with file-specific spinner and non-blocking UI for other actions. Consider subtle progress for long operations.
* [ ] **Robust Cancellation (`Ctrl+X`):** Propagate `context.WithCancel` through all long-running operations; ensure immediate UI feedback and clean state on cancellation.
* [ ] **Guided Intent Input:** Upgrade `huh` form for custom text intent with validation and dynamic suggestions/auto-completion for focus areas.

### 2. DevOps & CI/CD

* [ ] **Actionable Security Workflow:** Integrate OpenSSF Scorecard (`scorecard.yaml`) in CI; explore `revcli` consumption for in-terminal insights.
* [ ] **Secure Release Automation:** Configure GoReleaser (`.goreleaser.yaml`) for multi-platform builds, Homebrew tap, and integrate Cosign for artifact signing.
* [ ] **Fast & Comprehensive CI Pipeline:** Add `golangci-lint` (strict config) and `go test -race`; optimize for speed and provide local pre-commit targets.

# v0.4.1 - Structured Intelligence

### Bugs
- [ ] Change keymap for toggle Web Search

### Core features

#### 1. Core Logic & Data Structure (Prerequisite)

* [ ] **Define Schema:** Implement `ReviewIssue` struct and map it to **OpenAPI 3.0** schema.
* [ ] **Tool Configuration:** Configure `submit_review` tool to force **deterministic** JSON output.
* [ ] **JSON Unmarshaling:** Implement logic to bridge `map[string]interface{}` responses back to strict Go structs.
* [ ] **Safe Fallback:** Handle cases where the model refuses to call the function (fallback to text).

#### 2. TUI & Visualization (UX Focused)

* [ ] **List View:** Replace Markdown viewport with `bubbles/list` for navigable issue tracking.
* [ ] **Custom Delegate:** Implement `lipgloss` rendering for colored Severity pills and Category tags.
* [ ] **Detail State:** Create `StateDetailView` (Enter key) to render full suggestion/context using Glamour.
* [ ] **Token Transparency:** Extract `UsageMetadata` from JSON response; display "Tokens In/Out & Cost" in list footer.

### Planned features
#### Context Intelligence
* [ ] **Dependency Graph:** Implement Regex-based import scanning to find "Related Context" (files that import the changed code).
* [ ] **Smart Pruning:** Feed "Related Context" summaries into the prompt to detect breaking changes in other files.
#### Refactoring
* [ ] **`samber/lo` Integration:** Refactor slice logic in `diff` and `review` packages using declarative pipelines (Filter, Map).
* [ ] **Ignore Management:** Implement `.revignore` support (using `samber/lo` to filter).

# v0.4.2 - Panes & Export (Lazy-git Style)

**Status:** Planned

**Features:**

### The "Lazy" Experience (UX)
- [ ] **Interactive Patching:** `Apply` button that actually writes code.
- [ ] **Panes:** Reviews | Chat | Config (Tab to switch).

### Setting Management

- [ ] Can change default setting (new subcommand)

### Panes Management Mode

- [ ] Multi-pane layout inspired by lazy-git/lazy-docker
- [ ] Panes:
  - Reviews pane (list of reviews in session)
  - Conversation pane (current chat)
  - Config pane (model, API key, style)
- [ ] `Tab` to switch between panes
- [ ] `1/2/3` to jump to specific pane

### Review Actions

- [ ] `a` - Accept/apply suggestion
- [ ] `x` - Reject/ignore suggestion
- [ ] Add to ignore list (global/conversation)
- [ ] Navigate through suggestions with `[` and `]`

### Export & Save

- [ ] `e` - Export current review to file
- [ ] `E` - Export entire conversation
- [ ] Auto-save conversations to `~/.local/share/revcli/`
- [ ] `--format json|markdown` output formats

### Config Management

- [ ] `~/.config/revcli/config.yaml` support
- [ ] Settings: default model, base branch, ignore patterns
- [ ] In-app config editing via config pane

---

# v0.5 - Power User Features

**Status:** Future

**Features:**

### Token Rotation

- [ ] Support multiple API keys
- [ ] Round-robin rotation between keys
- [ ] Auto-switch on rate limit
- [ ] Key usage tracking per key

### Blacklist & Filters

- [ ] Blacklist review styles (e.g., "don't suggest X")
- [ ] Global vs conversation-level blacklist
- [ ] `--min-severity` flag to filter output

### Dry-Run & Preview

- [ ] `--dry-run` / `-n` - Preview payload without API call
- [ ] `--list-models` - Show available models
- [ ] Token cost estimation

# v0.6 - Code Block Management (Deferred)

**Status:** Deferred

**Features:**

### Code Block Highlighting & Navigation

- [ ] Code block detection in review/chat responses
- [ ] Visual highlighting with purple border
- [ ] Navigate with `[` / `]` keys
- [ ] Contextual hints and block indicators
- [ ] `yb` yanks highlighted block
- [ ] Code block index indicator (e.g., "Block 2/5")
- [ ] Jump to specific block with number prefix (e.g., `2]` jumps to block 2)

### Code Block Folding

- [ ] `zc` - Fold/collapse current code block
- [ ] `zo` - Unfold/expand current code block
- [ ] `za` - Toggle fold state
- [ ] `zM` - Fold all code blocks
- [ ] `zR` - Unfold all code blocks
- [ ] Collapsed indicator showing language and line count

---

# v1.0 - Production Ready

**Status:** Future

**Features:**

### Multiple AI Providers

- [ ] OpenAI GPT-4
- [ ] Anthropic Claude
- [ ] Local models (Ollama)
- [ ] `--provider` flag

### Build Mode

- [ ] `revcli build docs` - Generate documentation
- [ ] `revcli build postman` - Generate Postman collections
- [ ] Interactive file/folder selection with Vim navigation
- [ ] Read from controller, serializers, routers
- [ ] After implemented `build mode`, bring file/folder selection feature in `review mode`: add option to include other
  files than git diff. Interactive files/folders selection like `build mode`

### Team Features

- [ ] Shared config via `.revcli.yaml` in repo
- [ ] Team-specific prompts and rules
- [ ] Pre-commit hook integration

### Advanced UI

- [ ] VS Code extension
- [ ] Review annotations (inline comments)
- [ ] Diff viewer with syntax highlighting

---

# v2.0 - Future Vision

**Status:** Ideas

**Features:**

### Interview Mode

- [ ] `revcli interview` - Practice coding interviews
- [ ] Algorithm questions with hints
- [ ] Code review practice

### Auto-Fix

- [ ] Apply LLM suggestions automatically
- [ ] `--auto-fix` flag for non-breaking changes
- [ ] Git commit integration

### Integrations

- [ ] GitHub Action
- [ ] GitLab CI template
- [ ] PR/MR comment posting
- [ ] SonarQube/CodeClimate integration

---

# Ideas Backlog

> Raw ideas for future consideration

**Unit Test**
- Have an option in review mode to generate unit test
- Dedicated command to generate Unit Test (build/test)

**File mention**
- Prompt for choosing files when type `@`
- Can choose files to mention when first enter

**Presets**

- Remove `built-in` type, built-in treated as custom presets

**Uncategorized**

- Ask user for MR intention/Summary MR intention based on diff change to verify business logic
- Make the base prompt more generic/neutral (Not just Go reviewer)
- Compare two branches directly (`revcli diff main feature-branch`)
- Review specific files only (`revcli review src/api.go`)
- Ignore patterns via `.revignore` file
- Statistics dashboard (reviews done, issues found)
- Multi-language support (i18n for prompts)
- Plugin system for custom analyzers