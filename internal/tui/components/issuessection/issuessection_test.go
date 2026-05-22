package issuessection

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
)

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
