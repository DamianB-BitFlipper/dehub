package issuessection

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dehub/v4/internal/config"
	"github.com/dlvhdr/gh-dehub/v4/internal/data"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/theme"
)

func newIssueSectionToggleTestContext() *context.ProgramContext {
	cfg := &config.Config{}
	cfg.Defaults.IssuesLimit = 30
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

func TestToggleOpenClosedKeyUpdatesIssueFilters(t *testing.T) {
	m := NewModel(
		1,
		newIssueSectionToggleTestContext(),
		config.IssuesSectionConfig{Title: "Test", Filters: "repo:owner/repo is:closed author:@me"},
		time.Now(),
		time.Now(),
	)
	m.Issues = []data.IssueData{{Number: 1, State: "CLOSED"}}
	m.Table.SetRows(m.BuildRows())

	next, cmd := m.Update(tea.KeyPressMsg{Text: "T", Code: 'T'})
	updated := next.(*Model)

	require.NotNil(t, cmd)
	require.Equal(t, "repo:owner/repo is:open author:@me", updated.SearchValue)
	require.Equal(t, updated.SearchValue, updated.SearchBar.Value())
	require.Nil(t, updated.Issues)
	require.Nil(t, updated.PageInfo)
}

func TestToggleOpenClosedKeyNoopsWithoutIssueStateFilter(t *testing.T) {
	m := NewModel(
		1,
		newIssueSectionToggleTestContext(),
		config.IssuesSectionConfig{Title: "Test", Filters: "repo:owner/repo author:@me"},
		time.Now(),
		time.Now(),
	)

	next, cmd := m.Update(tea.KeyPressMsg{Text: "T", Code: 'T'})
	updated := next.(*Model)

	require.Nil(t, cmd)
	require.Equal(t, "repo:owner/repo author:@me", updated.SearchValue)
}

func TestSortIssuesUsesLoadedRows(t *testing.T) {
	oldCreated := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	newCreated := oldCreated.Add(time.Hour)
	oldUpdated := oldCreated.Add(2 * time.Hour)
	newUpdated := oldCreated.Add(3 * time.Hour)
	m := Model{
		BaseModel: section.BaseModel{SortOrder: data.SearchSortUpdated},
		Issues: []data.IssueData{
			{Number: 1, CreatedAt: newCreated, UpdatedAt: oldUpdated},
			{Number: 2, CreatedAt: oldCreated, UpdatedAt: newUpdated},
		},
	}

	m.sortIssues()
	require.Equal(t, 2, m.Issues[0].Number)

	m.ToggleSortOrder()
	m.sortIssues()
	require.Equal(t, 1, m.Issues[0].Number)
}

func TestLocalSearchFiltersIssuesByTitleNumberAndRepo(t *testing.T) {
	first := data.IssueData{Number: 123, Title: "Math is broken", State: "OPEN"}
	first.Repository.Name = "calculator"
	first.Repository.NameWithOwner = "owner/calculator"
	first.Author.Login = "alice"
	second := data.IssueData{Number: 456, Title: "Docs typo", State: "OPEN"}
	second.Repository.Name = "docs"
	second.Repository.NameWithOwner = "owner/docs"
	second.Author.Login = "bob"
	m := Model{Issues: []data.IssueData{first, second}}

	m.LocalSearchValue = "math"
	require.Len(t, m.filteredIssues(), 1)
	require.Equal(t, 123, m.GetCurrRow().(*data.IssueData).Number)

	m.LocalSearchValue = "#456"
	require.Len(t, m.filteredIssues(), 1)
	require.Equal(t, 456, m.filteredIssues()[0].Number)

	m.LocalSearchValue = "calculator"
	require.Len(t, m.filteredIssues(), 1)
	require.Equal(t, 123, m.filteredIssues()[0].Number)
}
