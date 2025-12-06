# Development Roadmap

---

## Design Principles

> These principles guide all feature development and UX decisions.

### Vim-Style Navigation
- All navigation should adapt Vim-style keybindings (`j/k`, `g/G`, `/`, etc.)
- Modal interface where appropriate (normal mode, chat mode, search mode)

### Concise CLI Flags
- All flags should have short aliases for easy typing
- Example: `--force` → `-f`, `--staged` → `-s`

### Keyboard-First UX
- Every action should be accessible via keyboard
- `?` shows help overlay with all keybindings
- **IMPORTANT:** When adding new keyboard shortcuts, always update the help panel (`internal/ui/help.go`) to document them
- Minimize mouse dependency

### Coding Styles
- Define constants in [...constants.go], no hardcoding

---

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

# v0.3 - Short Flags & Vim UX

**Status:** In Progress

**Features Implemented:**

### Short Flag Aliases ✅
- [x] Add short aliases for all flags:
  | Long | Short | Description |
  |------|-------|-------------|
  | `--staged` | `-s` | Review staged changes |
  | `--base` | `-b` | Base branch/commit |
  | `--model` | `-m` | Model selection |
  | `--force` | `-f` | Skip secret detection |
  | `--interactive` | `-i` | Interactive mode |
  | `--no-interactive` | `-I` | Non-interactive mode |
  | `--api-key` | `-k` | API key |
  | `--preset` | `-p` | Review preset |
- [x] Add version flag (`--version`, `-v`)

### Vim-Style Keybindings ✅
- [x] Navigation: `j/k` (down/up), `g/G` (top/bottom), `Ctrl+d/u` (half-page), `Ctrl+f/b` (full page)
- [x] Search: `/` to search, `n/N` for next/prev match, `Tab` to toggle highlight/filter mode
- [x] Help: `?` to show keybindings overlay

### Yank to Clipboard ✅
- [x] `y` - Yank entire review + chat history to clipboard
- [x] `Y` - Yank only last response (without chat history)
- [x] `yb` - Yank code block (currently yanks last code block)
- [x] Visual feedback when yanked (2-second toast notification)
- [x] Help panel (`?`) documents all yank keybindings

### Review Presets ✅
- [x] `--preset <name>` / `-p` - Use predefined review style
- [x] Built-in presets: `quick`, `strict`, `security`, `performance`, `logic`, `style`, `typo`, `naming`
- [x] Custom presets in `~/.config/revcli/presets/*.yaml`

**Remaining Features:**

### Code Block Highlighting
- [ ] Show cursor while navigating
- [ ] Detect code block under cursor based on viewport position
- [ ] Highlight active code block with distinct border
- [ ] Navigate between code blocks
- [ ] Show contextual hint "Press yb to copy" when code block is focused
- [ ] `yb` yanks the highlighted block (not just first/last block)
- [ ] Stop request, still retain the last request
- [ ] Able to navigate through previous request prompt while typing current prompt
- [ ] Add flag to manage presets config

### Yank Enhancements ✅
- [x] `Y` - Yank only the last/current review (without chat history)
- [x] `y` - Yank entire conversation (review + follow-up chat)

---

# v0.4 - Panes & Export (Lazy-git Style)

**Status:** Planned

**Features:**

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

- Compare two branches directly (`revcli diff main feature-branch`)
- Review specific files only (`revcli review src/api.go`)
- Ignore patterns via `.revignore` file
- Statistics dashboard (reviews done, issues found)
- Multi-language support (i18n for prompts)
- Plugin system for custom analyzers

---

# Known Bugs

> Track and fix these issues

| Bug | Status | Notes |
|-----|--------|-------|
| Navigation issues after reviews | Fixed | No longer auto-scrolls to bottom; users read from top |
| Redundant spaces below terminal | Fixed | Dynamic viewport height calculation based on UI state |
| Yank only copies initial review | Fixed | Now yanks full content including chat history |
| Yank code block limited | Open | Always yanks the **last** code block; no way to select specific block (requires Code Block Highlighting feature) |
| Panic when using --interactive flag | Fixed | Added nil checks for renderer fallback |
| Can't type `?` in chat mode | Fixed | `?` now only triggers help in reviewing mode, passes through in chat |
| Can't press Enter for newline in chat | Fixed | Changed to `Alt+Enter` to send; Enter creates newlines |
| Textarea has white/highlighted background | Fixed | Custom textarea styling with rounded borders |

---
