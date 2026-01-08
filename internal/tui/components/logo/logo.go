// Package logo renders a RevCLI wordmark in a stylized way.
package logo

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/slice"

	"github.com/trankhanh040147/revcli/internal/tui/styles"
)

// letterform represents a letterform. It can be stretched horizontally by
// a given amount via the boolean argument.
type letterform func(bool) string

const diag = `╱`

// Opts are the options for rendering the RevCLI title art.
type Opts struct {
	FieldColor   color.Color // diagonal lines
	TitleColorA  color.Color // left gradient ramp point
	TitleColorB  color.Color // right gradient ramp point
	CharmColor   color.Color // Charm™ text color
	VersionColor color.Color // Version text color
	Width        int         // width of the rendered logo, used for truncation
}

// Render renders the RevCLI logo. Set the argument to true to render the narrow
// version, intended for use in a sidebar.
//
// The compact argument determines whether it renders compact for the sidebar
// or wider for the main pane.
func Render(version string, compact bool, o Opts) string {
	const charm = " Charm™"

	fg := func(c color.Color, s string) string {
		return lipgloss.NewStyle().Foreground(c).Render(s)
	}

	// Title.
	const spacing = 1
	letterforms := []letterform{
		letterR,
		// letterE,
		// letterV,
		// letterC,
		// letterL,
		// letterI,
	}
	stretchIndex := -1 // -1 means no stretching.
	if !compact {
		stretchIndex = cachedRandN(len(letterforms))
	}

	revcliWord := renderWord(spacing, stretchIndex, letterforms...)
	revcliWidth := lipgloss.Width(revcliWord)
	b := new(strings.Builder)
	for r := range strings.SplitSeq(revcliWord, "\n") {
		fmt.Fprintln(b, styles.ApplyForegroundGrad(r, o.TitleColorA, o.TitleColorB))
	}
	revcliWord = b.String()

	// RevCLI and version.
	metaRowGap := 1
	maxVersionWidth := revcliWidth - lipgloss.Width(charm) - metaRowGap
	version = ansi.Truncate(version, maxVersionWidth, "…") // truncate version if too long.
	gap := max(0, revcliWidth-lipgloss.Width(charm)-lipgloss.Width(version))
	metaRow := fg(o.CharmColor, charm) + strings.Repeat(" ", gap) + fg(o.VersionColor, version)

	// Join the RevCLI and big RevCLI title.
	revcliWord = strings.TrimSpace(metaRow + "\n" + revcliWord)

	// Narrow version.
	if compact {
		field := fg(o.FieldColor, strings.Repeat(diag, revcliWidth))
		return strings.Join([]string{field, field, revcliWord, field, ""}, "\n")
	}

	fieldHeight := lipgloss.Height(revcliWord)

	// Left field.
	const leftWidth = 6
	leftFieldRow := fg(o.FieldColor, strings.Repeat(diag, leftWidth))
	leftField := new(strings.Builder)
	for range fieldHeight {
		fmt.Fprintln(leftField, leftFieldRow)
	}

	// Right field.
	rightWidth := max(15, o.Width-revcliWidth-leftWidth-2) // 2 for the gap.
	const stepDownAt = 0
	rightField := new(strings.Builder)
	for i := range fieldHeight {
		width := rightWidth
		if i >= stepDownAt {
			width = rightWidth - (i - stepDownAt)
		}
		fmt.Fprint(rightField, fg(o.FieldColor, strings.Repeat(diag, width)), "\n")
	}

	// Return the wide version.
	const hGap = " "
	logo := lipgloss.JoinHorizontal(lipgloss.Top, leftField.String(), hGap, revcliWord, hGap, rightField.String())
	if o.Width > 0 {
		// Truncate the logo to the specified width.
		lines := strings.Split(logo, "\n")
		for i, line := range lines {
			lines[i] = ansi.Truncate(line, o.Width, "")
		}
		logo = strings.Join(lines, "\n")
	}
	return logo
}

