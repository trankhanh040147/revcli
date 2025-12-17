# Development Roadmap

## Design Principles & Coding Standards

> **Reference:** All design principles, coding standards, and implementation guidelines are defined in [`.cursor/rules/rules.mdc`](../.cursor/rules/rules.mdc).

### How To Apply These Rules

Automatically loads rules from the `.cursor/rules/` directory. The `rules.mdc` file includes `alwaysApply: true` in its frontmatter, which ensures:

- **Automatic Application:** Rules are always active during coding sessions
- **Context Awareness:** Understands project-specific patterns (Vim navigation, TUI-first UX, Go conventions)
- **Consistency:** All code suggestions follow the defined principles without manual reminders

## Bug Fix Protocol

1. **Global Fix:** Search codebase (`rg`/`fd`) for similar patterns/implementations. Fix **all** occurrences, not just the reported one.
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
- [x] Decompose monolithic `Update` function into state-specific handlers (`updateKeyMsgReviewing`, `updateKeyMsgChatting`, etc.)
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
- [x] Refactor `CalculateViewportHeight` to derive search/chat state from `State` enum instead of redundant boolean parameters
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
- [x] **Smart Context:** If the user asks for "Security," automatically inject the `security` preset rules into the system prompt.
- [x] **Intent Integration:** Intent collected via `ui.CollectIntent()` in `cmd/review.go`, passed to `Builder.WithIntent()`, merged into system prompt via `BuildSystemPromptWithIntent()`.

#### 🧠 Context Pruning (Dynamic Ignore) ✅

- [x] **"Summarize & Prune" Action:** In the TUI, pressing `i` in reviewing mode:
    1. Enters file list view (`StateFileList`) using `bubbles/list`.
    2. User selects file and presses `i` to prune.
    3. Uses Gemini Flash model (`gemini-2.5-flash`) to summarize the code file.
    4. Replaces the actual code in the context window with summary in subsequent prompts.
    5. **Benefit:** Saves massive tokens for the _next_ turn of chat while keeping the "map" of the code.
- [x] **File List Navigation:** Vim-style navigation (`j/k`) through files, visual indicator (✓) for pruned files.
- [x] **Pruning Integration:** `PrunedFiles` map in `ReviewContext`, used by `BuildReviewPromptWithPruning()` in prompt template.
- [x] **Negative Prompting:** Negative constraints collected in intent form, added to system prompt as "User explicitly stated to ignore: [constraints]".

### Implementation Details

**New Files:**
- `internal/ui/intent_form.go` - Pre-review form using `huh.NewForm`
- `internal/context/intent.go` - Intent struct and `BuildSystemPromptWithIntent()` helper
- `internal/ui/file_list.go` - File list component using `bubbles/list`
- `internal/ui/prune.go` - `PruneFile()` function using Gemini Flash for summarization
- `internal/ui/update_filelist.go` - File list state update handlers
- `internal/ui/update_prune.go` - Pruning action handlers

**Modified Files:**
- `cmd/review.go` - Collects intent before building context
- `cmd/review_helpers.go` - Integrates intent with builder via `WithIntent()`
- `internal/context/builder.go` - Added `Intent` field and `WithIntent()` method, `PrunedFiles` in `ReviewContext`
- `internal/prompt/template.go` - Added `BuildReviewPromptWithPruning()` to use summaries
- `internal/ui/model.go` - Added `StateFileList` state and `fileList` model
- `internal/ui/update_reviewing.go` - Added `i` keybinding to enter file list mode
- `internal/ui/view_model.go` - Added `viewFileList()` renderer

# v0.3.3 - Chat Enhancements

**Status:** Planned

**Features:**

### Integrating new libs
- [ ] **`samber/lo` Integration:** Refactor slice logic in `diff` and `review` packages.

| **Feature**        | **Library** | **Implementation Concept**                                                                                                        |
| ------------------ | ----------- | --------------------------------------------------------------------------------------------------------------------------------- |
| **Data Filtering** | `samber/lo` | `reviews = lo.Filter(reviews, func(r Review, _ int) bool { return !config.IsIgnored(r.RuleID) })`. Much cleaner than `for` loops. |
|                    |             |                                                                                                                                   |

### Chat/Request Management (In Testing)
- [ ] `Ctrl+X` cancels streaming requests
- [ ] Prompt history navigation (`Ctrl+P`/`Ctrl+N`)
- [ ] Request cancellation feedback


# v0.3.4 - Extend reading
- Able to read all project for context, then combine with git diff 

# v0.4 - Panes & Export (Lazy-git Style)

**Status:** Planned

**Features:**

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
- [ ] After implemented `build mode`, bring file/folder selection feature in `review mode`: add option to include other files than git diff. Interactive files/folders selection like `build mode`

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

**Presets**
- Remove `built-in` type, built-in treated as custom presets

