package data

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ActionsWorkflowRunsResponse struct {
	TotalCount   int           `json:"total_count"`
	WorkflowRuns []WorkflowRun `json:"workflow_runs"`
	PageInfo     PageInfo
}

type ActionsWorkflowsResponse struct {
	TotalCount int        `json:"total_count"`
	Workflows  []Workflow `json:"workflows"`
}

type Workflow struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	State     string `json:"state"`
	HtmlUrl   string `json:"html_url"`
	BadgeUrl  string `json:"badge_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	RepoName  string
}

func (workflow Workflow) GetRepoNameWithOwner() string { return workflow.RepoName }

func (workflow Workflow) GetTitle() string { return workflow.Name }

func (workflow Workflow) GetNumber() int { return int(workflow.Id) }

func (workflow Workflow) GetUrl() string { return workflow.HtmlUrl }

func (workflow Workflow) GetUpdatedAt() time.Time { return parseWorkflowTime(workflow.UpdatedAt) }

func FetchActionsWorkflows(filters string) (ActionsWorkflowsResponse, error) {
	repo := ActionsRepoFromFilters(filters)
	if repo == "" {
		return ActionsWorkflowsResponse{}, fmt.Errorf("actions sections require a repo:<owner>/<name> filter")
	}

	client, err := getRESTClient()
	if err != nil {
		return ActionsWorkflowsResponse{}, err
	}

	var response ActionsWorkflowsResponse
	if err := client.Get(fmt.Sprintf("repos/%s/actions/workflows?per_page=100", repo), &response); err != nil {
		return response, err
	}
	for i := range response.Workflows {
		response.Workflows[i].RepoName = repo
	}
	return response, nil
}

func FetchActionsWorkflowRuns(filters string, limit int, _ *PageInfo) (ActionsWorkflowRunsResponse, error) {
	repo, params := parseActionsWorkflowRunFilters(filters)
	if repo == "" {
		return ActionsWorkflowRunsResponse{}, fmt.Errorf("actions sections require a repo:<owner>/<name> filter")
	}

	if limit <= 0 {
		limit = 20
	}
	params.Set("per_page", strconv.Itoa(limit))

	client, err := getRESTClient()
	if err != nil {
		return ActionsWorkflowRunsResponse{}, err
	}

	path := fmt.Sprintf("repos/%s/actions/runs?%s", repo, params.Encode())
	var response ActionsWorkflowRunsResponse
	if err := client.Get(path, &response); err != nil {
		return response, err
	}
	response.PageInfo = PageInfo{HasNextPage: false}
	return response, nil
}

func FetchActionsWorkflowRunsForWorkflow(repo string, workflowID int64, limit int) (ActionsWorkflowRunsResponse, error) {
	if repo == "" {
		return ActionsWorkflowRunsResponse{}, fmt.Errorf("actions sections require a repo:<owner>/<name> filter")
	}
	if workflowID == 0 {
		return ActionsWorkflowRunsResponse{}, fmt.Errorf("actions workflow runs require a workflow id")
	}
	if limit <= 0 {
		limit = 20
	}

	client, err := getRESTClient()
	if err != nil {
		return ActionsWorkflowRunsResponse{}, err
	}

	path := fmt.Sprintf("repos/%s/actions/workflows/%d/runs?per_page=%d", repo, workflowID, limit)
	var response ActionsWorkflowRunsResponse
	if err := client.Get(path, &response); err != nil {
		return response, err
	}
	response.PageInfo = PageInfo{HasNextPage: false}
	return response, nil
}

func ActionsRepoFromFilters(filters string) string {
	repo, _ := parseActionsWorkflowRunFilters(filters)
	return repo
}

func parseActionsWorkflowRunFilters(filters string) (string, url.Values) {
	params := url.Values{}
	var repo string
	for _, token := range strings.Fields(filters) {
		key, value, ok := strings.Cut(token, ":")
		if !ok || value == "" {
			continue
		}
		switch key {
		case "repo":
			repo = value
		case "branch":
			params.Set("branch", value)
		case "event":
			params.Set("event", value)
		case "actor":
			if value != "@me" {
				params.Set("actor", value)
			}
		case "status", "is":
			params.Set("status", normalizeActionsStatus(value))
		case "workflow":
			params.Set("workflow", value)
		}
	}
	return repo, params
}

func parseWorkflowTime(value string) time.Time {
	t, _ := time.Parse(time.RFC3339, value)
	return t
}

func normalizeActionsStatus(status string) string {
	switch status {
	case "failure", "failed":
		return "failure"
	case "cancelled", "canceled":
		return "cancelled"
	case "in-progress":
		return "in_progress"
	default:
		return status
	}
}
