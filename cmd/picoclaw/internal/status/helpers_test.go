package status

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestBuildSkillsEnvHealthReport(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(workspace, "skills", "todoist-manager"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(workspace, "skills", "git-summary"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "skills", "todoist-manager", "SKILL.md"), []byte("# test"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "skills", "git-summary", "SKILL.md"), []byte("# test"), 0o644))

	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Workspace = workspace
	cfg.Agents.Defaults.RestrictToWorkspace = true

	report := buildSkillsEnvHealthReport(
		cfg,
		[]string{"/home/test/.picoclaw/.env.picoclaw"},
		func(file string) (string, error) {
			switch file {
			case "git", "curl":
				return "/usr/bin/" + file, nil
			default:
				return "", errors.New("missing")
			}
		},
		func(key string) (string, bool) {
			switch key {
			case "TODOIST_API_TOKEN":
				return "token", true
			case "GIT_REPOS":
				return filepath.Join(t.TempDir(), "outside-repo"), true
			default:
				return "", false
			}
		},
	)

	require.Contains(t, report, "Skills Env Health:")
	require.Contains(t, report, "Env files loaded:")
	require.Contains(t, report, "todoist-manager")
	require.Contains(t, report, "git-summary")
	require.Contains(t, report, "TODOIST_API_TOKEN: ✓")
	require.Contains(t, report, "GIT_REPOS: ✓")
	require.Contains(t, report, "git: ✓")
	require.Contains(t, report, "jq: missing")
	require.Contains(t, report, "python3: missing")
	require.Contains(t, report, "Warning: restrict_to_workspace=true may block")
}

func TestBuildSkillsEnvHealthReport_NoTrackedSkills(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Workspace = t.TempDir()

	report := buildSkillsEnvHealthReport(
		cfg,
		nil,
		func(file string) (string, error) { return "", errors.New("missing") },
		func(key string) (string, bool) { return "", false },
	)

	require.Contains(t, report, "Env files loaded: none")
	require.Contains(t, report, "Tracked skills found: none in workspace/skills")
	require.False(t, strings.Contains(report, "Required env vars:"))
}

func TestBuildSkillsEnvHealthReport_GitRemoteRefsRequireGh(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(workspace, "skills", "git-summary"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "skills", "git-summary", "SKILL.md"), []byte("# test"), 0o644))

	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Workspace = workspace

	report := buildSkillsEnvHealthReport(
		cfg,
		nil,
		func(file string) (string, error) {
			switch file {
			case "git", "curl", "jq", "python3":
				return "/usr/bin/" + file, nil
			default:
				return "", errors.New("missing")
			}
		},
		func(key string) (string, bool) {
			if key == "GIT_REPOS" {
				return "thisisprabha/time-left,https://github.com/thisisprabha/networth", true
			}
			return "", false
		},
	)

	require.Contains(t, report, "gh: missing")
}
