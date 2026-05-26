package common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpandRepoPathsExact(t *testing.T) {
	root := t.TempDir()
	repo := makeRepo(t, root, "repo")

	repos := ExpandRepoPaths(map[string]string{"owner/repo": repo})

	require.Equal(t, []ResolvedRepo{{Name: "owner/repo", Path: repo}}, repos)
}

func TestExpandRepoPathsWildcard(t *testing.T) {
	root := t.TempDir()
	one := makeRepo(t, root, "one")
	two := makeRepo(t, root, "two")
	require.NoError(t, os.MkdirAll(filepath.Join(root, "not-a-repo"), 0o755))

	repos := ExpandRepoPaths(map[string]string{"owner/*": filepath.Join(root, "*")})

	require.Equal(t, []ResolvedRepo{
		{Name: "owner/one", Path: one},
		{Name: "owner/two", Path: two},
	}, repos)
}

func TestExpandRepoPathsDeduplicatesByPath(t *testing.T) {
	root := t.TempDir()
	repo := makeRepo(t, root, "repo")

	repos := ExpandRepoPaths(map[string]string{
		"owner/*":    filepath.Join(root, "*"),
		"owner/repo": repo,
	})

	require.Equal(t, []ResolvedRepo{{Name: "owner/repo", Path: repo}}, repos)
}

func TestExpandRepoPathsSkipsTemplateAndMissingDirs(t *testing.T) {
	root := t.TempDir()

	repos := ExpandRepoPaths(map[string]string{
		":owner/:repo":  filepath.Join(root, ":owner", ":repo"),
		"owner/missing": filepath.Join(root, "missing"),
	})

	require.Empty(t, repos)
}

func TestExpandRepoPathsIncludesWorktrees(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "worktree")
	require.NoError(t, os.MkdirAll(path, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(path, ".git"), []byte("gitdir: ../.git/worktrees/worktree"), 0o644))

	repos := ExpandRepoPaths(map[string]string{"owner/worktree": path})

	require.Equal(t, []ResolvedRepo{{Name: "owner/worktree", Path: path}}, repos)
}

func makeRepo(t *testing.T, root, name string) string {
	t.Helper()
	path := filepath.Join(root, name)
	require.NoError(t, os.MkdirAll(filepath.Join(path, ".git"), 0o755))
	return path
}