// SmallRender renders a smaller version of the RevCLI logo, suitable for
// smaller windows or sidebar usage.
func SmallRender(width int) string {
	t := styles.CurrentTheme()
	title := t.S().Base.Foreground(t.Secondary).Render("RevCLI™")
	title = fmt.Sprintf("%s %s", title, styles.ApplyBoldForegroundGrad("RevCLI", t.Secondary, t.Primary))
	remainingWidth := width - lipgloss.Width(title) - 1 // 1 for the space after "RevCLI"
	if remainingWidth > 0 {
		lines := strings.Repeat("╱", remainingWidth)
		title = fmt.Sprintf("%s %s", title, t.S().Base.Foreground(t.Primary).Render(lines))
	}
	return title
}

// renderWord renders letterforms to fork a word. stretchIndex is the index of
// the letter to stretch, or -1 if no letter should be stretched.
func renderWord(spacing int, stretchIndex int, letterforms ...letterform) string {
	if spacing < 0 {
		spacing = 0
	}

	renderedLetterforms := make([]string, len(letterforms))

	// pick one letter randomly to stretch
	for i, letter := range letterforms {
		renderedLetterforms[i] = letter(i == stretchIndex)
	}

	if spacing > 0 {
		// Add spaces between the letters and render.
		renderedLetterforms = slice.Intersperse(renderedLetterforms, strings.Repeat(" ", spacing))
	}
	return strings.TrimSpace(
		lipgloss.JoinHorizontal(lipgloss.Top, renderedLetterforms...),
	)
}

// letterC renders the letter C in a stylized way. It takes an integer that
// determines how many cells to stretch the letter. If the stretch is less than
// 1, it defaults to no stretching.
func letterC(stretch bool) string {
	// Here's what we're making:
	//
	// ▄▀▀▀▀
	// █
	//	▀▀▀▀

	left := heredoc.Doc(`
		▄
		█
	`)
	right := heredoc.Doc(`
		▀

		▀
	`)
	return joinLetterform(
		left,
		stretchLetterformPart(right, letterformProps{
			stretch:    stretch,
			width:      4,
			minStretch: 7,
			maxStretch: 12,
		}),
	)
}

// letterR renders the letter R in a stylized way. It takes an integer that
// determines how many cells to stretch the letter. If the stretch is less than
// 1, it defaults to no stretching.
func letterR(stretch bool) string {
	// Here's what we're making:
	//
	// █▀▀▀▄
	// █▀▀▀▄
	// ▀   ▀

	left := heredoc.Doc(`
		█
		█
		▀
	`)
	center := heredoc.Doc(`
		▀
		▀
	`)
	right := heredoc.Doc(`
		▄
		▄
		▀
	`)
	return joinLetterform(
		left,
		stretchLetterformPart(center, letterformProps{
			stretch:    stretch,
			width:      3,
			minStretch: 7,
			maxStretch: 12,
		}),
		right,
	)
}

// letterE renders the letter E in a stylized way. It takes a boolean that
// determines whether to stretch the letter horizontally.
func letterE(stretch bool) string {
	// Here's what we're making:
	//
	// █▀▀▀▀
	// █▀▀▀▀
	// ▀▀▀▀

	left := heredoc.Doc(`
		█
		█
		▀
	`)
	right := heredoc.Doc(`
		▀
		▀
		▀
	`)
	return joinLetterform(
		left,
		stretchLetterformPart(right, letterformProps{
			stretch:    stretch,
			width:      4,
			minStretch: 7,
			maxStretch: 12,
		}),
	)
}

