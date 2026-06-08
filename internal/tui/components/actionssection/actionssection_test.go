package actionssection

import (
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dehub/v4/internal/config"
	"github.com/dlvhdr/gh-dehub/v4/internal/data"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/theme"
)

func newActionsSectionTestContext() *context.ProgramContext {
	cfg := &config.Config{}
	cfg.Defaults.ActionsLimit = 20
	cfg.Theme = &config.ThemeConfig{}
	thm := *theme.DefaultTheme
	return &context.ProgramContext{
		Config: cfg,
		Theme:  thm,
		Styles: context.InitStyles(thm),
		StartTask: func(task context.Task) tea.Cmd {
			return func() tea.Msg { return nil }
		},
	}
}

func TestActionsRefreshPreservesSelectionFiltersAndFocus(t *testing.T) {
	ctx := newActionsSectionTestContext()
	m := NewModel(0, ctx, config.ActionsSectionConfig{
		Title:   "Actions",
		Filters: "repo:owner/repo",
	}, time.Now(), time.Now())

	m.Workflows = []data.Workflow{
		{Id: 1, Name: "Build", State: "active", RepoName: "owner/repo"},
		{Id: 2, Name: "Deploy", State: "active", RepoName: "owner/repo"},
	}
	m.Runs = []data.WorkflowRun{
		{Id: 10, DisplayTitle: "old first", UpdatedAt: time.Now().Add(-2 * time.Hour)},
		{Id: 20, DisplayTitle: "old selected", UpdatedAt: time.Now().Add(-time.Hour)},
	}
	m.Table.SetRows(m.BuildRows())
	m.Table.SetCurrItem(1)
	m.selectedWorkflow = &m.Workflows[1]
	m.RunsTable.SetRows(m.BuildRunRows())
	m.RunsTable.SetCurrItem(1)
	m.SearchValue = "repo:owner/repo branch:main"
	m.LocalSearchValue = ""
	m.SortOrder = data.SearchSortCreated
	m.IsFilteredByCurrentRemote = true
	m.SetFocusedPane(PaneDetails)

	next, cmd := m.Update(SectionActionsRefreshedMsg{
		Workflows: []data.Workflow{
			{Id: 1, Name: "Build", State: "active", RepoName: "owner/repo"},
			{Id: 2, Name: "Deploy updated", State: "active", RepoName: "owner/repo"},
		},
		TotalCount: 2,
		Runs: []data.WorkflowRun{
			{Id: 30, DisplayTitle: "new first", UpdatedAt: time.Now()},
			{Id: 20, DisplayTitle: "still selected", UpdatedAt: time.Now().Add(-time.Hour)},
		},
		RunsWorkflowID: 2,
		PageInfo:       data.PageInfo{HasNextPage: false},
		HasRuns:        true,
	})
	updated := next.(*Model)

	require.Nil(t, cmd)
	require.Equal(t, 1, updated.Table.GetCurrItem())
	require.NotNil(t, updated.selectedWorkflow)
	require.Equal(t, int64(2), updated.selectedWorkflow.Id)
	require.Equal(t, 1, updated.RunsTable.GetCurrItem())
	require.Equal(t, int64(20), updated.SelectedRun().Id)
	require.Equal(t, "repo:owner/repo branch:main", updated.SearchValue)
	require.Equal(t, "", updated.LocalSearchValue)
	require.Equal(t, data.SearchSortCreated, updated.SortOrder)
	require.True(t, updated.IsFilteredByCurrentRemote)
	require.Equal(t, PaneDetails, updated.FocusedPane())
}

func TestActionsRefreshDoesNotApplyRunsForMissingPreviousWorkflow(t *testing.T) {
	ctx := newActionsSectionTestContext()
	m := NewModel(0, ctx, config.ActionsSectionConfig{
		Title:   "Actions",
		Filters: "repo:owner/repo",
	}, time.Now(), time.Now())

	m.Workflows = []data.Workflow{
		{Id: 1, Name: "Build", State: "active", RepoName: "owner/repo"},
		{Id: 2, Name: "Deploy", State: "active", RepoName: "owner/repo"},
	}
	m.Runs = []data.WorkflowRun{{Id: 20, DisplayTitle: "old selected"}}
	m.Table.SetRows(m.BuildRows())
	m.Table.SetCurrItem(1)
	m.selectedWorkflow = &m.Workflows[1]
	m.RunsTable.SetRows(m.BuildRunRows())
	m.RunsTable.SetCurrItem(0)

	next, cmd := m.Update(SectionActionsRefreshedMsg{
		Workflows: []data.Workflow{
			{Id: 1, Name: "Build", State: "active", RepoName: "owner/repo"},
		},
		TotalCount:     1,
		Runs:           []data.WorkflowRun{{Id: 20, DisplayTitle: "stale deploy run"}},
		RunsTotalCount: 1,
		RunsWorkflowID: 2,
		PageInfo:       data.PageInfo{HasNextPage: false},
		HasRuns:        true,
	})
	updated := next.(*Model)

	require.NotNil(t, cmd)
	require.Equal(t, 0, updated.Table.GetCurrItem())
	require.NotNil(t, updated.selectedWorkflow)
	require.Equal(t, int64(1), updated.selectedWorkflow.Id)
	require.Empty(t, updated.Runs)
	require.Empty(t, updated.RunsTable.Rows)
	require.True(t, updated.RunsTable.IsLoading())
}

