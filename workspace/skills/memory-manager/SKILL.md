---
name: memory-manager
description: Maintain long-term memory in MEMORY.md with search, add, and prune operations.
metadata: {"nanobot":{"emoji":"🧠","requires":{"bins":["grep"]}}}
---

# Memory Manager

Maintain `memory/MEMORY.md` as a lightweight long-term memory store.
This serves as a simple RAG system — no vector DB needed on Pi3.

## Operations

### 1. Add a memory entry

When the user shares important information or you learn something new, append to MEMORY.md:

```
### [YYYY-MM-DD] [tag] Entry text
```

Tags: `[permanent]`, `[preference]`, `[task]`, `[note]`, `[decision]`, `[contact]`

Example:
```
### 2026-02-26 [preference] User prefers dark mode in all apps
### 2026-02-26 [task] Need to submit App Store review by March 1
### 2026-02-26 [permanent] User's Pi3 IP is 192.168.1.23
```

Use `append_file` tool to add entries to `memory/MEMORY.md`.

### 2. Search memory (lightweight RAG)

When the user asks a question that might be answered by memory:

```bash
grep -i "<search_term>" memory/MEMORY.md
```

Or for multi-word search:

```bash
grep -iE "(term1|term2|term3)" memory/MEMORY.md
```

Use the results to augment your response with recalled context.

### 3. Prune old entries

During heartbeat, remove entries older than 30 days UNLESS tagged `[permanent]`:

```bash
# List entries older than 30 days that are NOT permanent
CUTOFF=$(date -d "-30 days" +%Y-%m-%d 2>/dev/null || date -v-30d +%Y-%m-%d)
grep "^### " memory/MEMORY.md | grep -v "\[permanent\]" | while read line; do
  DATE=$(echo "$line" | grep -oP '\d{4}-\d{2}-\d{2}')
  if [[ "$DATE" < "$CUTOFF" ]]; then
    echo "PRUNE: $line"
  fi
done
```

After identifying entries to prune, rewrite MEMORY.md without them.
Keep total entries under 200.

### 4. Memory structure

The MEMORY.md file should maintain these sections:

```markdown
# Long-term Memory

## User Information
(Facts about the user — tagged [permanent])

## Preferences
(Learned preferences — tagged [preference])

## Active Tasks
(Current tasks and deadlines — tagged [task])

## Notes
(General notes — tagged [note])

## Decisions
(Important decisions made — tagged [decision])

## Contacts
(People and their info — tagged [contact])
```

### 5. When to update memory

- User explicitly says "remember this" or "note that"
- You learn a new preference from user behavior
- A task deadline is mentioned
- Important decisions are made
- Contact information is shared
- After completing significant work (log what was done)
