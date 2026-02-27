package internal

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var githubRepoRefRE = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

// FindGitReposOutsideWorkspace returns repo paths from a comma-separated GIT_REPOS
// value that resolve outside the configured workspace.
func FindGitReposOutsideWorkspace(workspace, gitReposCSV string) []string {
	workspace = normalizeEnvPath(workspace)
	if workspace == "" || strings.TrimSpace(gitReposCSV) == "" {
		return nil
	}

	outside := make([]string, 0)
	seen := make(map[string]struct{})
	for _, entry := range strings.Split(gitReposCSV, ",") {
		repo := strings.TrimSpace(entry)
		if repo == "" {
			continue
		}
		// owner/repo references are remote GitHub repos, not local filesystem paths.
		// They should not participate in workspace path checks.
		if githubRepoRefRE.MatchString(repo) {
			continue
		}

		repo = normalizeEnvPath(repo)
		if repo == "" {
			continue
		}
		if !isPathWithinWorkspace(repo, workspace) {
			if _, ok := seen[repo]; ok {
				continue
			}
			seen[repo] = struct{}{}
			outside = append(outside, repo)
		}
	}
	return outside
}

func isPathWithinWorkspace(path, workspace string) bool {
	path = strings.TrimSpace(path)
	workspace = strings.TrimSpace(workspace)
	if path == "" || workspace == "" {
		return false
	}

	pathAbs := normalizeEnvPath(path)
	workspaceAbs := normalizeEnvPath(workspace)
	if pathAbs == "" || workspaceAbs == "" {
		return false
	}

	rel, err := filepath.Rel(workspaceAbs, pathAbs)
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && rel != "..")
}
