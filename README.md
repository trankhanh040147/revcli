# revcli

[![Go Report Card](https://goreportcard.com/badge/github.com/trankhanh040147/revcli)](https://goreportcard.com/report/github.com/trankhanh040147/revcli)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **Gemini-powered code reviewer CLI.**

**revcli** is a local command-line tool that acts as an intelligent peer reviewer. It reads your local git changes and uses Google's Gemini LLM to analyze your code for bugs, optimization opportunities, and best practices—all before you push a single commit.

## Features

- **Smart Context:** Analyzes `git diff` plus full file contents to understand exactly what you changed and where it fits.
- **Branch Comparison:** Compare against any branch or commit with `--base` flag (perfect for MR/PR reviews).
- **Context Preview:** See exactly which files and how many tokens will be sent before the review.
- **Token Usage Display:** Track actual token usage after each review.
- **Privacy-First:** Runs locally with built-in secret detection to prevent accidentally sending credentials to the LLM.
- **Interactive Chat:** Ask follow-up questions about the review in an interactive TUI.
- **Gemini Integration:** Leverages the large context window and reasoning of Gemini 2.5 Pro.

## Prerequisites

Before using the tool, ensure you have the following installed:

- **Go** (version 1.21 or higher)
- **Git** installed and initialized in your project.
- A **Google Gemini API Key** (Get one [here](https://aistudio.google.com/)).

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

# Install new version
go install github.com/trankhanh040147/revcli@latest
```

## Configuration

Set your Gemini API key as an environment variable:

```bash
export GEMINI_API_KEY="your-api-key-here"
```

Or pass it directly via the `--api-key` flag.

## Usage

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

The default model is `gemini-2.5-pro`. You can also use other models:

```bash
revcli review --model gemini-2.5-flash
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
- **Navigate:** Use Vim-style keys (`j/k` for up/down, `g/G` for top/bottom) or arrow keys
- **Search:** Press `/` to search within the review, `n/N` for next/previous match
- **Yank to clipboard:** Press `y` (or `yy`) to copy entire review, `Y` for last response only
- **Prompt history:** In chat mode, use `Ctrl+P` (previous) and `Ctrl+N` (next) to navigate prompt history
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

| Flag | Short | Description |
|------|------|-------------|
| `--base <ref>` | `-b` | Base branch/commit to compare against |
| `--staged` | `-s` | Review only staged changes |
| `--model <name>` | `-m` | Gemini model (default: gemini-2.5-pro) |
| `--force` | `-f` | Skip secret detection |
| `--no-interactive` | `-I` | Disable interactive TUI |
| `--interactive` | `-i` | Enable interactive TUI (default) |
| `--api-key <key>` | `-k` | Override GEMINI_API_KEY |
| `--preset <name>` | `-p` | Use predefined review preset (quick, strict, security, etc.) |
| `--version` | `-v` | Show version information |

## Development

For development information, roadmap, and version-specific context:

- **[Development Roadmap](docs/DEVELOPMENT.md)** - Complete roadmap with all versions, features, and known bugs
- **[v0.3 Development Context](docs/v0.3.md)** - Detailed context for current version development

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
