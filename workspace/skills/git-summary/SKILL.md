---
name: git-summary
description: Weekly git activity summary across local repositories and/or GitHub repos.
metadata: {"nanobot":{"emoji":"📊","requires":{"bins":["git","gh","jq"]}}}
---

# Git Summary

Generate a weekly activity report across configured repositories.
Triggered weekly via heartbeat (Monday mornings) or manually.

## Required Environment Variables

- `GIT_REPOS` — Comma-separated entries.
  - Local path mode: `/home/yoyoboy/repos/project1`
  - GitHub mode (no local clone required):
    - `owner/repo` (e.g., `thisisprabha/time-left`)
    - `https://github.com/owner/repo`
    - `git@github.com:owner/repo.git`

## Steps

### 1. Collect git logs

For each entry in `GIT_REPOS`:

```bash
if [ -z "${GIT_REPOS:-}" ]; then
  echo "GIT_REPOS is not set. Example: export GIT_REPOS=/home/user/repo1,/home/user/repo2"
  exit 0
fi

SINCE_ISO=$(date -u -d '7 days ago' +%Y-%m-%dT%H:%M:%SZ 2>/dev/null)
[ -z "$SINCE_ISO" ] && SINCE_ISO=$(python3 - <<'PY'
from datetime import datetime, timedelta, timezone
print((datetime.now(timezone.utc) - timedelta(days=7)).strftime("%Y-%m-%dT%H:%M:%SZ"))
PY
)

# IMPORTANT: do NOT export/overwrite GIT_REPOS in this command.
# Use the runtime env value as-is.
TOTAL_COMMITS=0
TMP_REPOS=$(mktemp)
printf '%s\n' "$GIT_REPOS" | tr ',' '\n' > "$TMP_REPOS"
while IFS= read -r repo; do
  repo=$(echo "$repo" | xargs) # trim whitespace
  [ -z "$repo" ] && continue

  if [ -d "$repo/.git" ]; then
    echo "=== $(basename "$repo") ==="
    # Local repo mode
    COUNT=$(cd "$repo" && git rev-list --count --since="1 week ago" HEAD 2>/dev/null || echo 0)
    echo "commits: $COUNT"
    TOTAL_COMMITS=$((TOTAL_COMMITS + COUNT))
    if [ "$COUNT" -gt 0 ]; then
      (cd "$repo" && git log --since="1 week ago" --oneline -n 5 2>/dev/null)
    else
      echo "(no commits found in last week)"
    fi
    echo ""
    continue
  fi

  # GitHub repo mode (owner/repo), useful when work happens on another machine
  if printf '%s' "$repo" | grep -Eq '^(https://github\.com/|git@github\.com:)?[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+(\.git)?$'; then
    repo_norm=$(printf '%s' "$repo" | sed -E 's#^https://github\.com/##; s#^git@github\.com:##; s#\.git$##')
    echo "=== $repo_norm ==="

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

    COUNT=$(gh api "repos/$repo_norm/commits?since=$SINCE_ISO&per_page=100" --jq 'length' 2>/dev/null)
    case "$COUNT" in
      ''|*[!0-9]*)
      echo "(failed to query GitHub API)"
      echo ""
      continue
      ;;
    esac

    if [ "$COUNT" = "0" ]; then
      echo "commits: 0"
      echo "(no commits found in last week)"
      echo ""
      continue
    fi

    if [ "$COUNT" -lt 0 ]; then
      echo "(failed to query GitHub API)"
      echo ""
      continue
    fi

    echo "commits: $COUNT"
    TOTAL_COMMITS=$((TOTAL_COMMITS + COUNT))
    TOP5=$(gh api "repos/$repo_norm/commits?since=$SINCE_ISO&per_page=5" --jq '.[] | "\(.sha[0:7]) \(.commit.message | split("\n")[0])"' 2>/dev/null)
    if [ -n "$TOP5" ]; then
      printf '%s\n' "$TOP5"
    else
      echo "(failed to query GitHub API)"
    fi
    echo ""
    continue
  fi

  echo "=== Skip: $repo (neither local git path nor owner/repo) ==="
done < "$TMP_REPOS"
rm -f "$TMP_REPOS"
echo "TOTAL_COMMITS=$TOTAL_COMMITS"
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

Rules:
- Parse every `=== <repo> ===` block from command output.
- Do not drop repositories in summary.
- Keep summary short: total commits + top highlights.

### 3. Send via message tool

If there were any commits, deliver the summary to the user.
If no commits across all repos, send a brief "No git activity this week."
