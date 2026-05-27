package tui

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type githubPRRef struct {
	Owner  string
	Repo   string
	Number int
}

func parseGitHubPRURL(raw string) (githubPRRef, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return githubPRRef{}, err
	}

	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return githubPRRef{}, fmt.Errorf("expected a GitHub URL")
	}
	if parsed.Hostname() != "github.com" {
		return githubPRRef{}, fmt.Errorf("expected a github.com URL")
	}

	parts := strings.Split(strings.Trim(parsed.EscapedPath(), "/"), "/")
	if len(parts) != 4 || parts[2] != "pull" {
		return githubPRRef{}, fmt.Errorf("expected a GitHub pull request URL")
	}

	number, err := strconv.Atoi(parts[3])
	if err != nil || number <= 0 {
		return githubPRRef{}, fmt.Errorf("expected a valid PR number")
	}

	owner, err := url.PathUnescape(parts[0])
	if err != nil {
		return githubPRRef{}, err
	}
	repo, err := url.PathUnescape(parts[1])
	if err != nil {
		return githubPRRef{}, err
	}
	if owner == "" || repo == "" {
		return githubPRRef{}, fmt.Errorf("expected a repository owner and name")
	}

	return githubPRRef{Owner: owner, Repo: repo, Number: number}, nil
}

func (r githubPRRef) searchQuery() string {
	return fmt.Sprintf("repo:%s/%s number:%d", r.Owner, r.Repo, r.Number)
}
