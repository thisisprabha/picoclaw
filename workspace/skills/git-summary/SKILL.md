---
name: git-summary
description: Weekly git activity summary across local repositories.
metadata: {"nanobot":{"emoji":"📊","requires":{"bins":["git"]}}}
---

# Git Summary

Generate a weekly activity report across configured git repositories.
Triggered weekly via heartbeat (Monday mornings) or manually.

## Required Environment Variables

- `GIT_REPOS` — Comma-separated list of repo paths (e.g., `/home/yoyoboy/repos/project1,/home/yoyoboy/repos/project2`)

## Steps

### 1. Collect git logs

For each repo in `GIT_REPOS`:

```bash
if [ -z "${GIT_REPOS:-}" ]; then
  echo "GIT_REPOS is not set. Example: export GIT_REPOS=/home/user/repo1,/home/user/repo2"
  exit 0
fi

printf '%s\n' "$GIT_REPOS" | tr ',' '\n' | while IFS= read -r repo; do
  repo=$(echo "$repo" | xargs) # trim whitespace
  [ -z "$repo" ] && continue
  if [ -d "$repo/.git" ]; then
    echo "=== $(basename "$repo") ==="
    cd "$repo"
    # Show last week commits from any author since this is a personal machine
    git log --since="1 week ago" --oneline 2>/dev/null || echo "(no commits found in last week)"
    echo ""
  else
    echo "=== Skip: $repo (not a git repo) ==="
  fi
done
```

### 2. Summarize

Compile into a formatted report:

```
📊 Weekly Git Summary

## <repo-name-1>
- X commits this week
- Key changes: <brief summary of commit messages>

## <repo-name-2>
- Y commits this week
- Key changes: <brief summary>

Total: Z commits across N repos
```

### 3. Send via message tool

If there were any commits, deliver the summary to the user.
If no commits across all repos, send a brief "No git activity this week."
