# go-rev-cli

[![Go Report Card](https://goreportcard.com/badge/github.com/trankhanh040147/go-rev-cli)](https://goreportcard.com/report/github.com/trankhanh040147/go-rev-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **Gemini-powered code reviewer CLI for Go developers.**

**go-rev-cli** is a local command-line tool that acts as an intelligent peer reviewer. It reads your local git changes and uses Google's Gemini LLM to analyze your code for bugs, optimization opportunities, and idiomatic Go practices—all before you push a single commit.

## ✨ Features

- **Smart Context:** Analyzes `git diff` to understand exactly what you changed.
- **Privacy-First:** Runs locally and allows you to review the payload before sending it to the LLM.
- **Performance:** Focused on analyzing specific changes rather than the entire codebase to save tokens and time.
- **Gemini Integration:** Leverages the large context window and reasoning of Gemini 1.5.

## 🛠 Prerequisites

Before using the tool, ensure you have the following installed:

- **Go** (version 1.25 or higher)
- **Git** installed and initialized in your project.
- A **Google Gemini API Key** (Get one [here](https://aistudio.google.com/)).

## ⬇️ Installation

You can install the tool directly using `go install`:

```bash
go install [github.com/trankhanh040147/go-rev-cli@latest](https://github.com/trankhanh040147/go-rev-cli@latest)
```
