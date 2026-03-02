---
name: morning-briefing
description: Full daily productivity briefing — calendar, tasks, git, weather, commute, email.
metadata: {"nanobot":{"emoji":"🌅","requires":{"bins":["python3","curl","jq","git","gh"]}}}
---

# Morning Briefing

Compile and deliver one compact morning digest via Telegram.
Covers: Meetings · Todoist · Git (yesterday) · Weather · Commute · Unread email.

Run each step, collect results, then compose one message.

---

## Step 1: Today's Meetings (Google Calendar)

```bash
if [ -z "${GOOGLE_SERVICE_ACCOUNT_JSON:-}" ] || [ ! -f "${GOOGLE_SERVICE_ACCOUNT_JSON}" ]; then
  echo "CALENDAR_STATUS=not_configured"
else
python3 - <<'PY'
import os, sys, json
from datetime import datetime, timezone, timedelta

sa_file = os.environ.get('GOOGLE_SERVICE_ACCOUNT_JSON','')
cal_id  = os.environ.get('GOOGLE_CALENDAR_ID','primary')

try:
    import google.oauth2.service_account as sa
    import googleapiclient.discovery as discovery
except ImportError:
    print("CALENDAR_STATUS=missing_lib")
    sys.exit(0)

try:
    creds = sa.Credentials.from_service_account_file(
        sa_file,
        scopes=['https://www.googleapis.com/auth/calendar.readonly']
    )
    service = discovery.build('calendar', 'v3', credentials=creds, cache_discovery=False)
    now_utc = datetime.now(timezone.utc)
    start   = now_utc.replace(hour=0,  minute=0,  second=0,  microsecond=0).isoformat()
    end     = now_utc.replace(hour=23, minute=59, second=59, microsecond=0).isoformat()
    events  = service.events().list(
        calendarId=cal_id, timeMin=start, timeMax=end,
        maxResults=5, singleEvents=True, orderBy='startTime'
    ).execute().get('items', [])
    if not events:
        print("CALENDAR_STATUS=empty")
    else:
        print("CALENDAR_STATUS=ok")
        for e in events:
            t = e['start'].get('dateTime', e['start'].get('date',''))
            try:
                dt = datetime.fromisoformat(t)
                t_fmt = dt.strftime('%I:%M %p')
            except Exception:
                t_fmt = t
            print(f"MEETING={t_fmt} — {e.get('summary','(No title)')}")
except Exception as ex:
    print(f"CALENDAR_STATUS=error:{ex}")
PY
fi
```

---

## Step 2: Todoist Tasks (Today)

```bash
if [ -z "${TODOIST_API_TOKEN:-}" ]; then
  echo "TODOIST_STATUS=not_configured"
else
  RESP=$(curl -s --max-time 10 \
    "https://api.todoist.com/api/v1/tasks?filter=today" \
    -H "Authorization: Bearer $TODOIST_API_TOKEN")
  if echo "$RESP" | jq -e 'if type=="object" and has("results") then .results else . end | length > 0' > /dev/null 2>&1; then
    echo "TODOIST_STATUS=ok"
    echo "$RESP" | jq -r 'if type=="object" and has("results") then .results[] else .[] end | "TASK=[\(.priority // 1)p] \(.content)"' | head -7
  else
    COUNT=$(echo "$RESP" | jq -r 'if type=="object" and has("results") then .results else . end | length' 2>/dev/null || echo 0)
    if [ "$COUNT" = "0" ]; then
      echo "TODOIST_STATUS=empty"
    else
      echo "TODOIST_STATUS=error"
    fi
  fi
fi
```

---

## Step 3: Git Pushes (Yesterday)

