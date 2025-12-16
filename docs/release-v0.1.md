# v0.1.0 - MVP Release

**revcli** is a Gemini-powered code reviewer CLI that analyzes your git changes before you commit. This is the first public release, providing the core functionality for intelligent code review.

## What's New

### Core CLI Framework
- Built on Cobra CLI framework with `root` and `review` commands
- Clean command structure ready for future expansion

### Git Integration
- Automatic git diff extraction (`git diff` and `git diff --staged`)
- Full file context reading for better understanding of changes
- Seamless integration with your existing git workflow

### AI Review Engine
- **Gemini API Integration**: Streaming response support for real-time feedback
- **Senior Go Engineer Persona**: Reviews code with the perspective of an experienced Go engineer
- Focuses on bugs, optimizations, and best practices

### Interactive TUI
- **Bubbletea-powered interface**: Modern terminal UI experience
- **State machine**: Smooth transitions between Loading → Reviewing → Chatting states
- **Markdown rendering**: Beautiful markdown display with Glamour
- **Follow-up chat mode**: Ask questions about the review interactively
- **Keyboard shortcuts**: 
  - `q` - Quit
  - `Enter` - Enter chat mode
  - `Esc` - Exit chat mode

### Security & Filtering
- **Secret detection**: Automatically scans for API keys, tokens, passwords, and private keys
- **Smart file filtering**: Automatically excludes:
  - `vendor/` directory
  - Generated files (`*_generated.go`, `*.pb.go`)
  - Test files (`*_test.go`)
  - `go.sum` files
- Prevents accidentally sending sensitive data to the LLM

### Command Flags
- `--staged` - Review only staged changes
- `--model` - Specify Gemini model (default: `gemini-1.5-flash`)
- `--force` - Skip secret detection (use with caution)
- `--no-interactive` - Non-interactive mode for CI/scripts

## Installation

```bash
go install github.com/trankhanh040147/revcli@v0.1.0
```

Or build from source:

```bash
git clone https://github.com/trankhanh040147/revcli.git
cd revcli
git checkout v0.1.0
make build
```

## Configuration

Set your Gemini API key:

```bash
export GEMINI_API_KEY="your-api-key-here"
```

Get your API key from [Google AI Studio](https://aistudio.google.com/).

## Usage Examples

### Basic Review
Review all uncommitted changes:

```bash
revcli review
```

### Review Staged Changes
Review only what you've staged:

```bash
revcli review --staged
```

### Non-Interactive Mode
Perfect for CI/CD pipelines:

```bash
revcli review --no-interactive > review.txt
```

### Use Different Model
Try a different Gemini model:

```bash
revcli review --model gemini-1.5-pro
```

## What's Next

Future releases will include:
- Branch comparison (`--base` flag)
- Context preview before review
- Token usage tracking
- Review presets for different review styles
- Enhanced Vim-style navigation
- And much more!

See the [Development Roadmap](https://github.com/trankhanh040147/revcli/blob/main/docs/DEVELOPMENT.md) for details.

## Feedback

This is the first public release. We'd love to hear your feedback! Please open issues or discussions on GitHub.

---

**Full Changelog**: This is the initial release. See `docs/DEVELOPMENT.md` for the complete feature list.
