package internal

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigPath(t *testing.T) {
	t.Setenv("HOME", "/tmp/home")

	got := GetConfigPath()
	want := filepath.Join("/tmp/home", ".picoclaw", "config.json")

	assert.Equal(t, want, got)
}

func TestFormatVersion_NoGitCommit(t *testing.T) {
	oldVersion, oldGit := version, gitCommit
	t.Cleanup(func() { version, gitCommit = oldVersion, oldGit })

	version = "1.2.3"
	gitCommit = ""

	assert.Equal(t, "1.2.3", FormatVersion())
}

func TestFormatVersion_WithGitCommit(t *testing.T) {
	oldVersion, oldGit := version, gitCommit
	t.Cleanup(func() { version, gitCommit = oldVersion, oldGit })

	version = "1.2.3"
	gitCommit = "abc123"

	assert.Equal(t, "1.2.3 (git: abc123)", FormatVersion())
}

func TestFormatBuildInfo_UsesBuildTimeAndGoVersion_WhenSet(t *testing.T) {
	oldBuildTime, oldGoVersion := buildTime, goVersion
	t.Cleanup(func() { buildTime, goVersion = oldBuildTime, oldGoVersion })

	buildTime = "2026-02-20T00:00:00Z"
	goVersion = "go1.23.0"

	build, goVer := FormatBuildInfo()

	assert.Equal(t, buildTime, build)
	assert.Equal(t, goVersion, goVer)
}

func TestFormatBuildInfo_EmptyBuildTime_ReturnsEmptyBuild(t *testing.T) {
	oldBuildTime, oldGoVersion := buildTime, goVersion
	t.Cleanup(func() { buildTime, goVersion = oldBuildTime, oldGoVersion })

	buildTime = ""
	goVersion = "go1.23.0"

	build, goVer := FormatBuildInfo()

	assert.Empty(t, build)
	assert.Equal(t, goVersion, goVer)
}

func TestFormatBuildInfo_EmptyGoVersion_FallsBackToRuntimeVersion(t *testing.T) {
	oldBuildTime, oldGoVersion := buildTime, goVersion
	t.Cleanup(func() { buildTime, goVersion = oldBuildTime, oldGoVersion })

	buildTime = "x"
	goVersion = ""

	build, goVer := FormatBuildInfo()

	assert.Equal(t, "x", build)
	assert.Equal(t, runtime.Version(), goVer)
}

func TestGetConfigPath_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-specific HOME behavior varies; run on windows")
	}

	testUserProfilePath := `C:\Users\Test`
	t.Setenv("USERPROFILE", testUserProfilePath)

	got := GetConfigPath()
	want := filepath.Join(testUserProfilePath, ".picoclaw", "config.json")

	require.True(t, strings.EqualFold(got, want), "GetConfigPath() = %q, want %q", got, want)
}

func TestGetVersion(t *testing.T) {
	assert.Equal(t, "dev", GetVersion())
}

func TestLoadConfig_EnvLoadOrderAndNoOverwrite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PICOCLAW_ENV_FILE", "")

	homeDir := filepath.Join(home, ".picoclaw")
	workspaceDir := filepath.Join(homeDir, "workspace")
	require.NoError(t, os.MkdirAll(workspaceDir, 0o755))

	homeEnvPicoclaw := filepath.Join(homeDir, ".env.picoclaw")
	homeEnv := filepath.Join(homeDir, ".env")
	workspaceEnvPicoclaw := filepath.Join(workspaceDir, ".env.picoclaw")
	workspaceEnv := filepath.Join(workspaceDir, ".env")

	require.NoError(t, os.WriteFile(homeEnvPicoclaw, []byte("TEST_PICOCLAW_ENV_ORDER=home-picoclaw\n"), 0o600))
	require.NoError(t, os.WriteFile(homeEnv, []byte("TEST_PICOCLAW_ENV_ORDER=home\n"), 0o600))
	require.NoError(t, os.WriteFile(workspaceEnvPicoclaw, []byte("TEST_PICOCLAW_ENV_ORDER=workspace-picoclaw\n"), 0o600))
	require.NoError(t, os.WriteFile(workspaceEnv, []byte("TEST_PICOCLAW_ENV_ORDER=workspace\n"), 0o600))

	wd := t.TempDir()
	cwdEnv := filepath.Join(wd, ".env")
	require.NoError(t, os.WriteFile(cwdEnv, []byte("TEST_PICOCLAW_ENV_ORDER=cwd\n"), 0o600))

	prevWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(wd))
	t.Cleanup(func() { _ = os.Chdir(prevWD) })

	_, err = LoadConfig()
	require.NoError(t, err)
	require.Equal(t, "home-picoclaw", os.Getenv("TEST_PICOCLAW_ENV_ORDER"))

	loaded := GetLoadedEnvFiles()
	expected := []string{homeEnvPicoclaw, homeEnv, cwdEnv, workspaceEnvPicoclaw, workspaceEnv}
	for i := range expected {
		if resolved, resolveErr := filepath.EvalSymlinks(expected[i]); resolveErr == nil {
			expected[i] = resolved
		}
	}
	for i := range loaded {
		if resolved, resolveErr := filepath.EvalSymlinks(loaded[i]); resolveErr == nil {
			loaded[i] = resolved
		}
	}
	require.Equal(t, expected, loaded)
}

func TestLoadConfig_CustomEnvFileHasHighestPriority(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	homeDir := filepath.Join(home, ".picoclaw")
	require.NoError(t, os.MkdirAll(homeDir, 0o755))

	customEnv := filepath.Join(t.TempDir(), ".env.custom")
	require.NoError(t, os.WriteFile(customEnv, []byte("TEST_PICOCLAW_ENV_CUSTOM=from-custom\n"), 0o600))
	t.Setenv("PICOCLAW_ENV_FILE", customEnv)

	homeEnv := filepath.Join(homeDir, ".env.picoclaw")
	require.NoError(t, os.WriteFile(homeEnv, []byte("TEST_PICOCLAW_ENV_CUSTOM=from-home\n"), 0o600))

	_, err := LoadConfig()
	require.NoError(t, err)
	require.Equal(t, "from-custom", os.Getenv("TEST_PICOCLAW_ENV_CUSTOM"))
}

func TestLoadConfig_DoesNotOverwriteExistingEnvironment(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("TEST_PICOCLAW_ENV_NO_OVERWRITE", "from-process")

	homeDir := filepath.Join(home, ".picoclaw")
	require.NoError(t, os.MkdirAll(homeDir, 0o755))

	homeEnv := filepath.Join(homeDir, ".env.picoclaw")
	require.NoError(t, os.WriteFile(homeEnv, []byte("TEST_PICOCLAW_ENV_NO_OVERWRITE=from-file\n"), 0o600))

	_, err := LoadConfig()
	require.NoError(t, err)
	require.Equal(t, "from-process", os.Getenv("TEST_PICOCLAW_ENV_NO_OVERWRITE"))
}

func TestFindGitReposOutsideWorkspace(t *testing.T) {
	workspace := t.TempDir()
	inside := filepath.Join(workspace, "repo-inside")
	outside := filepath.Join(t.TempDir(), "repo-outside")

	got := FindGitReposOutsideWorkspace(workspace, inside+","+outside+","+inside)
	require.Equal(t, []string{outside}, got)
}
