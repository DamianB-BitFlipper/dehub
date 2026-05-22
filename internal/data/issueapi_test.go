package data

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMakeIssuesQuerySorts(t *testing.T) {
	require.Equal(
		t,
		"is:issue archived:false assignee:@me sort:updated-desc",
		makeIssuesQuery("assignee:@me", SearchSortUpdated),
	)
	require.Equal(
		t,
		"is:issue archived:false assignee:@me sort:created-desc",
		makeIssuesQuery("assignee:@me", SearchSortCreated),
	)
	require.Equal(
		t,
		"is:issue archived:false assignee:@me sort:created-desc",
		makeIssuesQuery("assignee:@me sort:updated", SearchSortCreated),
	)
}
