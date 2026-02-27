---
name: git-summary
description: Weekly git activity summary across local repositories and/or GitHub repos.
metadata: {"nanobot":{"emoji":"📊","requires":{"bins":["git","gh","jq","python3"]}}}
---

# Git Summary

Generate a weekly activity report across configured repositories.
Triggered weekly via heartbeat (Monday mornings) or manually.

## Required Environment Variables

- `GIT_REPOS` — Comma-separated entries.
  - Local path mode: `/home/yoyoboy/repos/project1`
  - GitHub mode (no local clone required): `owner/repo` (e.g., `thisisprabha/time-left`)

## Steps

### 1. Collect git logs

For each entry in `GIT_REPOS`:

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
    # Local repo mode
    git log --since="1 week ago" --oneline 2>/dev/null || echo "(no commits found in last week)"
    echo ""
    continue
  fi

  # GitHub repo mode (owner/repo), useful when work happens on another machine
  if printf '%s' "$repo" | grep -Eq '^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$'; then
    echo "=== $repo ==="

    if ! command -v gh >/dev/null 2>&1; then
      echo "(skip: gh CLI not installed)"
      echo ""
      continue
    fi

    if ! gh auth status >/dev/null 2>&1; then
      echo "(skip: gh not authenticated. Run: gh auth login -h github.com -p ssh -w)"
      echo ""
      continue
    fi

    since_iso=$(python3 - <<'PY'
from datetime import datetime, timedelta, timezone
print((datetime.now(timezone.utc) - timedelta(days=7)).strftime("%Y-%m-%dT%H:%M:%SZ"))
PY
)

    gh api -X GET "repos/$repo/commits?since=$since_iso&per_page=100" 2>/dev/null \
      | jq -r 'if type=="array" then (if length==0 then "(no commits found in last week)" else .[] | "\(.sha[0:7]) \(.commit.message | split("\n")[0])" end) else "(failed to query GitHub API)" end'
    echo ""
    continue
  fi

  echo "=== Skip: $repo (neither local git path nor owner/repo) ==="
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
