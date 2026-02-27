package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvFile_ParsesAndPreservesExistingVars(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := `
# comment
export TEST_ENV_A=from-file
TEST_ENV_B = "hello world"
TEST_ENV_C='single quoted'
TEST_ENV_D=
INVALID LINE
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	t.Setenv("TEST_ENV_A", "from-process")

	loaded, err := LoadEnvFile(path, false)
	if err != nil {
		t.Fatalf("LoadEnvFile() error: %v", err)
	}
	if !loaded {
		t.Fatal("expected loaded=true for existing env file")
	}

	if got := os.Getenv("TEST_ENV_A"); got != "from-process" {
		t.Fatalf("TEST_ENV_A = %q, want %q", got, "from-process")
	}
	if got := os.Getenv("TEST_ENV_B"); got != "hello world" {
		t.Fatalf("TEST_ENV_B = %q, want %q", got, "hello world")
	}
	if got := os.Getenv("TEST_ENV_C"); got != "single quoted" {
		t.Fatalf("TEST_ENV_C = %q, want %q", got, "single quoted")
	}
	if got := os.Getenv("TEST_ENV_D"); got != "" {
		t.Fatalf("TEST_ENV_D = %q, want empty string", got)
	}
}

func TestLoadEnvFile_OverwriteEnabled(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("TEST_ENV_OVERWRITE=new-value\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
	t.Setenv("TEST_ENV_OVERWRITE", "old-value")

	loaded, err := LoadEnvFile(path, true)
	if err != nil {
		t.Fatalf("LoadEnvFile() error: %v", err)
	}
	if !loaded {
		t.Fatal("expected loaded=true for existing env file")
	}
	if got := os.Getenv("TEST_ENV_OVERWRITE"); got != "new-value" {
		t.Fatalf("TEST_ENV_OVERWRITE = %q, want %q", got, "new-value")
	}
}

func TestLoadEnvFile_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.env")
	loaded, err := LoadEnvFile(path, false)
	if err != nil {
		t.Fatalf("LoadEnvFile() error: %v", err)
	}
	if loaded {
		t.Fatal("expected loaded=false for missing env file")
	}
}