// TestActionsRefreshResetsSelectionWhenWorkflowGone verifies that when the
// previously selected workflow disappears after a refresh and the list
// reorders, the selection resets to the top rather than landing on whatever
// unrelated workflow now occupies the old cursor index.
func TestActionsRefreshResetsSelectionWhenWorkflowGone(t *testing.T) {
	ctx := newActionsSectionTestContext()
	m := NewModel(0, ctx, config.ActionsSectionConfig{
		Title:   "Actions",
		Filters: "repo:owner/repo",
	}, time.Now(), time.Now())

	m.Workflows = []data.Workflow{
		{Id: 1, Name: "Build", State: "active", RepoName: "owner/repo"},
		{Id: 2, Name: "Deploy", State: "active", RepoName: "owner/repo"},
		{Id: 3, Name: "Release", State: "active", RepoName: "owner/repo"},
	}
	m.Table.SetRows(m.BuildRows())
	// Select the middle workflow (Id 2) at cursor index 1.
	m.Table.SetCurrItem(1)
	m.selectedWorkflow = &m.Workflows[1]

	// Refresh: Id 2 is gone and the list is reordered so index 1 now holds a
	// different workflow (Id 4) than the user had selected.
	next, _ := m.Update(SectionActionsRefreshedMsg{
		Workflows: []data.Workflow{
			{Id: 3, Name: "Release", State: "active", RepoName: "owner/repo"},
			{Id: 4, Name: "Lint", State: "active", RepoName: "owner/repo"},
		},
		TotalCount: 2,
		HasRuns:    false,
	})
	updated := next.(*Model)

	require.Equal(t, 0, updated.Table.GetCurrItem(),
		"selection must reset to top when the selected workflow is gone")
	require.NotNil(t, updated.selectedWorkflow)
	require.Equal(t, int64(3), updated.selectedWorkflow.Id,
		"must not drift to the workflow now occupying the stale cursor index (Id 4)")
}

func TestActionsRefreshFetchesRunsForNewlySelectedWorkflow(t *testing.T) {
	ctx := newActionsSectionTestContext()
	m := NewModel(0, ctx, config.ActionsSectionConfig{
		Title:   "Actions",
		Filters: "repo:owner/repo",
	}, time.Now(), time.Now())

	next, cmd := m.Update(SectionActionsRefreshedMsg{
		Workflows: []data.Workflow{
			{Id: 1, Name: "Build", State: "active", RepoName: "owner/repo"},
		},
		TotalCount: 1,
		HasRuns:    false,
	})
	updated := next.(*Model)

	require.NotNil(t, cmd)
	require.Equal(t, 0, updated.Table.GetCurrItem())
	require.NotNil(t, updated.selectedWorkflow)
	require.Equal(t, int64(1), updated.selectedWorkflow.Id)
	require.Empty(t, updated.Runs)
	require.Empty(t, updated.RunsTable.Rows)
	require.True(t, updated.RunsTable.IsLoading())
}

func TestActionsRefreshSectionRowsFetchesWorkflowsAndSelectedRuns(t *testing.T) {
	oldFetchWorkflows := fetchActionsWorkflowsForSectionRefresh
	oldFetchRuns := fetchActionsWorkflowRunsForSectionRefresh
	t.Cleanup(func() {
		fetchActionsWorkflowsForSectionRefresh = oldFetchWorkflows
		fetchActionsWorkflowRunsForSectionRefresh = oldFetchRuns
	})

	fetchActionsWorkflowsForSectionRefresh = func(filters string) (data.ActionsWorkflowsResponse, error) {
		require.Equal(t, "repo:owner/repo", filters)
		return data.ActionsWorkflowsResponse{
			TotalCount: 1,
			Workflows:  []data.Workflow{{Id: 2, Name: "Deploy", RepoName: "owner/repo"}},
		}, nil
	}
	fetchActionsWorkflowRunsForSectionRefresh = func(repo string, workflowID int64, limit int) (data.ActionsWorkflowRunsResponse, error) {
		require.Equal(t, "owner/repo", repo)
		require.Equal(t, int64(2), workflowID)
		require.Equal(t, 20, limit)
		return data.ActionsWorkflowRunsResponse{
			TotalCount:   1,
			WorkflowRuns: []data.WorkflowRun{{Id: 20, DisplayTitle: "selected"}},
			PageInfo:     data.PageInfo{HasNextPage: false},
		}, nil
	}

	ctx := newActionsSectionTestContext()
	m := NewModel(3, ctx, config.ActionsSectionConfig{
		Title:   "Actions",
		Filters: "repo:owner/repo",
	}, time.Now(), time.Now())
	m.selectedWorkflow = &data.Workflow{Id: 2, Name: "Deploy", RepoName: "owner/repo"}

	cmd := m.RefreshSectionRows()
	require.NotNil(t, cmd)
	msg, ok := cmd().(SectionActionsRefreshedMsg)
	require.True(t, ok)
	require.Equal(t, 3, msg.SectionId)
	require.Equal(t, 1, msg.TotalCount)
	require.True(t, msg.HasRuns)
	require.Equal(t, int64(2), msg.RunsWorkflowID)
	require.Len(t, msg.Runs, 1)
}

