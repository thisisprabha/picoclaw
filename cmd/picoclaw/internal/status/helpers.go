package status

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sipeed/picoclaw/cmd/picoclaw/internal"
	"github.com/sipeed/picoclaw/pkg/auth"
	"github.com/sipeed/picoclaw/pkg/config"
)

var (
	trackedSkillNames = []string{
		"email-digest",
		"todoist-manager",
		"git-summary",
	}
	skillEnvRequirements = map[string][]string{
		"email-digest":    {"EMAIL_IMAP_HOST", "EMAIL_ADDRESS", "EMAIL_PASSWORD"},
		"todoist-manager": {"TODOIST_API_TOKEN"},
		"git-summary":     {"GIT_REPOS"},
	}
	baseRequiredBins = []string{"git", "curl", "jq", "python3"}
	remoteRepoRefRE  = regexp.MustCompile(`^(https://github\.com/|git@github\.com:)?[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+(\.git)?$`)
)

func statusCmd() {
	cfg, err := internal.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	configPath := internal.GetConfigPath()

	fmt.Printf("%s picoclaw Status\n", internal.Logo)
	fmt.Printf("Version: %s\n", internal.FormatVersion())
	build, _ := internal.FormatBuildInfo()
	if build != "" {
		fmt.Printf("Build: %s\n", build)
	}
	fmt.Println()

	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("Config:", configPath, "✓")
	} else {
		fmt.Println("Config:", configPath, "✗")
	}

	workspace := cfg.WorkspacePath()
	if _, err := os.Stat(workspace); err == nil {
		fmt.Println("Workspace:", workspace, "✓")
	} else {
		fmt.Println("Workspace:", workspace, "✗")
	}

	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Model: %s\n", cfg.Agents.Defaults.GetModelName())

		hasOpenRouter := cfg.Providers.OpenRouter.APIKey != ""
		hasAnthropic := cfg.Providers.Anthropic.APIKey != ""
		hasOpenAI := cfg.Providers.OpenAI.APIKey != ""
		hasGemini := cfg.Providers.Gemini.APIKey != ""
		hasZhipu := cfg.Providers.Zhipu.APIKey != ""
		hasQwen := cfg.Providers.Qwen.APIKey != ""
		hasGroq := cfg.Providers.Groq.APIKey != ""
		hasVLLM := cfg.Providers.VLLM.APIBase != ""
		hasMoonshot := cfg.Providers.Moonshot.APIKey != ""
		hasDeepSeek := cfg.Providers.DeepSeek.APIKey != ""
		hasVolcEngine := cfg.Providers.VolcEngine.APIKey != ""
		hasNvidia := cfg.Providers.Nvidia.APIKey != ""
		hasOllama := cfg.Providers.Ollama.APIBase != ""

		status := func(enabled bool) string {
			if enabled {
				return "✓"
			}
			return "not set"
		}
		fmt.Println("OpenRouter API:", status(hasOpenRouter))
		fmt.Println("Anthropic API:", status(hasAnthropic))
		fmt.Println("OpenAI API:", status(hasOpenAI))
		fmt.Println("Gemini API:", status(hasGemini))
		fmt.Println("Zhipu API:", status(hasZhipu))
		fmt.Println("Qwen API:", status(hasQwen))
		fmt.Println("Groq API:", status(hasGroq))
		fmt.Println("Moonshot API:", status(hasMoonshot))
		fmt.Println("DeepSeek API:", status(hasDeepSeek))
		fmt.Println("VolcEngine API:", status(hasVolcEngine))
		fmt.Println("Nvidia API:", status(hasNvidia))
		if hasVLLM {
			fmt.Printf("vLLM/Local: ✓ %s\n", cfg.Providers.VLLM.APIBase)
		} else {
			fmt.Println("vLLM/Local: not set")
		}
		if hasOllama {
			fmt.Printf("Ollama: ✓ %s\n", cfg.Providers.Ollama.APIBase)
		} else {
			fmt.Println("Ollama: not set")
		}
	}

	fmt.Println(buildSkillsEnvHealthReport(cfg, internal.GetLoadedEnvFiles(), exec.LookPath, os.LookupEnv))

	store, _ := auth.LoadStore()
	if store != nil && len(store.Credentials) > 0 {
		fmt.Println("\nOAuth/Token Auth:")
		for provider, cred := range store.Credentials {
			status := "authenticated"
			if cred.IsExpired() {
				status = "expired"
			} else if cred.NeedsRefresh() {
				status = "needs refresh"
			}
			fmt.Printf("  %s (%s): %s\n", provider, cred.AuthMethod, status)
		}
	}
}

