---
name: slack-digest
description: Summarize recent Slack channel messages using the Slack API.
metadata: {"nanobot":{"emoji":"💬","requires":{"bins":["curl","jq"]}}}
---

# Slack Digest

Fetch recent messages from a Slack channel and generate a summary.
Can be triggered manually ("summarize my slack") or via heartbeat.

## Required Environment Variables

- `SLACK_BOT_TOKEN` — Slack Bot User OAuth Token (`xoxb-...`)
- `SLACK_CHANNEL_ID` — Channel ID to monitor (e.g., `C0XXXXXXX`)

### How to get these:
1. Go to https://api.slack.com/apps → Create New App
2. Add Bot Token Scopes: `channels:history`, `channels:read`
3. Install to workspace → Copy Bot User OAuth Token
4. Get channel ID: right-click channel name → "View channel details" → copy ID

## Steps

### 1. Fetch recent messages (last 24h)

```bash
OLDEST=$(python3 -c "import time; print(int(time.time()) - 86400)")
curl -s "https://slack.com/api/conversations.history?channel=$SLACK_CHANNEL_ID&oldest=$OLDEST&limit=50" \
  -H "Authorization: Bearer $SLACK_BOT_TOKEN" \
  | jq -r '.messages[] | "\(.user // "bot"): \(.text)"' 2>/dev/null
```

### 2. Summarize

Take the raw messages and produce:

```
💬 Slack Digest (last 24h)

Key discussions:
• <topic 1 — brief summary>
• <topic 2 — brief summary>

Action items mentioned:
• <any tasks/decisions>

Mentions of you:
• <any @mentions or direct references>

Total: X messages
```

### 3. Error handling

If `SLACK_BOT_TOKEN` is not set: "Slack not configured. Set SLACK_BOT_TOKEN and SLACK_CHANNEL_ID."
If API returns error: Report the error message.
If no messages: "No Slack activity in the last 24 hours."
