package tui

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dehub/v4/internal/data"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/components/tasks"
)

type mergePopupTestRow struct {
	headRef string
	baseRef string
}

func (mergePopupTestRow) GetRepoNameWithOwner() string { return "owner/repo" }
func (mergePopupTestRow) GetTitle() string             { return "Add popup merge" }
func (mergePopupTestRow) GetNumber() int               { return 42 }
func (mergePopupTestRow) GetUrl() string               { return "https://github.com/owner/repo/pull/42" }
func (mergePopupTestRow) GetUpdatedAt() time.Time      { return time.Time{} }
func (r mergePopupTestRow) GetHeadRefName() string     { return r.headRef }
func (r mergePopupTestRow) GetBaseRefName() string     { return r.baseRef }

func TestMergePRPopupRendersDefaults(t *testing.T) {
	m := newMessagePopupTestModel(t)
	m.openMergePRPopup(tasks.SectionIdentifier{Id: 2, Type: "pr"}, mergePopupTestRow{})

	view := ansi.Strip(m.renderMergePRPopup())

	require.Contains(t, view, "Merge PR")
	require.Contains(t, view, "#42 Add popup merge")
	require.Contains(t, view, "> Squash and merge")
	require.Contains(t, view, "[ ] Enable auto-merge")
	require.Contains(t, view, "[x] Delete branch after merge")
}

func TestMergePRPopupUpdatesOptions(t *testing.T) {
	m := newMessagePopupTestModel(t)
	m.openMergePRPopup(tasks.SectionIdentifier{Id: 2, Type: "pr"}, mergePopupTestRow{})

	cmd := m.updateMergePRPopup(tea.KeyPressMsg{Text: "r"})
	require.Nil(t, cmd)
	cmd = m.updateMergePRPopup(tea.KeyPressMsg{Text: "a"})
	require.Nil(t, cmd)
	cmd = m.updateMergePRPopup(tea.KeyPressMsg{Text: "d"})
	require.Nil(t, cmd)

	require.Equal(t, tasks.MergeMethodRebase, m.mergePRPopup.options().Method)
	require.True(t, m.mergePRPopup.options().Auto)
	require.False(t, m.mergePRPopup.options().DeleteBranch)
}

func TestMergePRPopupDismissesOnEscape(t *testing.T) {
	m := newMessagePopupTestModel(t)
	m.openMergePRPopup(tasks.SectionIdentifier{Id: 2, Type: "pr"}, mergePopupTestRow{})

	updated, cmd := m.Update(tea.KeyPressMsg{Text: "esc"})
	m = updated.(Model)

	require.Nil(t, cmd)
	require.Nil(t, m.mergePRPopup)
}

func TestMergePRRetargetsDependentPRsInSameRepo(t *testing.T) {
	selected := mergePopupPR(42, "feature/base", "main", "owner", "repo")
	prs := []prrow.Data{
		mergePopupPR(43, "dependent", "feature/base", "owner", "repo"),
		mergePopupPR(44, "other", "other", "owner", "repo"),
		mergePopupPR(45, "dependent", "feature/base", "other", "repo"),
	}

	retargets := mergePRRetargets(&selected, prs)

	require.Equal(t, []tasks.BranchRetarget{
		{Number: 43, RepoName: "owner/repo", BaseRefName: "main"},
	}, retargets)
}

func TestMergePRRetargetsIgnoreForkBranchNameCollision(t *testing.T) {
	selected := mergePopupPR(42, "feature/base", "main", "owner", "repo")
	selected.Primary.HeadRepository.Owner.Login = "fork-owner"
	prs := []prrow.Data{mergePopupPR(43, "dependent", "feature/base", "owner", "repo")}

	retargets := mergePRRetargets(&selected, prs)

	require.Empty(t, retargets)
}

func TestMergePRPopupOptionsIncludeRetargets(t *testing.T) {
	m := newMessagePopupTestModel(t)
	retargets := []tasks.BranchRetarget{{Number: 43, RepoName: "owner/repo", BaseRefName: "main"}}
	m.openMergePRPopup(tasks.SectionIdentifier{Id: 2, Type: "pr"}, mergePopupTestRow{}, retargets...)

	require.Equal(t, retargets, m.mergePRPopup.options().RetargetPRs)
}

func mergePopupPR(number int, headRef string, baseRef string, owner string, repo string) prrow.Data {
	return prrow.Data{Primary: &data.PullRequestData{
		Number:      number,
		Title:       "Add popup merge",
		Url:         "https://github.com/owner/repo/pull/42",
		HeadRefName: headRef,
		BaseRefName: baseRef,
		Repository: data.Repository{
			Name:          repo,
			NameWithOwner: owner + "/" + repo,
			Owner:         data.Owner{Login: owner},
		},
		HeadRepository: struct {
			Name  string
			Owner data.Owner
		}{Name: repo, Owner: data.Owner{Login: owner}},
	}}
}
