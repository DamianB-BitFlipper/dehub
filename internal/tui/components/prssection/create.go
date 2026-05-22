package prssection

import (
	"errors"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/git"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/common"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/fuzzyselect"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

type createPRCreatedMsg struct{}

type createPRBranchesFetchedMsg struct {
	RepoName  string
	RequestID uint64
	Branches  []fuzzyselect.Suggestion
	Head      string
	Base      string
	Err       error
}

var runCreatePRRepoCommand = common.RunRepoCommand

func (m *Model) validateCanCreatePR() error {
	repoName, ok := m.repoFromFilters()
	if !ok {
		return errors.New("current PR section must have exactly one repo:owner/name filter to create a PR")
	}

	if _, ok := common.GetRepoLocalPath(repoName, m.Ctx.Config.RepoPaths); !ok {
		return errors.New(
			"local path to repo not specified, set one in your config.yml under repoPaths",
		)
	}
	return nil
}

func (m *Model) prepareCreatePRForm() (tea.Cmd, error) {
	if err := m.validateCanCreatePR(); err != nil {
		return nil, err
	}
	repoName, _ := m.repoFromFilters()
	repoPath, _ := common.GetRepoLocalPath(repoName, m.Ctx.Config.RepoPaths)
	repoPath = common.ExpandRepoPath(repoPath)
	m.createPRBranchRequestID++
	requestID := m.createPRBranchRequestID
	m.CreatePRForm.SetBranchesLoading()
	return func() tea.Msg {
		repo, err := git.GetRepo(repoPath)
		if err != nil {
			return createPRBranchesFetchedMsg{RepoName: repoName, RequestID: requestID, Err: err}
		}

		branches := make([]fuzzyselect.Suggestion, 0, len(repo.Branches))
		base := ""
		for _, branch := range repo.Branches {
			detail := ""
			if branch.IsCheckedOut {
				detail = "current"
			}
			branches = append(branches, fuzzyselect.Suggestion{Value: branch.Name, Detail: detail})
			if base == "" && (branch.Name == "main" || branch.Name == "master") {
				base = branch.Name
			}
		}
		return createPRBranchesFetchedMsg{
			RepoName:  repoName,
			RequestID: requestID,
			Branches:  branches,
			Head:      repo.HeadBranchName,
			Base:      base,
		}
	}, nil
}

func (m *Model) createPR(title string, body string, head string, base string) (tea.Cmd, error) {
	if m.CreatePRForm.BranchesLoading() {
		return nil, errors.New("branches are still loading")
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, errors.New("PR title is required")
	}

	repoName, _ := m.repoFromFilters()
	repoPath, _ := common.GetRepoLocalPath(repoName, m.Ctx.Config.RepoPaths)

	taskId := fmt.Sprintf("create_pr_%s_%d", strings.ReplaceAll(repoName, "/", "_"), time.Now().Unix())
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Creating PR in %s", repoName),
		FinishedText: fmt.Sprintf("PR created in %s", repoName),
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		args := []string{"gh", "pr", "create", "--title", title, "--body", body}
		if head != "" {
			args = append(args, "--head", head)
		}
		if base != "" {
			args = append(args, "--base", base)
		}
		err := runCreatePRRepoCommand(repoPath, args...)
		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg:         createPRCreatedMsg{},
		}
	}), nil
}
