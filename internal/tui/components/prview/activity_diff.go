package prview

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	sourcediff "github.com/sourcegraph/go-diff/diff"
)

type reviewDiffLine struct {
	OldLine int
	NewLine int
	Prefix  byte
	Text    string
}

func parseReviewDiffHunk(path string, diffHunk string) ([]reviewDiffLine, error) {
	diffHunk = strings.TrimSpace(diffHunk)
	if diffHunk == "" {
		return nil, nil
	}

	fileDiff, err := sourcediff.ParseFileDiff([]byte(formatReviewDiff(path, diffHunk)))
	if err != nil {
		return nil, err
	}

	var lines []reviewDiffLine
	for _, hunk := range fileDiff.Hunks {
		oldLine := int(hunk.OrigStartLine)
		newLine := int(hunk.NewStartLine)
		for _, rawLine := range bytes.Split(hunk.Body, []byte{'\n'}) {
			if len(rawLine) == 0 || rawLine[0] == '\\' {
				continue
			}

			line := reviewDiffLine{Prefix: rawLine[0], Text: string(rawLine[1:])}
			switch line.Prefix {
			case '-':
				line.OldLine = oldLine
				oldLine++
			case '+':
				line.NewLine = newLine
				newLine++
			default:
				line.OldLine = oldLine
				line.NewLine = newLine
				oldLine++
				newLine++
			}

			lines = append(lines, line)
		}
	}

	return lines, nil
}

func formatReviewDiff(path string, diffHunk string) string {
	path = filepath.ToSlash(path)
	return fmt.Sprintf("--- a/%s\n+++ b/%s\n%s\n", path, path, strings.TrimRight(diffHunk, "\n"))
}

func (m *Model) renderReviewDiffPreview(threadID string, path string, diffHunk string, width int) string {
	lines := m.reviewDiffLines(threadID, path, diffHunk)
	if len(lines) == 0 {
		return ""
	}

	oldWidth, newWidth := lineNumberWidths(lines)
	gutterWidth := oldWidth + newWidth + 4
	codeWidth := max(1, width-gutterWidth)

	gutterStyle := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText)
	contextStyle := lipgloss.NewStyle().Foreground(m.ctx.Theme.SecondaryText)
	addStyle := lipgloss.NewStyle().Foreground(m.ctx.Theme.SuccessText)
	deleteStyle := lipgloss.NewStyle().Foreground(m.ctx.Theme.ErrorText)

	rendered := make([]string, 0, len(lines))
	for _, line := range lines {
		style := contextStyle
		sign := " "
		switch line.Prefix {
		case '+':
			style = addStyle
			sign = "+"
		case '-':
			style = deleteStyle
			sign = "-"
		}

		code := line.Text
		code = ansi.Truncate(code, codeWidth, "…")
		rendered = append(rendered, lipgloss.JoinHorizontal(
			lipgloss.Top,
			gutterStyle.Render(formatDiffLineNumber(line.OldLine, oldWidth)),
			" ",
			gutterStyle.Render(formatDiffLineNumber(line.NewLine, newWidth)),
			" ",
			style.Render(sign),
			" ",
			style.Render(code),
		))
	}

	return lipgloss.NewStyle().MarginBottom(1).Render(lipgloss.JoinVertical(lipgloss.Left, rendered...))
}

func (m *Model) reviewDiffLines(threadID string, path string, diffHunk string) []reviewDiffLine {
	if strings.TrimSpace(diffHunk) == "" {
		return nil
	}
	if m.reviewDiffCache == nil {
		m.reviewDiffCache = map[string][]reviewDiffLine{}
	}

	cacheKey := threadID + "\x00" + diffHunk
	if lines, ok := m.reviewDiffCache[cacheKey]; ok {
		return lines
	}

	lines, err := parseReviewDiffHunk(path, diffHunk)
	if err != nil {
		return nil
	}
	m.reviewDiffCache[cacheKey] = lines
	return lines
}

func lineNumberWidths(lines []reviewDiffLine) (int, int) {
	oldMax := 0
	newMax := 0
	for _, line := range lines {
		oldMax = max(oldMax, line.OldLine)
		newMax = max(newMax, line.NewLine)
	}

	return len(fmt.Sprintf("%d", oldMax)), len(fmt.Sprintf("%d", newMax))
}

func formatDiffLineNumber(line int, width int) string {
	if line == 0 {
		return strings.Repeat(" ", width)
	}
	return fmt.Sprintf("%*d", width, line)
}
