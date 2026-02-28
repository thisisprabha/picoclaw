---
name: morning-briefing
description: Daily productivity briefing (tasks + calendar + optional weather).
metadata: {"nanobot":{"emoji":"🌅","requires":{"bins":["curl","jq"]}}}
---

# Morning Briefing

Compile and deliver a single morning productivity digest via Telegram.
Keep it short and actionable. Do not include a news section.

## Steps

### 1. Weather (optional, wttr.in — no API key needed)

```bash
WEATHER_CITY="${WEATHER_CITY:-Chennai}"
WEATHER_LINE=$(curl -m 8 -s "wttr.in/${WEATHER_CITY}?format=%l:+%c+%t+%h+%w" || true)
[ -z "$WEATHER_LINE" ] && WEATHER_LINE="Weather unavailable"
echo "$WEATHER_LINE"
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

If Google Calendar is not configured, note "Calendar not configured."

### 3. Todoist (today's tasks)

```bash
curl -s "https://api.todoist.com/api/v1/tasks?filter=today" \
  -H "Authorization: Bearer $TODOIST_API_TOKEN" \
  | jq -r 'if type=="object" and has("results") then .results[] else .[] end | "☐ \(.content) (p\(.priority // 1))"'
```

If `TODOIST_API_TOKEN` is not set, note "Todoist not configured."

### 4. Compose & Send

Combine into one short message:

```
🌅 Good Morning, Prabhakaran!

✅ Today (Top priorities)
<top 5 tasks from Todoist, or "No tasks for today">

📅 Calendar (Today)
<events, or "No events today", or "Calendar not configured">

🌤️ Weather
<one-line weather or "Weather unavailable">
```

Rules:
- Productivity only. No news block.
- Single consolidated message (not multiple small messages).
- Keep response concise.
