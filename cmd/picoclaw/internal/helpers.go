package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/sipeed/picoclaw/pkg/config"
)

const Logo = "🦞"

var (
	version   = "dev"
	gitCommit string
	buildTime string
	goVersion string

	loadedEnvFilesMu sync.RWMutex
	loadedEnvFiles   []string
)

func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".picoclaw", "config.json")
}

func LoadConfig() (*config.Config, error) {
	configPath := GetConfigPath()
	loaded := make([]string, 0, 6)
	seen := make(map[string]struct{}, 6)

	if err := loadEnvCandidates(preConfigEnvPaths(configPath), &loaded, seen); err != nil {
		setLoadedEnvFiles(loaded)
		return nil, err
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		setLoadedEnvFiles(loaded)
		return nil, err
	}

	if err := loadEnvCandidates(postConfigEnvPaths(cfg.WorkspacePath()), &loaded, seen); err != nil {
		setLoadedEnvFiles(loaded)
		return nil, err
	}

	setLoadedEnvFiles(loaded)
	return cfg, nil
}

// GetLoadedEnvFiles returns the env files loaded by the latest LoadConfig call.
func GetLoadedEnvFiles() []string {
	loadedEnvFilesMu.RLock()
	defer loadedEnvFilesMu.RUnlock()
	out := make([]string, len(loadedEnvFiles))
	copy(out, loadedEnvFiles)
	return out
}

func setLoadedEnvFiles(files []string) {
	loadedEnvFilesMu.Lock()
	defer loadedEnvFilesMu.Unlock()
	loadedEnvFiles = append([]string(nil), files...)
}

func loadEnvCandidates(candidates []string, loaded *[]string, seen map[string]struct{}) error {
	for _, path := range candidates {
		normalized := normalizeEnvPath(path)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}

		ok, err := config.LoadEnvFile(normalized, false)
		if err != nil {
			return fmt.Errorf("loading env file %s: %w", normalized, err)
		}
		seen[normalized] = struct{}{}
		if ok {
			*loaded = append(*loaded, normalized)
		}
	}
	return nil
}

func preConfigEnvPaths(configPath string) []string {
	homeDir := filepath.Dir(configPath)
	paths := make([]string, 0, 4)
	if customPath := strings.TrimSpace(os.Getenv("PICOCLAW_ENV_FILE")); customPath != "" {
		paths = append(paths, customPath)
	}
	paths = append(paths,
		filepath.Join(homeDir, ".env.picoclaw"),
		filepath.Join(homeDir, ".env"),
	)
	if wd, err := os.Getwd(); err == nil && wd != "" {
		paths = append(paths, filepath.Join(wd, ".env"))
	}
	return paths
}

func postConfigEnvPaths(workspace string) []string {
	if strings.TrimSpace(workspace) == "" {
		return nil
	}
	return []string{
		filepath.Join(workspace, ".env.picoclaw"),
		filepath.Join(workspace, ".env"),
	}
}

func normalizeEnvPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			if path == "~" {
				path = home
			} else if strings.HasPrefix(path, "~/") {
				path = filepath.Join(home, path[2:])
			}
		}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return abs
}

// FormatVersion returns the version string with optional git commit
func FormatVersion() string {
	v := version
	if gitCommit != "" {
		v += fmt.Sprintf(" (git: %s)", gitCommit)
	}
	return v
}

// FormatBuildInfo returns build time and go version info
func FormatBuildInfo() (string, string) {
	build := buildTime
	goVer := goVersion
	if goVer == "" {
		goVer = runtime.Version()
	}
	return build, goVer
}

// GetVersion returns the version string
func GetVersion() string {
	return version
}
