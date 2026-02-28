# Periodic Tasks

Checked every few hours. Keep this minimal.
Only message the user for critical issues or due reminders.

## Quick Tasks (respond directly)

- Check system health quickly. Alert ONLY when critical:
  - RAM > 90%
  - Disk > 95%
  - CPU temp > 80C
- If reminders are due, deliver them.
- Do NOT send routine morning briefing from heartbeat.
- Do NOT include news in heartbeat output.
- Keep heartbeat response short and operational.

## Long Tasks (use spawn for async)

- If it's Monday at 9:00 AM IST, spawn weekly email-digest and git-summary only once for that day.
- If nothing requires attention, return exactly: `HEARTBEAT_OK`
