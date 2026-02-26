---
name: email-digest
description: Fetch and summarize recent emails via IMAP. Designed for weekly use.
metadata: {"nanobot":{"emoji":"📧","requires":{"bins":["python3"]}}}
---

# Email Digest

Connect to an IMAP mailbox, fetch recent unread emails, and generate a summary.
Best triggered weekly (Monday mornings) via heartbeat or cron.

## Required Environment Variables

- `EMAIL_IMAP_HOST` — IMAP server (e.g., `imap.gmail.com`)
- `EMAIL_IMAP_PORT` — Port (usually `993`)
- `EMAIL_ADDRESS` — Your email address
- `EMAIL_PASSWORD` — App password (for Gmail, use App Passwords)

## Steps

### 1. Fetch recent emails

Use this Python script to fetch the last 7 days of emails:

```bash
python3 -c "
import imaplib, email, os
from email.header import decode_header
from datetime import datetime, timedelta

host = os.environ.get('EMAIL_IMAP_HOST', 'imap.gmail.com')
port = int(os.environ.get('EMAIL_IMAP_PORT', '993'))
user = os.environ.get('EMAIL_ADDRESS', '')
pwd = os.environ.get('EMAIL_PASSWORD', '')

if not user or not pwd:
    print('EMAIL_ADDRESS or EMAIL_PASSWORD not set')
    exit(0)

mail = imaplib.IMAP4_SSL(host, port)
mail.login(user, pwd)
mail.select('inbox')

since = (datetime.now() - timedelta(days=7)).strftime('%d-%b-%Y')
_, msgs = mail.search(None, f'(SINCE {since})')
ids = msgs[0].split()[-20:]  # Last 20 emails max

summaries = []
for mid in ids:
    _, data = mail.fetch(mid, '(RFC822)')
    msg = email.message_from_bytes(data[0][1])
    subj = decode_header(msg['Subject'])[0][0]
    if isinstance(subj, bytes):
        subj = subj.decode('utf-8', errors='ignore')
    frm = msg['From']
    date = msg['Date']
    summaries.append(f'From: {frm}\nSubject: {subj}\nDate: {date}')

mail.logout()
print('\n---\n'.join(summaries) if summaries else 'No new emails in the last 7 days.')
"
```

### 2. Summarize with AI

Take the raw email subjects/senders list and summarize into:

```
📧 Weekly Email Digest (last 7 days)

Found X emails. Key highlights:
• <important email 1>
• <important email 2>
• ...

Action items:
• <anything requiring response>

Low priority (skippable):
• <newsletters, notifications>
```

### 3. Send via message tool

Deliver the summary to the user via Telegram.
