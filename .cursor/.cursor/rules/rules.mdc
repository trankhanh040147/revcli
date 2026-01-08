---
alwaysApply: true
---
# revcli - Language Learning CLI Tool

## Project Context
- **Core:** Go (1.25.0) Cobra (CLI)
- **TUI Stack:**
  - `bubbletea` (Elm architecture)
  - `lipgloss` (Styling & Layout)
  - `bubbles` (Components: list, textinput, viewport)
  - `huh` (Forms & Inputs)
  - `glamour` (Markdown rendering)
- **Utilities:**
  - `samber/lo` (Lodash-style generics)
  - `bytedance/sonic` (JSON)
- **Config:** `~/.config/revcli/config.yaml`

## Coding Principles

### 1. Innovation & Tech Adoption
- **Adopt Aggressively:** Actively introduce new technologies that reduce boilerplate or improve UX.
- **Library vs. Custom:** If a library (`lo`, `huh`) solves a problem better than custom code, delete the custom code.
- **UX over UI:** Prioritize flow and speed (keyboard-first) over purely visual flair.

### 2. UX & Interface Design
- **Forms (`huh`):** Use `huh.NewForm` for multi-field inputs (wizards, settings). Avoid manual state management.
- **Components (`bubbles`):** Prefer standard `bubbles` components. Wrap them, don't rewrite them.
- **Styling (`lipgloss`):**
  - Use `lipgloss.Place()` for alignment.
  - **Never** use `strings.Repeat` for padding/borders.
  - Center via style width; never calculate string lengths manually.
- **Vim Navigation:** Use `j/k`, `g/G`, `/`. Update `internal/ui/help.go` immediately if shortcuts change.
- **Keybindings Strategy:**
  - Map `Enter`/`Tab` to `save() + nextField()`.
  - `Tab` navigation must work in both editing and navigation modes.
  - Show help with `?`. Define help text in `constants` (never inline).
- **Key Handling:** Use `key.Matches()` with centralized `KeyMap` structs. **Never** use `msg.String()`.
- **TUI Structure:** Decompose `Update`/`View` > 20 complexity into helpers. Derive boolean flags inside helpers.
- **CRUD State:** `isEditMode` depends on **persistence** (e.g., non-nil ID), not content presence. Treat pre-filled new entities as `Create`.

### 3. Core Go Implementation & Structure
- **File Size (CRITICAL):** Manage code files into small parts to reduce token costs. Split large files, keep functions focused, prefer smaller modules.
- **Structure:** Keep files/modules small and focused. Constants in `[...constants.go]` (no hardcoding).
- **Environment Variables:** Define env var names as constants in `internal/config/constants.go`. 
- **Functional Helpers (`lo`):**
  - **Strictly** use `samber/lo` for slice/map filtering, mapping, and finding.
  - **Ban** `for` loops for simple data transformation.
- **Index Safety:** Verify bounds (`0 ≤ idx < len`) before access; explicitly handle empty slices (reset index to -1 or skip).
- **Safe Mutation:** Cache original IDs/Keys **before** editing; execute deletions against the cached values, never against modified properties.
- **JSON:** Use `bytedance/sonic` for all marshalling. **Ban** `encoding/json`.
- **Concurrency:** Use `errgroup.Group` (propagates errors) over `sync.WaitGroup`.
- **Input Safety:** Single `bufio.NewReader`. **Ban** `fmt.Scanln`. Use `reader.ReadString`.
- **Flow:** Use `Cobra.RunE`. Return errors to `main`. **No** `os.Exit` in library code. Listen for `os.Interrupt`.
- **Constructors:** Return pointers (`func NewType(...) *Type`) for stateful types.
- **Memory Safety:** **Never** address map index directly. Pre-allocate slices.
- **TUI I/O:** Wrap file I/O in `tea.Cmd` to prevent blocking the UI.
- **Dependency Injection:** Inject dependencies (API keys, config) as parameters. Store in model struct.

### 4. System & Configuration
- **Config:** YAML in `~/.config/revcli/config.yaml`. Use `os.MkdirAll`. Set defaults.
- **Platform:** Use `runtime.GOOS` for generic commands (open/explorer) and `$EDITOR`.
- **Streams:** Output → `os.Stdout`. Logs/Errors → `os.Stderr`.
- **TTY:** Check `isatty(stdout)`. Disable colors if `NO_COLOR` or piping.
- **Flags:** Unique short aliases. Do not redefine in `init()`.
- **Multi-item Processing:** Reload shared state (e.g., library) at start of each iteration to ensure fresh state for each item.
- **Width Safety:** Guard `strings.Repeat` with `max(0, count)`. Guard width calculations to prevent negative values. Default to minimum width (80) if `Width() == 0` before first `WindowSizeMsg`.

### 5. Quality & Maintenance
- **[CRITICAL] DRY Enforcement:** Flag repeated logic >2 times. Extract to helpers or use `samber/lo`.
- **Function Variants:** Avoid "With*" function variants (e.g., `NewX`, `NewXWithY`). Prefer optional parameters (nil maps/slices) to consolidate logic.
- **State Transitions:** Extract repeated state transition logic (e.g., `m.state = m.previousState; m.viewport.Height = ...`) into helper methods like `returnToPreviousState()`.
- **Error Handling:** Wrap errors: `fmt.Errorf("ctx: %w", err)`. Log before fallback.
- **Method Receivers:** Mutating methods MUST use pointer receivers `(m *Type)`.
- **Standards:** Complexity ≤15. Run `gofmt -s`, `staticcheck`.

## Bug Fix Protocol
1. **Global Fix:** Search (`rg`/`fd`) codebase for similar patterns. Fix **all** occurrences.
2. **Docs:** Update "Known Bugs" in `docs/DEVELOPMENT.md`.
3. **Verify:** Test Interactive, Piped (`|`), and Non-interactive modes.
4. **Diff:** Include all new/modified files in diffs.

## Code Style
- Informal, concise, bullet-points. No fluff.
- Prioritize logic clarity and modern Go idioms.