func TestSetRefreshingMarksWorkflowAndRunsLoading(t *testing.T) {
	ctx := newActionsSectionTestContext()
	m := NewModel(0, ctx, config.ActionsSectionConfig{
		Title:   "Actions",
		Filters: "repo:owner/repo",
	}, time.Now(), time.Now())

	// No selected workflow: only the workflow table should be loading.
	m.SetRefreshing()
	require.True(t, m.GetIsLoading())
	require.False(t, m.RunsTable.IsLoading())

	// With a selected workflow: the runs table should also be loading.
	m.SetIsLoading(false)
	m.selectedWorkflow = &data.Workflow{Id: 2, Name: "Deploy", RepoName: "owner/repo"}
	m.SetRefreshing()
	require.True(t, m.GetIsLoading())
	require.True(t, m.RunsTable.IsLoading())
}

func TestSetRefreshingClearsLocalSearch(t *testing.T) {
	ctx := newActionsSectionTestContext()
	m := NewModel(0, ctx, config.ActionsSectionConfig{
		Title:   "Actions",
		Filters: "repo:owner/repo",
	}, time.Now(), time.Now())
	m.Workflows = []data.Workflow{
		{Id: 1, Name: "Build", State: "active", RepoName: "owner/repo"},
		{Id: 2, Name: "Deploy", State: "active", RepoName: "owner/repo"},
	}
	m.Table.SetRows(m.BuildRows())
	m.LocalSearchValue = "deploy"
	require.Len(t, m.filteredWorkflows(), 1, "precondition: local search filters rows")

	m.SetRefreshing()

	require.Equal(t, "", m.LocalSearchValue, "manual refresh must clear local search")
	require.Len(t, m.filteredWorkflows(), 2, "all rows visible after local search cleared")
}

func TestRefreshSectionRowsDedupsWhileInFlight(t *testing.T) {
	oldFetchWorkflows := fetchActionsWorkflowsForSectionRefresh
	t.Cleanup(func() { fetchActionsWorkflowsForSectionRefresh = oldFetchWorkflows })

	var calls int
	fetchActionsWorkflowsForSectionRefresh = func(string) (data.ActionsWorkflowsResponse, error) {
		calls++
		return data.ActionsWorkflowsResponse{
			TotalCount: 1,
			Workflows:  []data.Workflow{{Id: 1, Name: "Build", RepoName: "owner/repo"}},
		}, nil
	}

	ctx := newActionsSectionTestContext()
	m := NewModel(0, ctx, config.ActionsSectionConfig{
		Title:   "Actions",
		Filters: "repo:owner/repo",
	}, time.Now(), time.Now())

	// First refresh starts and is in flight.
	cmd1 := m.RefreshSectionRows()
	require.NotNil(t, cmd1)
	// Second refresh while in flight is suppressed.
	require.Nil(t, m.RefreshSectionRows(), "concurrent refresh must be deduped")

	// Resolve the first refresh; its result clears the in-flight flag.
	msg := cmd1()
	refreshed, ok := msg.(SectionActionsRefreshedMsg)
	require.True(t, ok)
	next, _ := m.Update(refreshed)
	m = *next.(*Model)

	// A subsequent refresh is allowed again.
	cmd2 := m.RefreshSectionRows()
	require.NotNil(t, cmd2, "refresh allowed again after the previous one landed")
	cmd2()
	require.Equal(t, 2, calls, "exactly two fetches: deduped middle press did not fetch")
}

func TestRefreshFailedMsgClearsInFlightAndLoading(t *testing.T) {
	oldFetchWorkflows := fetchActionsWorkflowsForSectionRefresh
	t.Cleanup(func() { fetchActionsWorkflowsForSectionRefresh = oldFetchWorkflows })

	fetchActionsWorkflowsForSectionRefresh = func(string) (data.ActionsWorkflowsResponse, error) {
		return data.ActionsWorkflowsResponse{}, errors.New("boom")
	}

	ctx := newActionsSectionTestContext()
	m := NewModel(0, ctx, config.ActionsSectionConfig{
		Title:   "Actions",
		Filters: "repo:owner/repo",
	}, time.Now(), time.Now())
	m.SetRefreshing()

	cmd := m.RefreshSectionRows()
	require.NotNil(t, cmd)
	failed, ok := cmd().(SectionActionsRefreshFailedMsg)
	require.True(t, ok)
	require.Error(t, failed.Err)

	next, _ := m.Update(failed)
	m = *next.(*Model)

	require.False(t, m.GetIsLoading(), "loading cleared on failure")
	require.NotNil(t, m.RefreshSectionRows(), "refresh allowed again after failure")
}
