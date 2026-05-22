package tui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestExtractCopySelectionTextClampsToPaneContent(t *testing.T) {
	bounds := copySelectionBounds{x: 10, y: 5, width: 8, height: 3}
	content := "alpha bravo\ncharlie delta\necho foxtrot"

	got := extractCopySelectionText(content, bounds, 12, 5, 25, 6)
	want := "pha br\ncharlie"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestExtractCopySelectionTextHandlesReverseDrag(t *testing.T) {
	bounds := copySelectionBounds{x: 0, y: 0, width: 10, height: 2}
	content := "0123456789\nabcdefghij"

	got := extractCopySelectionText(content, bounds, 6, 1, 2, 0)
	want := "23456789\nabcdefg"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestExtractCopySelectionTextStripsANSI(t *testing.T) {
	bounds := copySelectionBounds{x: 0, y: 0, width: 10, height: 1}
	content := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("copy me")

	got := extractCopySelectionText(content, bounds, 0, 0, 6, 0)
	want := "copy me"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestClampCopySelectionPointStaysInsideBounds(t *testing.T) {
	bounds := copySelectionBounds{x: 5, y: 3, width: 10, height: 4}

	x, y := clampCopySelectionPoint(100, -1, bounds)
	if x != 14 || y != 3 {
		t.Fatalf("expected clamped point (14, 3), got (%d, %d)", x, y)
	}
}

func TestStripCopySelectionPreviewBorderRight(t *testing.T) {
	got := stripCopySelectionPreviewBorder("│hello\n│world", "right", 1)
	want := "hello\nworld"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestStripCopySelectionPreviewBorderRightStripsStyledDivider(t *testing.T) {
	divider := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("│")
	got := stripCopySelectionPreviewBorder(divider+"hello\n"+divider+"world", "right", 1)
	want := "hello\nworld"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
	if strings.Contains(got, "[38;") {
		t.Fatalf("expected ANSI fragments to be stripped, got %q", got)
	}
}

func TestStripCopySelectionPreviewBorderBottom(t *testing.T) {
	got := stripCopySelectionPreviewBorder("━━━━\nhello\nworld", "bottom", 1)
	want := "hello\nworld"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRenderCopySelectionHighlightDoesNotLeakRawANSI(t *testing.T) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	highlight := lipgloss.NewStyle().Background(lipgloss.Color("8"))
	content := style.Render("copy me")

	got := renderCopySelectionHighlight(content, copySelectionBounds{x: 0, y: 0, width: 10, height: 1}, 0, 0, 3, 0, highlight)
	if strings.Contains(ansi.Strip(got), "[38;2;") {
		t.Fatalf("highlight leaked raw ANSI text: %q", got)
	}
	if ansi.Strip(got) != "copy me" {
		t.Fatalf("expected visible text to remain unchanged, got %q", ansi.Strip(got))
	}
}

func TestRenderCopySelectionHighlightPreservesUnselectedStyling(t *testing.T) {
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	highlight := lipgloss.NewStyle().Background(lipgloss.Color("8"))
	content := textStyle.Render("copy me")

	got := renderCopySelectionHighlight(content, copySelectionBounds{x: 0, y: 0, width: 10, height: 1}, 2, 0, 3, 0, highlight)
	prefix := ansi.Cut(got, 0, 2)
	suffix := ansi.Cut(got, 4, lipgloss.Width(got))

	if !strings.Contains(prefix, "\x1b[") {
		t.Fatalf("expected styled prefix to retain ANSI styling, got %q", prefix)
	}
	if !strings.Contains(suffix, "\x1b[") {
		t.Fatalf("expected styled suffix to retain ANSI styling, got %q", suffix)
	}
}

func TestRenderCopySelectionHighlightDoesNotHighlightRightDivider(t *testing.T) {
	highlight := lipgloss.NewStyle().Background(lipgloss.Color("8"))
	got := renderCopySelectionHighlight("│hello", copySelectionBounds{x: 0, y: 0, width: 6, height: 1}, 1, 0, 4, 0, highlight)

	if !strings.HasPrefix(got, "│\x1b") {
		t.Fatalf("expected divider to remain outside highlight, got %q", got)
	}
}

func TestRenderCopySelectionHighlightDoesNotHighlightBottomDivider(t *testing.T) {
	highlight := lipgloss.NewStyle().Background(lipgloss.Color("8"))
	got := renderCopySelectionHighlight("━━━━\nhello", copySelectionBounds{x: 0, y: 0, width: 5, height: 2}, 0, 1, 3, 1, highlight)
	lines := strings.Split(got, "\n")

	if lines[0] != "━━━━" {
		t.Fatalf("expected divider line to remain unchanged, got %q", lines[0])
	}
}
