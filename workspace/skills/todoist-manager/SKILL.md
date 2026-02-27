---
name: todoist-manager
description: Create, list, and complete Todoist tasks via REST API.
metadata: {"nanobot":{"emoji":"✅","requires":{"bins":["curl","jq"]}}}
---

# Todoist Manager

Manage Todoist tasks from Telegram. Supports creating, listing, and completing tasks.

## Required Environment Variables

- `TODOIST_API_TOKEN` — Your Todoist API token (Settings > Integrations > Developer)

## Commands

The user can trigger this skill with messages like:
- "add task: buy groceries"
- "add task: review PR by Friday"
- "my tasks" / "list tasks" / "today's tasks"
- "complete task: buy groceries"

## Operations

### Token precheck

Run this before any Todoist API call:

```bash
if [ -z "${TODOIST_API_TOKEN:-}" ]; then
  echo "TODOIST_API_TOKEN is not set. Configure it in ~/.picoclaw/.env.picoclaw"
  exit 0
fi
```

### List today's tasks

Use the command exactly with double-quoted auth header so `$TODOIST_API_TOKEN` expands:

```bash
curl -s "https://api.todoist.com/api/v1/tasks?filter=today" \
  -H "Authorization: Bearer $TODOIST_API_TOKEN" \
  | jq -r 'if type=="object" and has("results") then .results[] else .[] end | "☐ [\(.id)] \(.content) (p\(.priority // 1))"'
```

### Create a task

```bash
curl -s -X POST "https://api.todoist.com/api/v1/tasks" \
  -H "Authorization: Bearer $TODOIST_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content": "<TASK_CONTENT>", "due_string": "<DUE_DATE_OR_today>"}'
```

Parse the user's message to extract:
- `content` — The task text
- `due_string` — Any date mentioned (e.g., "tomorrow", "Friday", "next week"). Default: "today"

### Complete a task

First find the task ID by listing tasks and matching by content, then:

```bash
curl -s -X POST "https://api.todoist.com/api/v1/tasks/<TASK_ID>/close" \
  -H "Authorization: Bearer $TODOIST_API_TOKEN"
```

### Response Format

After each operation, confirm to the user:

- **Create**: "✅ Task added: <content> (due: <date>)"
- **List**: Show numbered list with priorities
- **Complete**: "✅ Done: <content>"
- **Error**: "❌ Could not <action>. Check TODOIST_API_TOKEN and API response."
