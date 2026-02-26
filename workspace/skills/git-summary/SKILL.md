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
IFS=',' read -ra REPOS <<< "$GIT_REPOS"
for repo in "${REPOS[@]}"; do
  if [ -d "$repo/.git" ]; then
    echo "=== $(basename $repo) ==="
    cd "$repo"
    git log --since="1 week ago" --oneline --author="$(git config user.email)" 2>/dev/null || echo "(no commits)"
    echo ""
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
