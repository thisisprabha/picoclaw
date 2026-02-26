---
name: quick-capture
description: Capture notes, MoM points, read-it-later links, and bookmarks from Telegram.
metadata: {"nanobot":{"emoji":"📝"}}
---

# Quick Capture

Capture and organize quick notes from Telegram messages into structured files.

## Trigger Phrases

- "note: ..." or "save note: ..." → General notes
- "mom: ..." or "meeting note: ..." → Minutes of meeting
- "read later: ..." or "bookmark: ..." → Read-it-later list
- "idea: ..." → Ideas capture

## File Locations

All files are in the workspace directory:

- `capture/notes.md` — General notes
- `capture/mom.md` — Minutes of meeting
- `capture/reading-list.md` — Read-it-later URLs and articles
- `capture/ideas.md` — Ideas and inspirations

## Operations

### 1. Capture a note

Parse the user's message to determine the type, then append to the appropriate file.

Format for each entry:
```
- [YYYY-MM-DD HH:MM] <content>
```

Use `append_file` tool. Create the file with a header if it doesn't exist:

```markdown
# Notes

Captured from Telegram via PicoClaw.

---

- [2026-02-26 14:30] Remember to update the privacy policy
```

### 2. List captures

When user asks "show my notes" or "what did I bookmark":

```bash
cat capture/notes.md 2>/dev/null || echo "No notes yet."
```

Show the last 10 entries from the requested file.

### 3. Clear old entries

When user asks to "clear notes" or "archive notes":
- Move current file to `capture/archive/<filename>-<date>.md`
- Start a fresh file

### Response Format

After capturing:
```
📝 Saved to <type>!
"<content preview...>"
```

After listing:
```
📋 Your recent <type> (last 10):
1. [date] content
2. [date] content
...
```
