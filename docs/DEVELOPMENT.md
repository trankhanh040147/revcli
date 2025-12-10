# Development Roadmap

## Design Principles

> These principles guide all feature development and UX decisions.

### Vim-Style Navigation
- All navigation should adapt Vim-style keybindings (`j/k`, `g/G`, `/`, etc.)
- Modal interface where appropriate (normal mode, chat mode, search mode)
- **IMPORTANT:** When adding new keyboard shortcuts, always update the help panel (`internal/ui/help.go`) to document them

### Concise CLI Flags
- All flags should have short aliases for easy typing
- Example: `--force` → `-f`, `--staged` → `-s`
- Flags must be unique per command
- Do not redefine flags in `init()`

### Keyboard-First UX
- Every action should be accessible via keyboard
- `?` shows help overlay with all keybindings
- Minimize mouse dependency

## Coding Styles

- **Constants:** Define in `[...constants.go]`. No hardcoding.
- **Input Reading:** Avoid `fmt.Scanln` (stops at whitespace). Use `bufio.NewReader(os.Stdin).ReadString('\n')`. Trim using `strings.TrimSpace` or `TrimSuffix`.
- **Stdin:** Never create multiple `bufio.NewReader(os.Stdin)` instances in the same function. Instantiate **once** and reuse. Multiple instances cause data loss in pipes/file reads.
- **OS Ops:** Use `runtime.GOOS` for external commands: `xdg-open` (Linux), `open` (macOS), `explorer` (Windows). Editor: Use `$EDITOR` env var or fallback.
- **Config:** Path: `~/.config/langtut/config.yaml`. Ensure dir exists (`os.MkdirAll`). Use YAML. Set defaults if missing.
- **Flags:** Ensure flags are unique per command. Do not redefine in `init()`. Verify existence before adding.
- **Streams:** Strict separation: logical output → `os.Stdout`, logs/errors/debug → `os.Stderr`. Enables clean piping (`cmd > file`).
- **Signal Handling:** Listen for `os.Interrupt` (`SIGINT`/`SIGTERM`). Cancel root `context` to trigger graceful shutdown/cleanup. Do not use `os.Exit` deep in library code.
- **Cobra Usage:** Use `RunE` instead of `Run`. Return errors to `main` for centralized handling/exit codes. Validate inputs in `Args` or `PreRunE`, not logic body.
- **TTY Detection:** Check if `stdout` is a terminal (`isatty`). Disable colors, spinners, and interactive prompts if piping or if `NO_COLOR` env is present.
- **Concurrency:** Use `errgroup.Group` over raw `sync.WaitGroup` to propagate errors and handle context cancellation across multiple goroutines.
- **Timeouts:** Default to a timeout for all network/IO contexts. Never allow a CLI command to hang indefinitely without user feedback.
- **Iterators:** When using Google API iterators (`google.golang.org/api/iterator`), check `if err == iterator.Done` before treating errors as exceptions. `iterator.Done` signals normal end-of-stream, not an error condition.
- **File Size:** Manage code files into small parts to reduce token costs. Split large files, keep functions focused, prefer smaller modules.
- **Line Endings:** When reading files edited by external editors, handle both Windows (`\r\n`) and Unix (`\n`) line endings. Remove trailing line endings in order: `\r\n` first, then `\n`. Prevents trailing carriage returns. 
- **YAML Marshaling:** When use `MarshalYAML()`, return root node (MappingNode/SequenceNode) directly, not wrapped in a DocumentNode. 

## Bug Fix Protocol

1. **Global Fix:** Search codebase (`rg`/`fd`) for similar patterns/implementations. Fix **all** occurrences, not just the reported one.
2. **Documentation:**
    - Update "Known Bugs" table (Status: Fixed).
    - Update "Coding Styles" if the bug reflects a common anti-pattern.
3. **Testing:** Verify edge cases: Interactive, Piped (`|`), Redirected (`<`), and Non-interactive modes.

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
- Default model changed: `gemini-1.5-flash` → `gemini-2.5-pro`

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

# v0.3.1 - Code Block Navigation & Chat Enhancements

**Status:** Planned

**Features:**

### Code Block Highlighting & Navigation
- [ ] Code block detection in review/chat responses
- [ ] Visual highlighting with purple border
- [ ] Navigate with `[` / `]` keys
- [ ] Contextual hints and block indicators
- [ ] `yb` yanks highlighted block

### Chat/Request Management
- [ ] `Ctrl+X` cancels streaming requests
- [ ] Prompt history navigation (`Ctrl+P`/`Ctrl+N`)
- [ ] Request cancellation feedback

### Refactor
- [ ] Refactor hardcoded values to constants


---

# v0.4 - Panes & Export (Lazy-git Style)

**Status:** Planned

**Features:**

### Setting Management
- [ ] Can change default setting (new subcommand)

### Code Block Navigation
- [ ] `[` / `]` - Navigate to previous/next code block
- [ ] Code block index indicator (e.g., "Block 2/5")
- [ ] Jump to specific block with number prefix (e.g., `2]` jumps to block 2)

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

### Code Block Folding
- [ ] `zc` - Fold/collapse current code block
- [ ] `zo` - Unfold/expand current code block
- [ ] `za` - Toggle fold state
- [ ] `zM` - Fold all code blocks
- [ ] `zR` - Unfold all code blocks
- [ ] Collapsed indicator showing language and line count

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
| Yank code block limited                                                   | Open   | Always yanks the **last** code block; no way to select specific block (requires Code Block Highlighting feature)                                                                 |                                         |
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

---

