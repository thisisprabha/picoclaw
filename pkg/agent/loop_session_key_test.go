package agent

import "testing"

func TestResolveSessionKey_UsesRouteWhenNoMessageSession(t *testing.T) {
	got := resolveSessionKey("agent:main:main", "", "cli", "main")
	if got != "agent:main:main" {
		t.Fatalf("expected routed key, got %q", got)
	}
}

func TestResolveSessionKey_HonorsAgentScopedSession(t *testing.T) {
	got := resolveSessionKey("agent:main:main", "agent:main:custom", "cli", "main")
	if got != "agent:main:custom" {
		t.Fatalf("expected explicit agent-scoped session, got %q", got)
	}
}

func TestResolveSessionKey_NamespacesCliSession(t *testing.T) {
	got := resolveSessionKey("agent:main:main", "cli:git-remote-check", "cli", "main")
	if got != "agent:main:cli:git-remote-check" {
		t.Fatalf("expected namespaced cli session, got %q", got)
	}
}

func TestResolveSessionKey_NonCliDoesNotOverride(t *testing.T) {
	got := resolveSessionKey("agent:main:main", "telegram:abc", "telegram", "main")
	if got != "agent:main:main" {
		t.Fatalf("expected routed key for non-cli channel, got %q", got)
	}
}

