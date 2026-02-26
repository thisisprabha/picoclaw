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

### List today's tasks

```bash
curl -s "https://api.todoist.com/rest/v2/tasks?filter=today" \
  -H "Authorization: Bearer $TODOIST_API_TOKEN" \
  | jq -r '.[] | "☐ [\(.id)] \(.content) (p\(.priority))"'
```

### Create a task

```bash
curl -s -X POST "https://api.todoist.com/rest/v2/tasks" \
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
curl -s -X POST "https://api.todoist.com/rest/v2/tasks/<TASK_ID>/close" \
  -H "Authorization: Bearer $TODOIST_API_TOKEN"
```

### Response Format

After each operation, confirm to the user:

- **Create**: "✅ Task added: <content> (due: <date>)"
- **List**: Show numbered list with priorities
- **Complete**: "✅ Done: <content>"
- **Error**: "❌ Could not <action>. Is TODOIST_API_TOKEN set?"
