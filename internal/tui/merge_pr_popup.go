package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dehub/v4/internal/data"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dehub/v4/internal/tui/components/tasks"
)

type mergePRPopup struct {
	section      tasks.SectionIdentifier
	pr           data.RowData
	methodIndex  int
	auto         bool
	deleteBranch bool
	retargetPRs  []tasks.BranchRetarget
}

var mergePRMethods = []struct {
	label  string
	method tasks.MergeMethod
}{
	{label: "Create a merge commit", method: tasks.MergeMethodMerge},
	{label: "Squash and merge", method: tasks.MergeMethodSquash},
	{label: "Rebase and merge", method: tasks.MergeMethodRebase},
}

func newMergePRPopup(section tasks.SectionIdentifier, pr data.RowData, retargetPRs []tasks.BranchRetarget) *mergePRPopup {
	return &mergePRPopup{section: section, pr: pr, methodIndex: 1, deleteBranch: true, retargetPRs: retargetPRs}
}

func (p *mergePRPopup) options() tasks.MergePROptions {
	branchRef := ""
	if pr, ok := p.pr.(branchRefData); ok {
		branchRef = pr.GetHeadRefName()
	}

	return tasks.MergePROptions{
		Method:       mergePRMethods[p.methodIndex].method,
		Auto:         p.auto,
		DeleteBranch: p.deleteBranch,
		RetargetPRs:  p.retargetPRs,
		BranchRef:    branchRef,
	}
}

func (p *mergePRPopup) move(delta int) {
	p.methodIndex = (p.methodIndex + delta + len(mergePRMethods)) % len(mergePRMethods)
}

func (m *Model) openMergePRPopup(section tasks.SectionIdentifier, pr data.RowData, retargetPRs ...tasks.BranchRetarget) {
	if pr == nil {
		return
	}
	m.mergePRPopup = newMergePRPopup(section, pr, retargetPRs)
}

type branchRefData interface {
	data.RowData
	GetHeadRefName() string
	GetBaseRefName() string
}

func mergePRRetargets(pr data.RowData, prs []prrow.Data) []tasks.BranchRetarget {
	selected, ok := pr.(branchRefData)
	if !ok {
		return nil
	}
	selectedPR, ok := pr.(*prrow.Data)
	if !ok || selectedPR.Primary == nil || !isSameRepoHead(selectedPR.Primary) {
		return nil
	}

	headRefName := selected.GetHeadRefName()
	baseRefName := selected.GetBaseRefName()
	repoName := selected.GetRepoNameWithOwner()
	if headRefName == "" || baseRefName == "" || repoName == "" {
		return nil
	}

	retargets := make([]tasks.BranchRetarget, 0)
	seen := map[int]struct{}{}
	for _, candidate := range prs {
		if candidate.Primary == nil {
			continue
		}
		if candidate.Primary.Number == selected.GetNumber() || candidate.Primary.Repository.NameWithOwner != repoName {
			continue
		}
		if candidate.Primary.BaseRefName != headRefName || candidate.Primary.BaseRefName == baseRefName {
			continue
		}
		if _, ok := seen[candidate.Primary.Number]; ok {
			continue
		}
		seen[candidate.Primary.Number] = struct{}{}
		retargets = append(retargets, tasks.BranchRetarget{
			Number:      candidate.Primary.Number,
			RepoName:    repoName,
			BaseRefName: baseRefName,
		})
	}
	return retargets
}

func isSameRepoHead(pr *data.PullRequestData) bool {
	return pr.HeadRepository.Name == pr.Repository.Name && pr.HeadRepository.Owner.Login == pr.Repository.Owner.Login
}

func (m *Model) updateMergePRPopup(msg tea.KeyMsg) tea.Cmd {
	if m.mergePRPopup == nil {
		return nil
	}

	switch msg.String() {
	case "esc", "ctrl+c", "q":
		m.mergePRPopup = nil
	case "up":
		m.mergePRPopup.move(-1)
	case "down":
		m.mergePRPopup.move(1)
	case "m":
		m.mergePRPopup.methodIndex = 0
	case "s":
		m.mergePRPopup.methodIndex = 1
	case "r":
		m.mergePRPopup.methodIndex = 2
	case "a":
		m.mergePRPopup.auto = !m.mergePRPopup.auto
	case "d":
		m.mergePRPopup.deleteBranch = !m.mergePRPopup.deleteBranch
	case "enter":
		popup := m.mergePRPopup
		m.mergePRPopup = nil
		return tasks.MergePRWithOptions(m.ctx, popup.section, popup.pr, popup.options())
	}

	return nil
}

func (m Model) renderMergePRPopup() string {
	if m.mergePRPopup == nil || m.mergePRPopup.pr == nil {
		return ""
	}

	width := min(max(52, m.ctx.ScreenWidth-10), 90)
	contentWidth := width - 4
	accent := m.ctx.Theme.PrimaryBorder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.ctx.Theme.PrimaryText).
		Render("Merge PR")

	pr := m.mergePRPopup.pr
	subtitle := lipgloss.NewStyle().
		Width(contentWidth).
		Foreground(m.ctx.Theme.FaintText).
		Render(fmt.Sprintf("#%d %s", pr.GetNumber(), pr.GetTitle()))

	methodLines := make([]string, 0, len(mergePRMethods))
	for i, method := range mergePRMethods {
		marker := "  "
		style := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText)
		if i == m.mergePRPopup.methodIndex {
			marker = "> "
			style = lipgloss.NewStyle().Foreground(m.ctx.Theme.PrimaryText).Bold(true)
		}
		methodLines = append(methodLines, style.Render(marker+method.label))
	}

	auto := checkbox(m.mergePRPopup.auto) + " Enable auto-merge"
	deleteBranch := checkbox(m.mergePRPopup.deleteBranch) + " Delete branch after merge"
	options := lipgloss.NewStyle().
		Foreground(m.ctx.Theme.PrimaryText).
		Render(strings.Join([]string{auto, deleteBranch}, "\n"))

	hint := lipgloss.NewStyle().
		Foreground(m.ctx.Theme.FaintText).
		Render("up/down choose method | a auto | d delete branch | enter merge | esc cancel")

	return lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Background(m.ctx.Theme.SelectedBackground).
		Padding(1, 2).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			subtitle,
			"",
			strings.Join(methodLines, "\n"),
			"",
			options,
			"",
			hint,
		))
}

func checkbox(checked bool) string {
	if checked {
		return "[x]"
	}
	return "[ ]"
}