```bash
if [ -z "${GIT_REPOS:-}" ]; then
  echo "GIT_STATUS=not_configured"
else
  SINCE_ISO=$(python3 - <<'PY'
from datetime import datetime, timedelta, timezone
d = datetime.now(timezone.utc) - timedelta(days=1)
d = d.replace(hour=0, minute=0, second=0, microsecond=0)
print(d.strftime("%Y-%m-%dT%H:%M:%SZ"))
PY
)
  UNTIL_ISO=$(python3 - <<'PY'
from datetime import datetime, timedelta, timezone
d = datetime.now(timezone.utc)
d = d.replace(hour=0, minute=0, second=0, microsecond=0)
print(d.strftime("%Y-%m-%dT%H:%M:%SZ"))
PY
)
  TOTAL=0
  TMP=$(mktemp)
  printf '%s\n' "$GIT_REPOS" | tr ',' '\n' > "$TMP"
  while IFS= read -r repo; do
    repo=$(echo "$repo" | xargs)
    [ -z "$repo" ] && continue
    if [ -d "$repo/.git" ]; then
      c=$(cd "$repo" && git rev-list --count --after="$SINCE_ISO" --before="$UNTIL_ISO" HEAD 2>/dev/null || echo 0)
      echo "GIT_REPO=$(basename "$repo"):$c"
      TOTAL=$((TOTAL + c))
      continue
    fi
    if printf '%s' "$repo" | grep -Eq '^(https://github\.com/|git@github\.com:)?[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+(\.git)?$'; then
      repo_norm=$(printf '%s' "$repo" | sed -E 's#^https://github\.com/##; s#^git@github\.com:##; s#\.git$##')
      if command -v gh > /dev/null 2>&1 && gh auth status > /dev/null 2>&1; then
        c=$(gh api "repos/$repo_norm/commits?since=${SINCE_ISO}&until=${UNTIL_ISO}&per_page=100" --jq 'length' 2>/dev/null || echo 0)
        case "$c" in ''|*[!0-9]*) c=0;; esac
        echo "GIT_REPO=$repo_norm:$c"
        TOTAL=$((TOTAL + c))
      else
        echo "GIT_REPO=$repo_norm:gh_not_auth"
      fi
    fi
  done < "$TMP"
  rm -f "$TMP"
  if [ "$TOTAL" -eq 0 ]; then
    echo "GIT_STATUS=empty"
  else
    echo "GIT_STATUS=ok"
  fi
  echo "GIT_TOTAL=$TOTAL"
fi
```

---

## Step 4: Weather (Chennai)

```bash
CITY="${WEATHER_CITY:-Chennai}"
CITY_ENC=$(python3 -c "import urllib.parse, sys; print(urllib.parse.quote(sys.argv[1]))" "$CITY" 2>/dev/null || echo "$CITY" | sed 's/ /+/g')
WEATHER=$(curl -s --max-time 8 "wttr.in/${CITY_ENC}?format=%l:+%c+%t,+feels+%f,+💧%h,+💨%w" 2>/dev/null)
if [ -z "$WEATHER" ]; then
  WEATHER="Weather unavailable (network?)"
fi
echo "WEATHER=$WEATHER"
```

---

## Step 5: Commute Time (Home → Hardy Tower, Sholinganallur)

Destination is fixed: **Hardy Tower, Ramanujan IT Park, Sholinganallur, Chennai**.

