package tools

import "github.com/trankhanh040147/revcli/internal/agent/tools/constants"

// Deprecated tool name constants for backward compatibility.
// Use constants.* instead.

// AgenticFetchToolName is the name of the agentic fetch tool.
// Deprecated: Use constants.AgenticFetchToolName instead.
const AgenticFetchToolName = constants.AgenticFetchToolName

// WebFetchToolName is the name of the web_fetch tool.
// Deprecated: Use constants.WebFetchToolName instead.
const WebFetchToolName = constants.WebFetchToolName

// WebSearchToolName is the name of the web_search tool for sub-agents.
// Deprecated: Use constants.WebSearchToolName instead.
const WebSearchToolName = constants.WebSearchToolName

// BashToolName is the name of the bash tool.
// Deprecated: Use constants.BashToolName instead.
const BashToolName = constants.BashToolName

// DownloadToolName is the name of the download tool.
// Deprecated: Use constants.DownloadToolName instead.
const DownloadToolName = constants.DownloadToolName

// EditToolName is the name of the edit tool.
// Deprecated: Use constants.EditToolName instead.
const EditToolName = constants.EditToolName

// WriteToolName is the name of the write tool.
// Deprecated: Use constants.WriteToolName instead.
const WriteToolName = constants.WriteToolName

// MultiEditToolName is the name of the multiedit tool.
// Deprecated: Use constants.MultiEditToolName instead.
const MultiEditToolName = constants.MultiEditToolName

// FetchToolName is the name of the fetch tool.
// Deprecated: Use constants.FetchToolName instead.
const FetchToolName = constants.FetchToolName

// ViewToolName is the name of the view tool.
// Deprecated: Use constants.ViewToolName instead.
const ViewToolName = constants.ViewToolName

// GrepToolName is the name of the grep tool.
// Deprecated: Use constants.GrepToolName instead.
const GrepToolName = constants.GrepToolName

// LSToolName is the name of the ls tool.
// Deprecated: Use constants.LSToolName instead.
const LSToolName = constants.LSToolName

// SourcegraphToolName is the name of the sourcegraph tool.
// Deprecated: Use constants.SourcegraphToolName instead.
const SourcegraphToolName = constants.SourcegraphToolName

// DiagnosticsToolName is the name of the diagnostics tool.
// Deprecated: Use constants.DiagnosticsToolName instead.
const DiagnosticsToolName = constants.DiagnosticsToolName

// TodosToolName is the name of the todos tool.
// Deprecated: Use constants.TodosToolName instead.
const TodosToolName = constants.TodosToolName

// LargeContentThreshold is the size threshold for saving content to a file.
const LargeContentThreshold = 50000 // 50KB

// AgenticFetchParams defines the parameters for the agentic fetch tool.
type AgenticFetchParams struct {
	URL    string `json:"url,omitempty" description:"The URL to fetch content from (optional - if not provided, the agent will search the web)"`
	Prompt string `json:"prompt" description:"The prompt describing what information to find or extract"`
}

// AgenticFetchPermissionsParams defines the permission parameters for the agentic fetch tool.
type AgenticFetchPermissionsParams struct {
	URL    string `json:"url,omitempty"`
	Prompt string `json:"prompt"`
}

// WebFetchParams defines the parameters for the web_fetch tool.
type WebFetchParams struct {
	URL string `json:"url" description:"The URL to fetch content from"`
}

// WebSearchParams defines the parameters for the web_search tool.
type WebSearchParams struct {
	Query      string `json:"query" description:"The search query to find information on the web"`
	MaxResults int    `json:"max_results,omitempty" description:"Maximum number of results to return (default: 10, max: 20)"`
}

// FetchParams defines the parameters for the simple fetch tool.
type FetchParams struct {
	URL     string `json:"url" description:"The URL to fetch content from"`
	Format  string `json:"format" description:"The format to return the content in (text, markdown, or html)"`
	Timeout int    `json:"timeout,omitempty" description:"Optional timeout in seconds (max 120)"`
}

// FetchPermissionsParams defines the permission parameters for the simple fetch tool.
type FetchPermissionsParams struct {
	URL     string `json:"url"`
	Format  string `json:"format"`
	Timeout int    `json:"timeout,omitempty"`
}
