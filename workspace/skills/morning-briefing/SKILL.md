---
name: morning-briefing
description: Daily morning briefing with weather, calendar events, and today's tasks.
metadata: {"nanobot":{"emoji":"🌅","requires":{"bins":["curl","jq"]}}}
---

# Morning Briefing

Compile and deliver a morning digest via Telegram. Run this between 8:00–8:30 AM IST.

## Steps

### 1. Weather (wttr.in — no API key needed)

```bash
curl -s "wttr.in/Chennai?format=%l:+%c+%t+%h+%w"
```

### 2. Google Calendar (today's events)

If `GOOGLE_SERVICE_ACCOUNT_JSON` is set, fetch today's events:

```bash
# Get access token from service account
TOKEN=$(python3 -c "
import json, time, urllib.request, urllib.parse
import hmac, hashlib, base64
# Use google-auth if available, otherwise skip calendar
try:
    from google.oauth2 import service_account
    from google.auth.transport.requests import Request
    creds = service_account.Credentials.from_service_account_file(
        '$(echo $GOOGLE_SERVICE_ACCOUNT_JSON)',
        scopes=['https://www.googleapis.com/auth/calendar.readonly']
    )
    creds.refresh(Request())
    print(creds.token)
except:
    print('SKIP')
")

if [ "$TOKEN" != "SKIP" ]; then
  TODAY=$(date -u +%Y-%m-%dT00:00:00Z)
  # Compatible date arithmetic
  TOMORROW=$(python3 -c "from datetime import datetime, timedelta; print((datetime.utcnow() + timedelta(days=1)).strftime('%Y-%m-%dT00:00:00Z'))")
  curl -s -H "Authorization: Bearer $TOKEN" \
    "https://www.googleapis.com/calendar/v3/calendars/${GOOGLE_CALENDAR_ID:-primary}/events?timeMin=$TODAY&timeMax=$TOMORROW&singleEvents=true&orderBy=startTime" \
    | jq -r '.items[] | "• \(.start.dateTime // .start.date) — \(.summary)"'
fi
```

If Google Calendar is not configured, skip and note "Calendar not configured."

### 3. Todoist (today's tasks)

```bash
curl -s "https://api.todoist.com/rest/v2/tasks?filter=today" \
  -H "Authorization: Bearer $TODOIST_API_TOKEN" \
  | jq -r '.[] | "☐ \(.content) (p\(.priority))"'
```

If `TODOIST_API_TOKEN` is not set, skip and note "Todoist not configured."

### 4. Compose & Send

Combine all sections into a single message:

```
🌅 Good Morning, Prabhakaran!

🌤️ Weather: <weather output>

📅 Calendar:
<calendar events or "No events today">

✅ Today's Tasks:
<todoist tasks or "No tasks for today">

Have a great day! 🦐
```

Send this via the message tool to the user's Telegram.
