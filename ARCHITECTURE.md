# PicoClaw Architecture and Pi Runbook

This document is the operational source for running PicoClaw on a Raspberry Pi.
It focuses on how the system is wired, where files live, how to test, and how to troubleshoot quickly.

## 1. Runtime Architecture

PicoClaw on Pi has two main layers:

1. Code repository: `/home/yoyoboy/picoclaw`
2. Runtime home: `/home/yoyoboy/.picoclaw`

Execution flow:

1. `systemd` starts `picoclaw gateway`
2. PicoClaw loads config from `~/.picoclaw/config.json`
3. PicoClaw auto-loads env files in this order (first value wins):
   1. `PICOCLAW_ENV_FILE` (if set)
   2. `~/.picoclaw/.env.picoclaw`
   3. `~/.picoclaw/.env`
   4. `./.env` (working dir)
   5. `<workspace>/.env.picoclaw`
   6. `<workspace>/.env`
4. Agent builds context from workspace markdown files and loaded skills
5. Agent executes tools (`exec`, `read_file`, `message`, etc.)

## 2. Source of Truth

Use these as the canonical runtime files:

- Config: `~/.picoclaw/config.json`
- Runtime workspace: `~/.picoclaw/workspace`
- Agent instructions: `~/.picoclaw/workspace/AGENTS.md` (plural)
- Skill definitions: `~/.picoclaw/workspace/skills/*/SKILL.md`
- Env secrets: `~/.picoclaw/.env.picoclaw`

Avoid keeping conflicting values in `~/.picoclaw/.env` unless required.

## 3. What Each Markdown File Does

Inside `~/.picoclaw/workspace`:

- `AGENTS.md`: behavior policy and execution rules for the agent
- `HEARTBEAT.md`: periodic task instructions
- `IDENTITY.md`: assistant identity and long-lived profile
- `USER.md`: user preferences and communication style
- `SOUL.md`: tone/philosophy guidance
- `memory/MEMORY.md`: persistent facts learned over time

## 4. Repo vs Runtime Files

Why both exist:

- `~/picoclaw/workspace/*`: template/default files tracked in git
- `~/.picoclaw/workspace/*`: active runtime files used by gateway

Typical sync from repo templates to runtime:

```bash
rsync -a ~/picoclaw/workspace/skills/ ~/.picoclaw/workspace/skills/
cp ~/picoclaw/workspace/AGENTS.md ~/.picoclaw/workspace/AGENTS.md
```

## 5. Service Setup (systemd)

Recommended unit fields:

- `WorkingDirectory=/home/yoyoboy/picoclaw`
- `ExecStart=/home/yoyoboy/picoclaw/build/picoclaw gateway`
- `EnvironmentFile=/home/yoyoboy/.picoclaw/.env.picoclaw`

Check current unit:

```bash
sudo systemctl cat picoclaw
```

Reload after edits:

```bash
sudo systemctl daemon-reload
sudo systemctl restart picoclaw
```

## 6. Build and Test Commands

### Build

```bash
cd ~/picoclaw
make build
./build/picoclaw status
```

### Core smoke tests

```bash
./build/picoclaw agent -s "cli:todoist-smoke" -m "Run todoist-manager skill now and return today's tasks."
./build/picoclaw agent -s "cli:email-smoke" -m "Run email-digest skill now and return the digest."
./build/picoclaw agent -s "cli:git-smoke" -m "Run git-summary skill now and return weekly summary."
```

### Direct API/env checks

```bash
source ~/.picoclaw/.env.picoclaw

# Todoist (v1 API)
curl -s "https://api.todoist.com/api/v1/tasks?filter=today" \
  -H "Authorization: Bearer $TODOIST_API_TOKEN" | jq '.results | length'

# IMAP login
python3 - <<'PY'
import os, imaplib
h=os.getenv("EMAIL_IMAP_HOST","imap.gmail.com")
p=int(os.getenv("EMAIL_IMAP_PORT","993"))
u=os.getenv("EMAIL_ADDRESS","")
pw=os.getenv("EMAIL_PASSWORD","")
m=imaplib.IMAP4_SSL(h,p); m.login(u,pw); print("IMAP OK"); m.logout()
PY

# Git repo validity from GIT_REPOS
echo "$GIT_REPOS"
for r in $(echo "$GIT_REPOS" | tr ',' ' '); do [ -d "$r/.git" ] && echo "OK $r" || echo "BAD $r"; done
```

## 7. Monitoring and Logs

Service health:

```bash
sudo systemctl status picoclaw --no-pager
```

Live logs:

```bash
journalctl -u picoclaw -f
```

Recent logs:

```bash
journalctl -u picoclaw -n 200 --no-pager
```

Gateway health endpoint:

```bash
curl -s http://127.0.0.1:18790/health
curl -s http://127.0.0.1:18790/ready
```

## 8. Common Troubleshooting

### `picoclaw: command not found`

Use direct binary path:

```bash
./build/picoclaw status
```

Or install:

```bash
make install
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Skill uses wrong endpoint or old behavior

- Ensure runtime `AGENTS.md` has explicit rules
- Verify runtime skill file under `~/.picoclaw/workspace/skills/...`
- Re-run with a fresh session key (`-s "cli:...-fresh"`)

### Git summary says repos are invalid

`GIT_REPOS` must be local absolute paths, not `owner/repo` names.

Example:

```bash
export GIT_REPOS=/home/yoyoboy/picoclaw,/home/yoyoboy/another-local-repo
```

### Env conflicts

If multiple env files define same variable, earlier-loaded file wins.
Keep one canonical file (`~/.picoclaw/.env.picoclaw`) and remove conflicting entries from others.

### GitHub CLI (`gh`) login with SSH

```bash
gh auth login -h github.com -p ssh -w
gh auth status
```

## 9. Update Workflow (Safe)

```bash
cd ~/picoclaw
git pull origin main
make build
sudo systemctl restart picoclaw
./build/picoclaw status
```

If pull fails due local template changes:

```bash
git stash push -- workspace/AGENT.md
git pull origin main
```
