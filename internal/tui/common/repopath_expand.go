package common

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type ResolvedRepo struct {
	Name string
	Path string
}

func ExpandRepoPaths(cfgPaths map[string]string) []ResolvedRepo {
	keys := make([]string, 0, len(cfgPaths))
	for key := range cfgPaths {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	repos := make([]ResolvedRepo, 0, len(keys))
	seen := map[string]struct{}{}
	for _, key := range keys {
		if strings.HasSuffix(key, "/*") || key == ":owner/:repo" {
			continue
		}

		path := cfgPaths[key]
		path = cleanRepoPath(path)
		if !isDir(path) || !isGitRepo(path) {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		repos = append(repos, ResolvedRepo{Name: key, Path: path})
	}

	for _, key := range keys {
		path := cfgPaths[key]
		if key == ":owner/:repo" {
			continue
		}

		if strings.HasSuffix(key, "/*") {
			repos = append(repos, expandWildcardRepoPath(key, path, seen)...)
			continue
		}
	}

	return repos
}

func expandWildcardRepoPath(key, path string, seen map[string]struct{}) []ResolvedRepo {
	owner := strings.TrimSuffix(key, "/*")
	path = strings.TrimSuffix(ExpandRepoPath(path), string(filepath.Separator)+"*")
	path = strings.TrimSuffix(path, "/*")
	matches, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return nil
	}

	repos := make([]ResolvedRepo, 0, len(matches))
	for _, match := range matches {
		match = cleanRepoPath(match)
		if !isDir(match) || !isGitRepo(match) {
			continue
		}
		if _, ok := seen[match]; ok {
			continue
		}
		seen[match] = struct{}{}
		repos = append(repos, ResolvedRepo{
			Name: owner + "/" + filepath.Base(match),
			Path: match,
		})
	}
	return repos
}

func cleanRepoPath(path string) string {
	path = filepath.Clean(ExpandRepoPath(path))
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isGitRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}