**Uncategorized**
- Ask user for MR intention/Summary MR intention based on diff change to verify business logic
- Make the base prompt more generic/neutral (Not just Go reviewer)
- Compare two branches directly (`revcli diff main feature-branch`)
- Review specific files only (`revcli review src/api.go`)
- Ignore patterns via `.revignore` file
- Statistics dashboard (reviews done, issues found)
- Multi-language support (i18n for prompts)
- Plugin system for custom analyzers

---

# Known Bugs

> Track and fix these issues

| Bug                                                                       | Status | Notes                                                                                                                                                                            |                                         |
| ------------------------------------------------------------------------- | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------- |
| Navigation issues after reviews                                           | Fixed  | No longer auto-scrolls to bottom; users read from top                                                                                                                            |                                         |
| Redundant spaces below terminal                                           | Fixed  | Dynamic viewport height calculation based on UI state                                                                                                                            |                                         |
| Yank only copies initial review                                           | Fixed  | Now yanks full content including chat history                                                                                                                                    |                                         |
| Code block navigation removed                                             | Fixed    | Code block navigation (`[`, `]`, `yb`) removed in v0.3.1; deferred to v0.6 for complexity/UX reasons                                                                          |                                         |
| Panic when using --interactive flag                                       | Fixed  | Added nil checks for renderer fallback                                                                                                                                           |                                         |
| Can't type `?` in chat mode                                               | Fixed  | `?` now only triggers help in reviewing mode, passes through in chat                                                                                                             |                                         |
| Can't press Enter for newline in chat                                     | Fixed  | Changed to `Alt+Enter` to send; Enter creates newlines                                                                                                                           |                                         |
| Textarea has white/highlighted background                                 | Fixed  | Custom textarea styling with rounded borders                                                                                                                                     |                                         |
| Preset create fails with multi-word descriptions                          | Fixed  | `fmt.Scanln()` only reads first word; replaced with `bufio.Reader` for full-line input                                                                                           |                                         |
| Multiple stdin readers cause data loss                                    | Fixed  | Creating multiple `bufio.NewReader(os.Stdin)` instances causes buffered data loss when input is piped. Fixed by creating a single reader and reusing it.                         |                                         |
| IS01: Users can't edit custom presets                                     | Fixed  | Added `preset edit` command to allow editing custom presets interactively                                                                                                        |                                         |
| IS02: Missing feature to edit preset in command line or manually          | Fixed  | Added `preset edit` command and `preset open` command for manual editing                                                                                                         |                                         |
| IS03: Missing feature to open preset folder/file                          | Fixed  | Added `preset open` (opens in editor/file manager) and `preset path` (shows path) commands                                                                                       |                                         |
| IS04: Missing feature to set default preset                               | Fixed  | Added `preset default` command and config.yaml support for default preset                                                                                                        |                                         |
| IS05: Preset gets appended to system prompt, should optionally replace it | Fixed  | Added `replace` field to preset YAML and `--preset-replace` flag to review command                                                                                               |                                         |
| Flag redefined panic in preset command                                    | Fixed  | Duplicate flag definitions in `init()` function caused panic. Fixed by removing duplicate flag registrations. Always check for duplicate flag definitions when adding new flags. |                                         |
| IS06: No helper found when run `rv review -h`                             | Fixed  | Added `--preset-replace` flag to help examples in command Long description                                                                                                       |                                         |
| IS07: Flag is too long, not great for typing                              | Fixed  | Added short alias `-R` for `--preset-replace` flag using `BoolVarP`                                                                                                              |                                         |
| IS08: Missing feature: edit system prompt                                 | Fixed  | Added `preset system` command with `show/edit/reset` subcommands. System prompt can be customized via `~/.config/revcli/presets/system.yaml`                                     |                                         |
| IS09: Fail to edit when enter a different name                            | Fixed  | Disabled name editing in `preset edit` command. Name is now read-only to prevent data loss.                                                                                      |                                         |
| IS10: Not enter default value when editing                                | Fixed  | Improved default value prompts with clearer instructions and explicit current value display.                                                                                     |                                         |
| IS11: Can not move cursor up and down when edit prompt                    | Fixed  | Replaced line-by-line stdin input with external editor (`$EDITOR` or `vi` fallback) for multiline prompt editing.                                                                |                                         |
| IS12: Preset files when saved got `\n` instead of breaks                  | Fixed  | Added custom `MarshalYAML()` method to `Preset` struct to use literal block scalars (`                                                                                           | `) for multiline prompts in YAML files. |
| Manual layout calculation in help overlay                                | Fixed  | Replaced manual padding calculation with `lipgloss.Place()` for proper centering.                                                                                                |
| Swallowed renderer initialization error                                  | Fixed  | Added error logging to `os.Stderr` before fallback to maintain visibility during development.                                                                                    |
| Inconsistent method receivers in Model                                   | Fixed  | Converted all state-mutating methods from value receivers to pointer receivers for consistency and correctness.                                                                 |
| Stringly-typed message fields                                            | Fixed  | Added typed constants (`YankTypeReview`, `YankTypeLastResponse`, `ChatRoleUser`, `ChatRoleAssistant`) to replace raw string literals.                                           |
| Incomplete review request: core logic missing                            | Fixed  | Process issue: missing new files in diff. TUI refactor files (`update.go`, `view.go`) now included. Core logic properly decomposed.                                                |
| Inconsistent pointer receiver usage in constructor                      | Fixed  | Changed `NewModel` to return `*Model` instead of `Model`. Aligns with idiomatic Go for mutable types and prevents accidental copies.                                             |
| Stringly-typed enum (ChatRole, YankType)                                | Fixed  | Replaced string constants with typed enums (`type ChatRole int`, `type YankType int`). Enables compile-time type safety and prevents invalid assignments.                      |
| Monolithic Model file                                                    | Fixed  | Split `startReview` into `model_review.go`, viewport helpers into `viewport.go`. `model.go` now focused on struct definition and constructor only.                            |
| Incorrect integer-to-string conversion in RenderSecretWarning            | Fixed  | Replaced `string(rune(s.Line + '0'))` with `strconv.Itoa(s.Line)`. Also replaced custom `itoa()` function with `strconv.Itoa()`. Never use rune arithmetic for numeric formatting. |
| Direct stdout writes in RunSimple                                       | Fixed  | Refactored `RunSimple` to accept `io.Writer` parameter. Enables testing and output redirection. Library functions should accept `io.Writer`, only CLI commands bind to `os.Stdout`. |
| Unused Content field in YankMsg                                         | Fixed  | Removed `Content` field from `YankMsg`, now intent-only with `Type` field. Message buses should pass intents, not duplicate large content already in model state. |
| Magic number in RenderLoadingDots modulo                                | Fixed  | Replaced `dots[tick%4]` with `dots[tick%len(dots)]`. Never hardcode slice lengths in modulo/indexing operations. |
| File size: update.go exceeds 200-line limit                             | Fixed  | Split into state-specific files: `update_reviewing.go`, `update_chatting.go`, `update_searching.go`, `update_filelist.go`, `update_help.go`, `update_prune.go`. Main `update.go` now ~150 lines. |
| os.Exit in command logic (printSecretsWarning)                          | Fixed  | Changed to return `ErrSecretsDetected` error. Only `main.go` calls `os.Exit`. |
| Blocking I/O in tea.Cmd (os.Getenv inside pruneFileCmd)                 | Fixed  | Added `apiKey` field to Model, inject as parameter to `pruneFileCmd`. Dependency injection pattern for all tea.Cmd functions. |
| File size: builder.go exceeds 200-line limit                            | Fixed  | Extracted formatters (`Summary`, `DetailedSummary`, `formatBytes`) to `formatters.go`. |
| File size: model.go exceeds 200-line limit                              | Fixed  | Extracted `RunSimple` to `simple_run.go`, moved `ChatMessage`/`ChatRole` to `chat.go`. |
| Function size: runReview exceeds 50-line limit                          | Fixed  | Extracted helpers: `loadActivePreset`, `buildReviewContext`, `initializeAPIClient` to `review_helpers.go`. |
| Duplicated state transition logic                                       | Fixed  | Added `returnToPreviousState()` helper method, replaced 5 duplicated occurrences. |
| Duplicated help strings                                                  | Fixed  | Use `RenderCompactHelp(state)` in `viewFooter()` and `viewFileList()`. Centralized in `help.go`. |
| Ambiguous key binding (FileList and PruneFile both "i")                 | Fixed  | Renamed `PruneFile` to `FileListPrune` for clarity. Contextual handling remains. |
| Hardcoded environment variable name                                     | Fixed  | Created `internal/config/constants.go` with `EnvGeminiAPIKey` constant. All references updated. |
| Inconsistent UI rendering (fmt.Println in command)                      | Fixed  | Created `review_output.go` with writer-based helpers: `printReviewHeader`, `printContextSummary`, `printSecretsWarning`. |
| Streaming API blocks UI in model_review.go                              | Fixed  | Refactored `startReview()` to use goroutine + channel pattern. Streaming chunks sent via `StreamChunkMsg`, completion via `StreamDoneMsg`. UI remains responsive during streaming. |
| Non-nil return with error in builder.go                                | Fixed  | Created `SecretsError` type and `ErrSecretsDetected` sentinel. `Build()` now returns `(nil, SecretsError{Matches: ...})` instead of non-nil context with error. |
| Redundant nil checks for PrunedFiles                                    | Fixed  | Removed redundant `PrunedFiles == nil` checks from `update.go`, `update_filelist.go`, and `file_list.go`. Contract guarantees non-nil initialization in `builder.go:91`. |
| Hardcoded error type check in cmd/review.go                            | Fixed  | Replaced struct field check with `errors.As()` to extract `SecretsError` and access `Matches`. Uses proper error type checking instead of checking struct fields. |

---