func buildSkillsEnvHealthReport(
	cfg *config.Config,
	loadedEnvFiles []string,
	lookupPath func(file string) (string, error),
	lookupEnv func(key string) (string, bool),
) string {
	var b strings.Builder
	b.WriteString("\nSkills Env Health:\n")

	if len(loadedEnvFiles) == 0 {
		b.WriteString("Env files loaded: none\n")
	} else {
		b.WriteString("Env files loaded:\n")
		for _, path := range loadedEnvFiles {
			fmt.Fprintf(&b, "  - %s\n", path)
		}
	}

	installedSkills := detectInstalledTrackedSkills(cfg.WorkspacePath())
	if len(installedSkills) == 0 {
		b.WriteString("Tracked skills found: none in workspace/skills\n")
	} else {
		b.WriteString("Tracked skills found:\n")
		for _, name := range trackedSkillNames {
			if installedSkills[name] {
				fmt.Fprintf(&b, "  - %s\n", name)
			}
		}
	}

	requiredVars := requiredVarsForInstalledSkills(installedSkills)
	gitRepos, _ := lookupEnv("GIT_REPOS")
	if len(requiredVars) > 0 {
		b.WriteString("Required env vars:\n")
		for _, key := range requiredVars {
			val, ok := lookupEnv(key)
			ok = ok && strings.TrimSpace(val) != ""
			fmt.Fprintf(&b, "  - %s: %s\n", key, statusMarker(ok))
		}
	}

	b.WriteString("Required binaries:\n")
	for _, bin := range requiredBinsForInstalledSkills(installedSkills, gitRepos) {
		_, err := lookupPath(bin)
		fmt.Fprintf(&b, "  - %s: %s\n", bin, statusMarker(err == nil))
	}

	if cfg.Agents.Defaults.RestrictToWorkspace {
		outside := internal.FindGitReposOutsideWorkspace(cfg.WorkspacePath(), gitRepos)
		if len(outside) > 0 {
			fmt.Fprintf(&b, "Warning: restrict_to_workspace=true may block %d GIT_REPOS path(s) outside workspace:\n", len(outside))
			for _, repo := range outside {
				fmt.Fprintf(&b, "  - %s\n", repo)
			}
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

func detectInstalledTrackedSkills(workspace string) map[string]bool {
	installed := make(map[string]bool, len(trackedSkillNames))
	for _, name := range trackedSkillNames {
		skillPath := filepath.Join(workspace, "skills", name, "SKILL.md")
		if _, err := os.Stat(skillPath); err == nil {
			installed[name] = true
		}
	}
	return installed
}

func requiredVarsForInstalledSkills(installed map[string]bool) []string {
	seen := make(map[string]struct{})
	var keys []string
	for _, skill := range trackedSkillNames {
		if !installed[skill] {
			continue
		}
		for _, key := range skillEnvRequirements[skill] {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			keys = append(keys, key)
		}
	}
	return keys
}

func statusMarker(ok bool) string {
	if ok {
		return "✓"
	}
	return "missing"
}

func requiredBinsForInstalledSkills(installed map[string]bool, gitReposCSV string) []string {
	seen := make(map[string]struct{}, len(baseRequiredBins)+1)
	out := make([]string, 0, len(baseRequiredBins)+1)
	for _, bin := range baseRequiredBins {
		if _, ok := seen[bin]; ok {
			continue
		}
		seen[bin] = struct{}{}
		out = append(out, bin)
	}

	if installed["git-summary"] && hasRemoteRepoRefs(gitReposCSV) {
		if _, ok := seen["gh"]; !ok {
			seen["gh"] = struct{}{}
			out = append(out, "gh")
		}
	}

	return out
}

func hasRemoteRepoRefs(gitReposCSV string) bool {
	for _, entry := range strings.Split(gitReposCSV, ",") {
		repo := strings.TrimSpace(entry)
		if repo == "" {
			continue
		}
		if remoteRepoRefRE.MatchString(repo) {
			return true
		}
	}
	return false
}
