package constants

// Tool name constants used across the codebase.
// This package has no dependencies to avoid import cycles.
const (
	// Agent tool
	AgentToolName = "agent"

	// Execution tools
	BashToolName      = "bash"
	JobOutputToolName = "job_output"
	JobKillToolName   = "job_kill"

	// File operation tools
	DownloadToolName  = "download"
	EditToolName      = "edit"
	MultiEditToolName = "multiedit"
	WriteToolName     = "write"

	// LSP tools
	DiagnosticsToolName = "lsp_diagnostics"
	ReferencesToolName  = "lsp_references"

	// Fetch tools
	FetchToolName        = "fetch"
	AgenticFetchToolName = "agentic_fetch"
	WebFetchToolName     = "web_fetch"
	WebSearchToolName    = "web_search"

	// Search/read tools
	GlobToolName        = "glob"
	GrepToolName        = "grep"
	LSToolName          = "ls"
	SourcegraphToolName = "sourcegraph"
	ViewToolName        = "view"

	// Other tools
	TodosToolName = "todos"
)
