# Periodic Tasks

Checked every 3 hours. Keep responses minimal.
Only message the user for critical issues, due reminders, or the morning briefing.

## Quick Tasks (respond directly, no spawn)

- Check system health quickly. Alert ONLY when critical:
  - RAM > 90%
  - Disk > 95%
  - CPU temp > 80C
- If reminders are due, deliver them.
- Keep heartbeat response short and operational.
- If nothing requires attention, return exactly: `HEARTBEAT_OK`

## Scheduled Tasks (use spawn for async)

- **8:00 AM IST daily** — If current IST time is between 08:00–08:10, spawn the `morning-briefing` skill.
  - Spawn once per day only (check `memory/MEMORY.md` for `[heartbeat:morning-brief:YYYY-MM-DD]` to avoid duplicates).
  - After spawning, write `[heartbeat:morning-brief:YYYY-MM-DD]` to MEMORY.md.
  - The spawned subagent reads `skills/morning-briefing/SKILL.md` and delivers the full briefing via `message` tool.

- **Monday 9:00 AM IST** — Spawn `email-digest` and `git-summary` (weekly review), but only once per week.
  - Track with `[heartbeat:weekly:YYYY-Www]` in MEMORY.md.

## Subagent Behavior

When spawning:
- Pass a clear task: `"Run the morning-briefing skill. Read skills/morning-briefing/SKILL.md and follow all steps. Send output via message tool."`
- The subagent communicates directly with the user via the `message` tool.
- Do not wait for subagent result before returning `HEARTBEAT_OK`.
