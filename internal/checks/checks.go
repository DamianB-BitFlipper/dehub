package checks

import ghchecks "github.com/dlvhdr/x/gh-checks"

type Category int

const (
	CategorySuccess Category = iota
	CategorySkipped
	CategoryFailure
	CategoryCancelled
	CategoryPending
	CategoryNeutral
)

type Stats struct {
	Succeeded        int
	Neutral          int
	Failed           int
	Skipped          int
	InProgress       int
	AwaitingApproval int
}

func CategoryForState(state string) Category {
	if ghchecks.IsStatusWaiting(state) {
		return CategoryPending
	}
	if state == "CANCELLED" {
		return CategoryCancelled
	}
	if ghchecks.IsConclusionAFailure(state) {
		return CategoryFailure
	}
	if ghchecks.IsConclusionASkip(state) {
		return CategorySkipped
	}
	if ghchecks.IsConclusionNeutral(state) {
		return CategoryNeutral
	}
	if ghchecks.IsConclusionASuccess(state) {
		return CategorySuccess
	}
	return CategoryPending
}

func CategoryForCheckRun(status string, conclusion string) Category {
	if ghchecks.IsStatusWaiting(status) {
		return CategoryPending
	}
	return CategoryForState(conclusion)
}

func AddStateCount(stats *Stats, state string, count int) {
	switch CategoryForState(state) {
	case CategoryPending:
		stats.InProgress += count
	case CategoryFailure:
		stats.Failed += count
	case CategorySkipped:
		stats.Skipped += count
	case CategoryNeutral:
		stats.Neutral += count
	case CategorySuccess:
		stats.Succeeded += count
	}
}
