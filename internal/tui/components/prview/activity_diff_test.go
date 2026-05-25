package prview

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func TestParseReviewDiffHunk(t *testing.T) {
	diffHunk := `@@ -344,4 +344,4 @@ func committedReservationKey() string {
 return f"v1-sandbox:{sandbox_id}:committed-capacity-reservation"
-def old_key(self) -> str:
+def _committed_reservation_key_pattern(self) -> str:
     return "v1-sandbox:*:committed-capacity-reservation"`

	lines, err := parseReviewDiffHunk("capacity_store.py", diffHunk)
	require.NoError(t, err)
	require.Len(t, lines, 4)
	require.Equal(t, reviewDiffLine{OldLine: 344, NewLine: 344, Prefix: ' ', Text: "return f\"v1-sandbox:{sandbox_id}:committed-capacity-reservation\""}, lines[0])
	require.Equal(t, reviewDiffLine{OldLine: 345, Prefix: '-', Text: "def old_key(self) -> str:"}, lines[1])
	require.Equal(t, reviewDiffLine{NewLine: 345, Prefix: '+', Text: "def _committed_reservation_key_pattern(self) -> str:"}, lines[2])
	require.Equal(t, reviewDiffLine{OldLine: 346, NewLine: 346, Prefix: ' ', Text: "    return \"v1-sandbox:*:committed-capacity-reservation\""}, lines[3])
}

func TestRenderReviewDiffPreview(t *testing.T) {
	m := newDiffPreviewTestModel(t)
	preview := m.renderReviewDiffPreview("thread-1", "capacity_store.py", `@@ -344,2 +344,3 @@
 return f"v1-sandbox:{sandbox_id}:committed-capacity-reservation"
+def _committed_reservation_key_pattern(self) -> str:`, 80)

	plain := ansi.Strip(preview)
	require.Contains(t, plain, "344 344")
	require.Contains(t, plain, "+ def _committed_reservation_key_pattern")
	require.NotContains(t, plain, "@@")
}

func TestRenderReviewDiffPreviewEmptyWhenNoHunk(t *testing.T) {
	m := newDiffPreviewTestModel(t)
	require.Empty(t, m.renderReviewDiffPreview("thread-1", "capacity_store.py", "", 80))
}

func TestReviewDiffLinesCachesParsedHunk(t *testing.T) {
	m := newDiffPreviewTestModel(t)
	diffHunk := `@@ -344,1 +344,1 @@
 return f"v1-sandbox:{sandbox_id}:committed-capacity-reservation"`

	first := m.reviewDiffLines("thread-1", "capacity_store.py", diffHunk)
	second := m.reviewDiffLines("thread-1", "different.py", diffHunk)

	require.Len(t, first, 1)
	require.Equal(t, first, second)
	require.Len(t, m.reviewDiffCache, 1)
}

func newDiffPreviewTestModel(t *testing.T) *Model {
	t.Helper()
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../../../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	require.NoError(t, err)
	thm := theme.ParseTheme(&cfg)
	ctx := &context.ProgramContext{
		Config: &cfg,
		Theme:  thm,
		Styles: context.InitStyles(thm),
	}
	m := NewModel(ctx)
	m.UpdateProgramContext(ctx)
	m.SetWidth(100)
	return &m
}
