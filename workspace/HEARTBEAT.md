# Periodic Tasks

Checked every 30 minutes. Only message the user when there's something noteworthy.

## Quick Tasks (respond directly)

- Check system health: RAM usage, disk space, CPU temp. Alert if RAM > 80% or disk > 90%.
- If it's between 8:00-8:30 AM IST and the morning briefing hasn't been sent today, trigger the morning-briefing skill.
- Review memory/MEMORY.md — prune entries older than 30 days (unless tagged [permanent]). Keep under 200 entries.

## Long Tasks (use spawn for async)

- If any pending reminders are due, deliver them to the user via message.
- If it's Monday at 9:00 AM IST, spawn the weekly email-digest and git-summary skills.
