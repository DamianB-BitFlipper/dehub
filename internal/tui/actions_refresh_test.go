package tui

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dehub/v4/internal/config"
	"github.com/dlvhdr/gh-dehub/v4/internal/data"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/components/actionssection"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/components/footer"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/components/issueview"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/components/notificationview"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/components/prview"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/components/sidebar"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/components/tabs"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/keys"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/theme"
)

func newActionsRefreshTestModel(t *testing.T) Model {
	t.Helper()

	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	require.NoError(t, err)

	ctx := &context.ProgramContext{
		Config:            &cfg,
		View:              config.ActionsView,
		ScreenWidth:       200,
		ScreenHeight:      40,
		MainContentWidth:  200,
		MainContentHeight: 30,
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)
	ctx.StartTask = func(context.Task) tea.Cmd {
		return func() tea.Msg { return nil }
	}

	sec0 := actionssection.NewModel(0, ctx, config.ActionsSectionConfig{Title: "A"}, time.Now(), time.Now())
	sec0.Workflows = []data.Workflow{{Id: 1, Name: "original-0", State: "active"}}
	sec0.TotalCount = 1
	sec0.Table.SetRows(sec0.BuildRows())

	sec1 := actionssection.NewModel(1, ctx, config.ActionsSectionConfig{Title: "B"}, time.Now(), time.Now())
	sec1.Workflows = []data.Workflow{{Id: 2, Name: "original-1", State: "active"}}
	sec1.TotalCount = 1
	sec1.Table.SetRows(sec1.BuildRows())

	m := Model{
		ctx:                ctx,
		keys:               keys.Keys,
		actions:            []section.Section{&sec0, &sec1},
		sidebar:            sidebar.NewModel(),
		prView:             prview.NewModel(ctx),
		issueSidebar:       issueview.NewModel(ctx),
		notificationView:   notificationview.NewModel(ctx),
		footer:             footer.NewModel(ctx),
		tabs:               tabs.NewModel(ctx),
		activePane:         mainPane,
		prPreviewStates:    map[string]map[int]int{},
		issuePreviewStates: map[string]int{},
		viewStates:         map[config.ViewType]*viewState{},
	}
	m.footer.UpdateProgramContext(ctx)
	return m
}

// TestManualActionsRefreshAppliesToOriginatingSection verifies a bare
// SectionActionsRefreshedMsg is routed back to the section identified by its
// SectionId, even when a different section is currently focused. This guards
// against the manual-refresh result bleeding into the wrong section when the
// user switches sections before the async fetch completes.
func TestManualActionsRefreshAppliesToOriginatingSection(t *testing.T) {
	m := newActionsRefreshTestModel(t)
	// Focus section 0, but the refresh result targets section 1.
	m.currSectionId = 0

	updated, _ := m.Update(actionssection.SectionActionsRefreshedMsg{
		SectionId:  1,
		Workflows:  []data.Workflow{{Id: 9, Name: "refreshed-1", State: "active"}},
		TotalCount: 1,
	})
	um := updated.(Model)

	sec0 := um.actions[0].(*actionssection.Model)
	sec1 := um.actions[1].(*actionssection.Model)

	require.Equal(t, "refreshed-1", sec1.Workflows[0].Name,
		"refresh result must apply to the originating section (id 1)")
	require.Equal(t, "original-0", sec0.Workflows[0].Name,
		"the currently focused section (id 0) must not be mutated")
}

// TestManualActionsRefreshUpdatesCurrentSection verifies the result applies to
// the originating section when it also happens to be the current section.
func TestManualActionsRefreshUpdatesCurrentSection(t *testing.T) {
	m := newActionsRefreshTestModel(t)
	m.currSectionId = 1

	updated, _ := m.Update(actionssection.SectionActionsRefreshedMsg{
		SectionId:  1,
		Workflows:  []data.Workflow{{Id: 9, Name: "refreshed-1", State: "active"}},
		TotalCount: 1,
	})
	um := updated.(Model)

	sec1 := um.actions[1].(*actionssection.Model)
	require.Equal(t, "refreshed-1", sec1.Workflows[0].Name)
}
