# revcli

[![Latest Release](https://img.shields.io/github/v/release/trankhanh040147/revcli.svg)](https://github.com/trankhanh040147/revcli/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/trankhanh040147/revcli)](https://goreportcard.com/report/github.com/trankhanh040147/revcli)
[![Go Reference](https://pkg.go.dev/badge/github.com/trankhanh040147/revcli.svg)](https://pkg.go.dev/github.com/trankhanh040147/revcli)
[![OpenSSF](https://img.shields.io/badge/OpenSSF-deps.dev-purple)](https://deps.dev/go/github.com%2Ftrankhanh040147%2Frevcli)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **LLM-powered code reviewer CLI.**

**revcli** is a local command-line tool that acts as an intelligent peer reviewer. It reads your local git changes and uses your chosen LLM provider to analyze your code for bugs, optimization opportunities, and best practices—all before you push a single commit.

## Features

- **Smart Context:** Analyzes `git diff` plus full file contents to understand exactly what you changed and where it fits.
- **Branch Comparison:** Compare against any branch or commit with `--base` flag (perfect for MR/PR reviews).
- **Context Preview:** See exactly which files and how many tokens will be sent before the review.
- **Token Usage Display:** Track actual token usage after each review.
- **Privacy-First:** Runs locally with built-in secret detection to prevent accidentally sending credentials to the LLM.
- **Interactive Chat:** Ask follow-up questions about the review in an interactive TUI.
- **Multi-Provider:** Use cloud APIs (Gemini, Claude, GPT, etc.) or local LLMs (Ollama, LM Studio, vLLM) via a single config.

## Supported providers

revcli supports many LLM providers. The list is supplied by [Catwalk](https://github.com/charmbracelet/catwalk) (Charm’s provider registry) and can be updated with `revcli update-providers`. Supported **provider types** include:

| Type | Description |
|------|-------------|
| **Google (Gemini)** | Google AI Studio – e.g. `gemini-3.0-pro`, `gemini-3.0-flash` |
| **Google Vertex AI** | Google Cloud Vertex AI |
| **OpenAI** | OpenAI API – e.g. `gpt-5.4`, `gpt-5.4-mini` |
| **Anthropic** | Claude models – e.g. `claude-opus-4.6`, `claude-sonnet-4-6` |
| **Azure OpenAI** | Azure-hosted OpenAI-compatible endpoints |
| **AWS Bedrock** | Amazon Bedrock (e.g. Claude on Bedrock) |
| **OpenRouter** | [OpenRouter](https://openrouter.ai) – many models behind one API |
| **OpenAI-compatible (local)** | Any server that speaks the OpenAI API: **Ollama**, **LM Studio**, **vLLM**, **LocalAI**, etc. |

You choose a provider and model in `~/.config/revcli/config.yaml` or via the TUI when you run `revcli` without arguments. Code review uses your configured **large** model.

## Prerequisites

Before using the tool, ensure you have the following installed:

- **Go** (version 1.25 or higher)
- **Git** installed and initialized in your project.
- At least one **LLM provider** configured (API key or local endpoint). Examples:
  - **Google Gemini:** Get an API key [here](https://aistudio.google.com/).
  - **OpenAI:** API key from [platform.openai.com](https://platform.openai.com/).
  - **Anthropic:** API key from [console.anthropic.com](https://console.anthropic.com/).
  - **Local (Ollama, LM Studio, etc.):** No key needed; set `base_url` in config (e.g. `http://localhost:11434/v1` for Ollama).

## Installation

You can install the tool directly using `go install`:

```bash
go install github.com/trankhanh040147/revcli@latest
```

Or build from source:

```bash
git clone https://github.com/trankhanh040147/revcli.git
cd revcli
make build
```

Then run **`revcli`** to start the interactive TUI (see [Usage](#usage) below).

## Configuration

Configure your provider(s) in `~/.config/revcli/config.yaml`. The app can also prompt you to pick a provider and model when you run `revcli` (no args).

- **Cloud providers:** Set the provider’s API key in the config or via the relevant env var (e.g. `GEMINI_API_KEY`, `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`). Keys are not overridable via a review command flag; use the config or env.
- **Local LLMs (Ollama, LM Studio, vLLM, etc.):** Add a provider with `type: openai-compat` and `base_url` pointing to your server (e.g. `http://localhost:11434/v1` for Ollama). No API key required for most local setups.

To refresh the list of available providers and models from the Catwalk registry:

```bash
revcli update-providers
```

## Usage

### Getting started — run revcli

The main way to use revcli is to run:

```bash
revcli
```

This launches the **interactive TUI**: you choose a provider and model (or set one up if needed), then you can start a code review or chat with the AI from the same interface. Use this when you want the full experience (sessions, follow-up chat, file list, presets).

To review code directly from the command line (e.g. in scripts or CI), use `revcli review` instead (see below).

### Basic Review

Review all uncommitted changes in your repository:

```bash
revcli review
```

### Review Against a Branch (MR/PR Style)

Compare your current changes against a base branch - perfect for merge request reviews:

```bash
# Compare against main branch
revcli review --base main

# Compare against develop branch
revcli review --base develop

# Compare against a specific commit
revcli review --base abc1234
```

### Review Staged Changes Only

Review only the changes you've staged for commit:

```bash
revcli review --staged
```

### Use a Specific Model

The model used for review is your configured **large** model in `~/.config/revcli`. You can change it via the TUI (run `revcli` and pick a provider/model) or by editing the config. The `--model` flag lets you request a specific model name for the current provider:

```bash
revcli review --model gemini-3.0-flash
```

### Non-Interactive Mode

Get the review output without the interactive chat interface:

```bash
revcli review --no-interactive
```

### Skip Secret Detection

If you're confident there are no secrets in your code (use with caution):

```bash
revcli review --force
```

### Use Review Presets

Apply predefined review styles for focused analysis:

```bash
# Quick, high-level review
revcli review --preset quick

# Comprehensive, detailed review
revcli review --preset strict

# Security-focused review
revcli review --preset security

# Performance optimization focus
revcli review --preset performance
```

Available presets: `quick`, `strict`, `security`, `performance`, `logic`, `style`, `typo`, `naming`

You can also create custom presets in `~/.config/revcli/presets/*.yaml`. See [Development Roadmap](docs/DEVELOPMENT.md) for details.

### Manage Presets

Manage your custom presets with dedicated commands:

```bash
# List all presets (built-in and custom)
revcli preset list

# Create a new custom preset
revcli preset create my-preset

# Show preset details
revcli preset show my-preset

# Delete a custom preset
revcli preset delete my-preset
```

## Interactive Mode

When running in interactive mode (default), you can:

- **View the review:** The AI analysis is displayed in a scrollable viewport
- **Ask follow-up questions:** Press `Enter` to enter chat mode, then `Alt+Enter` to send
- **Navigate:** Use Vim-style keys (`j/k` for up/down, `g/G` for top/bottom) or arrow keys; half/full page: `Ctrl+d` / `Ctrl+u`, `Ctrl+f` / `Ctrl+b`
- **Search:** Press `/` to search within the review, `n/N` for next/previous match, `Tab` to toggle highlight/filter mode, `Esc` to exit search
- **File list:** Press `i` to enter file list (prune files from context), `j/k` to navigate, `Enter` to view selected file, `Esc` to back to review
- **Yank to clipboard:** Press `y` (or `yy`) to copy entire review, `Y` for last response only
- **Prompt history:** In chat mode, use `Ctrl+P` (previous) and `Ctrl+N` (next) to navigate prompt history
- **Web search:** In chat, press `Ctrl+W` to toggle web search for the model
- **Cancel requests:** Press `Ctrl+X` to cancel a streaming request
- **Help:** Press `?` to see all available keybindings
- **Exit:** Press `q` to quit, `Esc` to exit chat mode

See the [help overlay](docs/DEVELOPMENT.md#vim-style-keybindings) for the complete list of keyboard shortcuts.

## Context Preview

Before sending to the API, revcli shows you exactly what will be reviewed:

```
📋 Review Context
─────────────────

📁 Files to review:
   • internal/api/handler.go (2.3 KB)
   • internal/api/middleware.go (1.1 KB)
   • cmd/server.go (856 B)

   Total: 3 files, 4.3 KB

🚫 Ignored files:
   • go.sum
   • internal/api/handler_test.go

📊 Token Estimate: ~1,250 tokens
```

## Token Usage

After each review, you'll see the actual token usage:

```
✓ Review completed in 3.2s
📊 Token Usage: 1,247 prompt + 892 completion = 2,139 total
```

## What Gets Reviewed

The tool analyzes:
- All modified source files
- The git diff showing exact changes
- Full file context for better understanding

The tool automatically filters out:
- `go.sum` and `go.mod` files
- `vendor/` directory
- Generated files (`*_generated.go`, `*.pb.go`)
- Test files (`*_test.go`)
- Mock files

## Security

The tool includes basic secret detection that scans for:
- API keys and tokens
- Passwords and secrets
- Private keys
- Database URLs with credentials
- Common credential patterns

If potential secrets are detected, the review is aborted unless `--force` is used.

## Review Focus Areas

The AI reviewer acts as a Senior Engineer and focuses on:

1. **Bug Detection** - Logic errors, nil pointer dereferences, race conditions
2. **Idiomatic Patterns** - Best practices for your language
3. **Performance Optimizations** - Unnecessary allocations, inefficient loops
4. **Security Concerns** - Input validation, injection risks
5. **Code Quality** - Readability, documentation, test coverage suggestions

## Example Output

```
🔍 Code Review

📋 Review Context
─────────────────
📁 Files to review:
   • internal/api/handler.go (2.3 KB)

📊 Token Estimate: ~850 tokens

### Summary
The changes implement a new user authentication handler...

### Issues Found
🔴 **Critical**: Missing error check on line 45
🟠 **Warning**: Race condition in concurrent access
🟡 **Suggestion**: Consider using sync.Pool for better performance

### Code Suggestions
...

✓ Review completed in 2.8s
📊 Token Usage: 847 prompt + 523 completion = 1,370 total
```

## Command Reference

### `revcli review` flags

| Flag | Short | Description |
|------|------|-------------|
| `--base <ref>` | `-b` | Base branch/commit to compare against |
| `--staged` | `-s` | Review only staged changes |
| `--model <name>` | `-m` | Model to use (e.g. gemini-2.5-pro, gpt-4o); depends on configured provider |
| `--force` | `-f` | Skip secret detection |
| `--no-interactive` | `-I` | Disable interactive TUI |
| `--interactive` | `-i` | Enable interactive TUI (default) |
| `--preset <name>` | `-p` | Use predefined review preset (quick, strict, security, etc.) |
| `--preset-replace` | `-R` | Replace base prompt with preset prompt instead of appending |

API keys are configured in `~/.config/revcli/config.yaml` (or via env vars); there is no `--api-key` flag on the review command.

### Global flags (all commands)

| Flag | Short | Description |
|------|------|-------------|
| `--cwd <dir>` | `-c` | Current working directory |
| `--data-dir <dir>` | `-D` | Custom revcli data directory |
| `--debug` | `-d` | Enable debug logging |
| `--yolo` | `-y` | Auto-accept all permission prompts (dangerous) |
| `--version` | `-v` | Show version information |

## Architecture & design

revcli is built around clear separation of concerns and patterns suited to AI-powered applications:

- **Coordinator pattern** — A single orchestration layer owns the agent lifecycle: it receives a review request, resolves which model and provider to use, runs the agent, and delegates persistence (sessions, messages) to dedicated services. That keeps orchestration (retries, token refresh, provider-specific options) in one place and leaves storage and UI to their own layers.

- **Streaming-first, non-blocking UI** — The AI response is streamed to the user as it's generated. The UI subscribes to updates (text chunks, tool calls, token usage) over an event bus instead of polling or blocking on I/O, so the interface stays responsive and in sync with backend state.

- **Rich tool ecosystem** — The reviewer can run shell commands; read, edit, and search files; fetch URLs and use agentic web fetch; and query LSP (diagnostics, references) when configured. Tools from [Model Context Protocol](https://modelcontextprotocol.io) (MCP) servers are exposed the same way. Which tools an agent can use is configurable, and permission prompts (e.g. "Allow bash?") are sent through the same event system so the TUI can show allow/deny dialogs without blocking.

- **Template-driven personas** — Review style and behavior are defined in **templates** (rules, communication style, review workflow, output format), not hardcoded. Context such as working directory, git status, and platform is injected at runtime. Changing "Senior Engineer" behavior or adding a new review mode is a matter of editing templates, not application code.

- **Event-driven sync** — Sessions, messages, permissions, file history, MCP, and LSP all emit create/update/delete events. The app funnels these into one stream the TUI consumes, so list views, chat, and dialogs stay in sync with no polling.

- **Multi-provider, registry-based** — Supported LLM providers and models come from a **registry** ([Catwalk](https://github.com/charmbracelet/catwalk)), not a fixed list in code. Users pick a model in config (or via the TUI); the app resolves the right backend (cloud or local OpenAI-compatible). The registry can be refreshed with `revcli update-providers`, so new providers and models appear without shipping a new binary.

For sequence diagrams and more detail, see the [Development Roadmap](docs/DEVELOPMENT.md#coordinator-architecture).

## Development

For development information, roadmap, and version-specific context:

- **[Development Roadmap](docs/DEVELOPMENT.md)** - Complete roadmap with all versions, features, and known bugs

The development documentation includes:
- Design principles and coding standards
- Feature implementation status
- Bug tracking and fixes
- Technical implementation notes
- Related file references

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

Before contributing, please review the [Development Roadmap](docs/DEVELOPMENT.md) to understand the project's direction and design principles.

## License

MIT License - see [LICENSE](LICENSE) for details.