```bash
ORIGIN="${COMMUTE_HOME_ADDRESS:-Velachery,+Chennai}"
DEST="Hardy+Tower,+Ramanujan+IT+Park,+Sholinganallur,+Chennai"
ORIGIN_ENC=$(python3 -c "import urllib.parse, sys; print(urllib.parse.quote(sys.argv[1]))" "$ORIGIN" 2>/dev/null || echo "$ORIGIN" | sed 's/ /+/g')

if [ -z "${GOOGLE_MAPS_API_KEY:-}" ]; then
  echo "COMMUTE=Maps API key not set (add GOOGLE_MAPS_API_KEY)"
else
  MAPS=$(curl -s --max-time 10 \
    "https://maps.googleapis.com/maps/api/distancematrix/json?origins=${ORIGIN_ENC}&destinations=${DEST}&mode=driving&departure_time=now&key=${GOOGLE_MAPS_API_KEY}" 2>/dev/null)
  DURATION=$(echo "$MAPS" | jq -r '.rows[0].elements[0].duration_in_traffic.text // .rows[0].elements[0].duration.text // "unknown"' 2>/dev/null)
  DISTANCE=$(echo "$MAPS" | jq -r '.rows[0].elements[0].distance.text // ""' 2>/dev/null)
  STATUS=$(echo "$MAPS" | jq -r '.rows[0].elements[0].status // "UNKNOWN"' 2>/dev/null)
  if [ "$STATUS" = "OK" ]; then
    echo "COMMUTE=~${DURATION} (${DISTANCE})"
  else
    echo "COMMUTE=Could not fetch (status: $STATUS)"
  fi
fi
```

---

## Step 6: Unread Emails (Last 24h)

```bash
if [ -z "${EMAIL_IMAP_HOST:-}" ] || [ -z "${EMAIL_ADDRESS:-}" ] || [ -z "${EMAIL_PASSWORD:-}" ]; then
  echo "EMAIL_STATUS=not_configured"
else
python3 - <<'PY'
import imaplib, email, os, sys
from email.header import decode_header
from datetime import datetime, timedelta

host = os.environ.get('EMAIL_IMAP_HOST', 'imap.gmail.com')
port = int(os.environ.get('EMAIL_IMAP_PORT', '993'))
user = os.environ.get('EMAIL_ADDRESS', '')
pwd  = os.environ.get('EMAIL_PASSWORD', '')

try:
    mail = imaplib.IMAP4_SSL(host, port)
    mail.login(user, pwd)
except Exception as e:
    print(f"EMAIL_STATUS=auth_failed:{e}")
    sys.exit(0)

mail.select('inbox')
since = (datetime.now() - timedelta(hours=24)).strftime('%d-%b-%Y')
try:
    _, msgs = mail.search(None, f'(UNSEEN SINCE {since})')
    ids = msgs[0].split()
    total = len(ids)
    print(f"EMAIL_TOTAL={total}")
    if total == 0:
        print("EMAIL_STATUS=empty")
    else:
        print("EMAIL_STATUS=ok")
        for mid in ids[-3:]:
            _, data = mail.fetch(mid, '(RFC822)')
            msg = email.message_from_bytes(data[0][1])
            subj = decode_header(msg.get('Subject', '(no subject)'))[0][0]
            if isinstance(subj, bytes):
                subj = subj.decode('utf-8', errors='ignore')
            frm = msg.get('From', 'unknown').split('<')[0].strip()
            print(f"EMAIL={frm}: {subj[:60]}")
    mail.logout()
except Exception as e:
    print(f"EMAIL_STATUS=error:{e}")
PY
fi
```

---

## Step 7: Compose & Send

Collect all outputs above, then compose **one single Telegram message** using the template below.

Rules:
- Use only collected data — no invented values.
- Each section header must always appear, even if data is missing.
- For empty/unconfigured sections show the graceful note (see below).
- Maximum ~18 lines. Bullets only.
- Send via `message` tool once, not multiple times.

```
🌅 Morning Brief — Mon 02 Mar

📅 Meetings
  • <time> — <title>           (or: Not configured / None today)

✅ Tasks Today
  • [2p] <task content>        (or: None today / Not configured)

📊 Git Yesterday
  • <repo>: <N> commits        (or: No pushes / Not configured)

🌤 Weather
  • <wttr result>

🚗 Commute to Hardy Tower
  • ~<duration> (<distance>)   (or: Maps API key not set)

📧 Unread Email (24h)
  • Total: <N>
  • <sender>: <subject>        (show up to 3, or: Not configured / None)
```

Do not add any preamble, explanation, or postface outside this block.
