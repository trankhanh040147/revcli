package cmd

import (
	"fmt"
	"io"

	"github.com/trankhanh040147/revcli/internal/filter"
	appcontext "github.com/trankhanh040147/revcli/internal/context"
	"github.com/trankhanh040147/revcli/internal/preset"
	"github.com/trankhanh040147/revcli/internal/ui"
)

// ErrSecretsDetected is returned when secrets are detected in the code
var ErrSecretsDetected = fmt.Errorf("review aborted due to potential secrets")

// printReviewHeader prints the review header with preset and comparison info
func printReviewHeader(w io.Writer, preset *preset.Preset, baseBranch string, staged bool) {
	fmt.Fprintln(w, ui.RenderTitle("🔍 Code Review"))
	fmt.Fprintln(w)

	if preset != nil {
		mode := "append"
		if preset.Replace {
			mode = "replace"
		}
		fmt.Fprintf(w, "Using preset: %s (%s) [mode: %s]\n", preset.Name, preset.Description, mode)
	}

	if baseBranch != "" {
		fmt.Fprintf(w, "Comparing against: %s\n", baseBranch)
	} else if staged {
		fmt.Fprintln(w, "Reviewing staged changes...")
	} else {
		fmt.Fprintln(w, "Reviewing uncommitted changes...")
	}
}

// printContextSummary prints the detailed context summary
func printContextSummary(w io.Writer, ctx *appcontext.ReviewContext) {
	fmt.Fprintln(w, ui.RenderSuccess("Changes collected!"))
	fmt.Fprintln(w)
	fmt.Fprintln(w, ctx.DetailedSummary())
	fmt.Fprintln(w)
}

// printSecretsWarning prints a warning about detected secrets and returns an error
func printSecretsWarning(w io.Writer, secrets []filter.SecretMatch) error {
	fmt.Fprintln(w, ui.RenderError("Potential secrets detected in your code!"))
	fmt.Fprintln(w)

	for _, s := range secrets {
		fmt.Fprintf(w, "  • %s (line %d): %s\n", s.FilePath, s.Line, s.Match)
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, ui.RenderWarning("Review aborted to prevent sending secrets to external API."))
	fmt.Fprintln(w, ui.RenderHelp("Use --force to proceed anyway (not recommended)"))
	fmt.Fprintln(w)

	return ErrSecretsDetected
}

