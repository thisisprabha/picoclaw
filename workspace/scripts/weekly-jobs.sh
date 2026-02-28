#!/bin/bash
# Weekly Jobs — triggered by cron every Monday at 9:00 AM IST
# Runs email digest and git summary via picoclaw agent one-shot
#
# Install: crontab -e → 0 9 * * 1 ~/.picoclaw/workspace/scripts/weekly-jobs.sh >> /tmp/picoclaw-weekly.log 2>&1

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Weekly jobs started"

PICOCLAW_BIN="${PICOCLAW_BIN:-picoclaw}"
ENV_FILE="${PICOCLAW_ENV_FILE:-$HOME/.picoclaw/.env.picoclaw}"

# Load environment
if [ -f "$ENV_FILE" ]; then
  set -a
  source "$ENV_FILE"
  set +a
fi

if ! command -v "$PICOCLAW_BIN" >/dev/null 2>&1; then
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: picoclaw binary not found: $PICOCLAW_BIN"
  exit 1
fi

# Run email digest
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Running email digest..."
"$PICOCLAW_BIN" agent -s "cron:weekly-email:$(date +%G-W%V)" -m "Run the email-digest skill and send a short actionable summary for this week. No long narrative." 2>&1

# Run git summary
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Running git summary..."
"$PICOCLAW_BIN" agent -s "cron:weekly-git:$(date +%G-W%V)" -m "Run the git-summary skill and send a concise weekly repo summary using all repos from GIT_REPOS." 2>&1

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Weekly jobs complete"