// letterV renders the letter V in a stylized way. It takes a boolean that
// determines whether to stretch the letter horizontally.
func letterV(stretch bool) string {
	// Here's what we're making:
	//
	// █   █
	// █   █
	//  ▀▀

	left := heredoc.Doc(`
		█
		█
		
	`)
	middle := heredoc.Doc(`
		
		▀
	`)
	right := heredoc.Doc(`
		█
		█
		
	`)

	middleSpace := stretchLetterformPart(middle, letterformProps{
		stretch:    stretch,
		width:      3,
		minStretch: 7,
		maxStretch: 12,
	})

	// Build the letter line by line to handle bottom convergence
	leftLines := strings.Split(left, "\n")
	rightLines := strings.Split(right, "\n")
	middleLines := strings.Split(middleSpace, "\n")

	// Ensure all have exactly 3 lines
	for len(leftLines) < 3 {
		leftLines = append(leftLines, "")
	}
	for len(rightLines) < 3 {
		rightLines = append(rightLines, "")
	}
	for len(middleLines) < 3 {
		middleLines = append(middleLines, "")
	}

	// FIX: Ensure top line has width (spaces) matching the bar below
	// because the heredoc starts with a newline, making line 0 empty.
	// Find the first non-empty line to get the width
	var width int
	for _, line := range middleLines {
		if w := lipgloss.Width(line); w > 0 {
			width = w
			break
		}
	}
	if width > 0 && lipgloss.Width(middleLines[0]) == 0 {
		middleLines[0] = strings.Repeat(" ", width)
	}

	// On the bottom line, make middle narrower (centered)
	bottomMiddle := stretchLetterformPart(heredoc.Doc(`
		▀
	`), letterformProps{
		stretch:    false,
		width:      2,
		minStretch: 2,
		maxStretch: 3,
	})

	// Calculate padding to center bottom
	topWidth := lipgloss.Width(middleLines[0])
	bottomWidth := lipgloss.Width(bottomMiddle)
	padding := (topWidth - bottomWidth) / 2
	if padding < 0 {
		padding = 0
	}
	middleLines[2] = strings.Repeat(" ", padding) + bottomMiddle

	// Combine line by line
	result := make([]string, 3)
	for i := 0; i < 3; i++ {
		result[i] = leftLines[i] + middleLines[i] + rightLines[i]
	}
	return strings.Join(result, "\n")
}

// letterL renders the letter L in a stylized way. It takes a boolean that
// determines whether to stretch the letter horizontally.
func letterL(stretch bool) string {
	// Here's what we're making:
	//
	// █
	// █
	// █▀▀▀

	left := heredoc.Doc(`
		█
		█
		█
	`)
	right := heredoc.Doc(`


		▀
	`)
	return joinLetterform(
		left,
		stretchLetterformPart(right, letterformProps{
			stretch:    stretch,
			width:      3,
			minStretch: 7,
			maxStretch: 12,
		}),
	)
}

// letterI renders the letter I in a stylized way. It takes a boolean that
// determines whether to stretch the letter horizontally.
func letterI(stretch bool) string {
	// Here's what we're making:
	//
	// ▄▀▀▀▄
	//   █
	//   ▀

	topLeft := heredoc.Doc(`
		▄
	`)
	topBar := heredoc.Doc(`
		▀
	`)
	topRight := heredoc.Doc(`
		▄
	`)
	middleLeft := heredoc.Doc(`

	`)
	middleCenter := heredoc.Doc(`
		█
	`)
	middleRight := heredoc.Doc(`

	`)
	bottomLeft := heredoc.Doc(`

	`)
	bottomCenter := heredoc.Doc(`
		▀
	`)
	bottomRight := heredoc.Doc(`

	`)

	topBarStretched := stretchLetterformPart(topBar, letterformProps{
		stretch:    stretch,
		width:      3,
		minStretch: 5,
		maxStretch: 8,
	})
	bottomBarStretched := stretchLetterformPart(bottomCenter, letterformProps{
		stretch:    stretch,
		width:      1,
		minStretch: 1,
		maxStretch: 3,
	})

	line1 := joinLetterform(topLeft, topBarStretched, topRight)
	line2 := joinLetterform(middleLeft, middleCenter, middleRight)
	line3 := joinLetterform(bottomLeft, bottomBarStretched, bottomRight)

	return lipgloss.JoinVertical(lipgloss.Center, line1, line2, line3)
}

func joinLetterform(letters ...string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, letters...)
}

// letterformProps defines letterform stretching properties.
// for readability.
type letterformProps struct {
	width      int
	minStretch int
	maxStretch int
	stretch    bool
}

// stretchLetterformPart is a helper function for letter stretching. If randomize
// is false the minimum number will be used.
func stretchLetterformPart(s string, p letterformProps) string {
	if p.maxStretch < p.minStretch {
		p.minStretch, p.maxStretch = p.maxStretch, p.minStretch
	}
	n := p.width
	if p.stretch {
		n = cachedRandN(p.maxStretch-p.minStretch) + p.minStretch //nolint:gosec
	}
	parts := make([]string, n)
	for i := range parts {
		parts[i] = s
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}
