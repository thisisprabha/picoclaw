#!/bin/bash
# Daily Briefing — run once per day via cron (recommended: 8:15 AM IST)
#
# Install:
#   crontab -e
#   15 8 * * * ~/.picoclaw/workspace/scripts/daily-briefing.sh >> /tmp/picoclaw-daily.log 2>&1

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Daily briefing started"

PICOCLAW_BIN="${PICOCLAW_BIN:-picoclaw}"
ENV_FILE="${PICOCLAW_ENV_FILE:-$HOME/.picoclaw/.env.picoclaw}"

# Load environment
if [ -f "$ENV_FILE" ]; then
  set -a
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a
fi

if ! command -v "$PICOCLAW_BIN" >/dev/null 2>&1; then
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: picoclaw binary not found: $PICOCLAW_BIN"
  exit 1
fi

SESSION_KEY="cron:daily-briefing:$(date +%F)"

"$PICOCLAW_BIN" agent -s "$SESSION_KEY" -m "Run the morning-briefing skill now. Keep it short and productivity-only: today's tasks, calendar commitments, quick weather line. No news section." 2>&1

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Daily briefing complete"
