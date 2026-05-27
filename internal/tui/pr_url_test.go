package tui

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseGitHubPRURL(t *testing.T) {
	cases := []struct {
		name      string
		rawURL    string
		want      githubPRRef
		wantError bool
	}{
		{
			name:   "valid PR URL",
			rawURL: "https://github.com/PrimeIntellect-ai/platform/pull/2539",
			want:   githubPRRef{Owner: "PrimeIntellect-ai", Repo: "platform", Number: 2539},
		},
		{
			name:   "trailing slash",
			rawURL: "https://github.com/PrimeIntellect-ai/platform/pull/2539/",
			want:   githubPRRef{Owner: "PrimeIntellect-ai", Repo: "platform", Number: 2539},
		},
		{
			name:   "query and fragment",
			rawURL: "https://github.com/PrimeIntellect-ai/platform/pull/2539?foo=bar#discussion",
			want:   githubPRRef{Owner: "PrimeIntellect-ai", Repo: "platform", Number: 2539},
		},
		{
			name:      "wrong host",
			rawURL:    "https://example.com/PrimeIntellect-ai/platform/pull/2539",
			wantError: true,
		},
		{
			name:      "issue URL",
			rawURL:    "https://github.com/PrimeIntellect-ai/platform/issues/2539",
			wantError: true,
		},
		{
			name:      "invalid number",
			rawURL:    "https://github.com/PrimeIntellect-ai/platform/pull/not-a-number",
			wantError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseGitHubPRURL(tc.rawURL)
			if tc.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
			require.Equal(t, "repo:PrimeIntellect-ai/platform number:2539", got.searchQuery())
		})
	}
}
