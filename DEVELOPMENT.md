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
- Minimize mouse dependency

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
- [x] **Project rename:** `go-rev-cli` → `rev-cli`
  - Module path: `github.com/trankhanh040147/rev-cli`
  - Binary name: `rev-cli`
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
- Binary renamed: `go-rev-cli` → `rev-cli`
- Module path changed: `github.com/trankhanh040147/go-rev-cli` → `github.com/trankhanh040147/rev-cli`
- Default model changed: `gemini-1.5-flash` → `gemini-2.5-pro`

---

# v0.3 - Short Flags & Vim UX

**Status:** Planned

**Features:**

### Short Flag Aliases
- [ ] Add short aliases for all flags:
  | Long | Short | Description |
  |------|-------|-------------|
  | `--staged` | `-s` | Review staged changes |
  | `--base` | `-b` | Base branch/commit |
  | `--model` | `-m` | Model selection |
  | `--force` | `-f` | Skip secret detection |
  | `--interactive` | `-i` | Interactive mode |
  | `--no-interactive` | `-I` | Non-interactive mode |
  | `--api-key` | `-k` | API key |

### Vim-Style Keybindings
- [ ] Navigation: `j/k` (down/up), `g/G` (top/bottom), `Ctrl+d/u` (half-page)
- [ ] Search: `/` to search, `n/N` for next/prev match
- [ ] Help: `?` to show keybindings overlay

### Yank to Clipboard
- [ ] `y` - Yank entire review to clipboard
- [ ] `yb` - Yank code block under cursor
- [ ] Visual feedback when yanked

### Review Presets
- [ ] `--preset <name>` - Use predefined review style
- [ ] Built-in presets: `strict`, `security`, `performance`, `quick`
- [ ] Custom presets in `~/.config/rev-cli/presets/`

---

# v0.4 - Panes & Export (Lazy-git Style)

**Status:** Planned

**Features:**

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
- [ ] Navigate through reviews with `[` and `]`

### Export & Save
- [ ] `e` - Export current review to file
- [ ] `E` - Export entire conversation
- [ ] Auto-save conversations to `~/.local/share/rev-cli/`
- [ ] `--format json|markdown` output formats

### Config Management
- [ ] `~/.config/rev-cli/config.yaml` support
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
- [ ] `rev-cli build docs` - Generate documentation
- [ ] `rev-cli build postman` - Generate Postman collections
- [ ] Interactive file/folder selection with Vim navigation
- [ ] Read from controller, serializers, routers

### Team Features
- [ ] Shared config via `.rev-cli.yaml` in repo
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
- [ ] `rev-cli interview` - Practice coding interviews
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

- Compare two branches directly (`rev-cli diff main feature-branch`)
- Review specific files only (`rev-cli review src/api.go`)
- Ignore patterns via `.revignore` file
- Statistics dashboard (reviews done, issues found)
- Multi-language support (i18n for prompts)
- Plugin system for custom analyzers

---

# Known Bugs

> Track and fix these issues

| Bug | Status | Notes |
|-----|--------|-------|
| Navigation issues after reviews | Open | Viewport not updating properly |
| Redundant spaces below terminal | Open | Viewport height calculation issue |

---
