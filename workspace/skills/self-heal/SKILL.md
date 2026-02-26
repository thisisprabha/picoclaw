---
name: self-heal
description: Monitor Pi3 system health, auto-restart services, and maintain workspace.
metadata: {"nanobot":{"emoji":"🩺","requires":{"bins":["free","df"]}}}
---

# Self-Heal

Monitor the Raspberry Pi 3 system health and take corrective actions.
Triggered by heartbeat every 30 minutes, or manually ("system health").

## Health Checks

### 1. Memory (RAM)

```bash
free -m | awk '/Mem:/ {printf "RAM: %dMB used / %dMB total (%d%%)\n", $3, $2, $3*100/$2}'
```

Alert if usage > 80%. Action: identify top memory consumers:

```bash
ps aux --sort=-%mem | head -6
```

### 2. Disk Space

```bash
df -h / | awk 'NR==2 {printf "Disk: %s used / %s total (%s)\n", $3, $2, $5}'
```

Alert if usage > 90%. Action: suggest cleanup of old logs/sessions:

```bash
find ~/.picoclaw/workspace/sessions -name "*.json" -mtime +30 | wc -l
```

### 3. CPU Temperature

```bash
cat /sys/class/thermal/thermal_zone0/temp 2>/dev/null | awk '{printf "CPU Temp: %.1f°C\n", $1/1000}'
```

Alert if > 75°C.

### 4. PicoClaw Process

```bash
pgrep -f "picoclaw gateway" > /dev/null && echo "PicoClaw: ✅ Running" || echo "PicoClaw: ❌ Not running"
```

### 5. Uptime

```bash
uptime -p
```

## Response Format

```
🩺 System Health Report

💾 RAM: 450MB / 1024MB (44%) ✅
💽 Disk: 5.2GB / 29GB (18%) ✅
🌡️ CPU Temp: 52.3°C ✅
🦐 PicoClaw: Running ✅
⏱️ Uptime: 3 days, 7 hours

Status: All systems nominal 🟢
```

If any check is in warning/critical state, include recommended actions.

## Auto-Healing Actions

- If PicoClaw is not running: `sudo systemctl restart picoclaw`
- If RAM > 90%: Kill zombie processes, clear old session files
- If disk > 95%: Remove session files older than 7 days, compress logs
- After any auto-healing action, log it to `memory/MEMORY.md` with timestamp

## Workspace Maintenance

On each health check, also:
- Verify `HEARTBEAT.md` exists and is readable
- Verify `memory/MEMORY.md` exists
- Check that skills directory has expected skills
- Log last health check timestamp to `state/last_health_check`
