package data

import "strings"

type SearchSort string

const (
	SearchSortUpdated SearchSort = "updated-desc"
	SearchSortCreated SearchSort = "created-desc"
)

func (s SearchSort) QueryValue() string {
	switch s {
	case SearchSortCreated:
		return string(SearchSortCreated)
	default:
		return string(SearchSortUpdated)
	}
}

func MakeSearchQuery(kind, query string, sort SearchSort) string {
	fields := strings.Fields(query)
	cleaned := make([]string, 0, len(fields))
	for _, field := range fields {
		if strings.HasPrefix(field, "sort:") {
			continue
		}
		cleaned = append(cleaned, field)
	}

	parts := []string{kind, "archived:false"}
	parts = append(parts, cleaned...)
	parts = append(parts, "sort:"+sort.QueryValue())
	return strings.Join(parts, " ")
}
