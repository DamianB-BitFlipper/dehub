package prssection

import (
	tea "charm.land/bubbletea/v2"

	"github.com/dlvhdr/gh-dehub/v4/internal/tui/common"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/components/prrow"
)

func (m Model) diff() tea.Cmd {
	currRowData := m.GetCurrRow()
	if currRowData == nil {
		return nil
	}
	baseRefName := ""
	headRefName := ""
	if pr, ok := currRowData.(*prrow.Data); ok && pr.Primary != nil {
		baseRefName = pr.Primary.BaseRefName
		headRefName = pr.Primary.HeadRefName
	}

	return common.DiffPR(
		currRowData.GetNumber(),
		currRowData.GetRepoNameWithOwner(),
		currRowData.GetTitle(),
		currRowData.GetUrl(),
		baseRefName,
		headRefName,
		m.Ctx.Config.Pager.Diff,
		m.Ctx.Config.RunDiffPagerInBackground(),
		m.Ctx.Config.GetFullScreenDiffPagerEnv(),
	)
}
