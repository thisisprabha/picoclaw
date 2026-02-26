#!/bin/bash
# PicoClaw Watchdog — runs via cron every 5 minutes
# Ensures picoclaw gateway stays alive and monitors system health
#
# Install: crontab -e → */5 * * * * /home/yoyoboy/.picoclaw/workspace/scripts/picoclaw-watchdog.sh >> /tmp/picoclaw-watchdog.log 2>&1

LOGFILE="/tmp/picoclaw-watchdog.log"
MAX_LOG_SIZE=1048576  # 1MB

# Rotate log if too large
if [ -f "$LOGFILE" ] && [ $(stat -f%z "$LOGFILE" 2>/dev/null || stat -c%s "$LOGFILE" 2>/dev/null) -gt $MAX_LOG_SIZE ]; then
  tail -100 "$LOGFILE" > "${LOGFILE}.tmp" && mv "${LOGFILE}.tmp" "$LOGFILE"
fi

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Watchdog check started"

# Check if picoclaw gateway is running
if ! pgrep -f "picoclaw gateway" > /dev/null 2>&1; then
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] ❌ PicoClaw gateway not running! Restarting..."
  sudo systemctl restart picoclaw
  sleep 5
  if pgrep -f "picoclaw gateway" > /dev/null 2>&1; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ✅ PicoClaw gateway restarted successfully"
  else
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ❌ Failed to restart PicoClaw gateway!"
  fi
else
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] ✅ PicoClaw gateway is running"
fi

# Check RAM usage
RAM_PERCENT=$(free | awk '/Mem:/ {printf "%d", $3*100/$2}')
if [ "$RAM_PERCENT" -gt 90 ]; then
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] ⚠️ RAM usage critical: ${RAM_PERCENT}%"
  # Kill zombie processes
  zombie_count=$(ps aux | awk '$8=="Z"' | wc -l)
  if [ "$zombie_count" -gt 0 ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Cleaning $zombie_count zombie processes"
    ps aux | awk '$8=="Z" {print $2}' | xargs -r kill -9
  fi
  # Clean old session files (older than 7 days)
  find ~/.picoclaw/workspace/sessions -name "*.json" -mtime +7 -delete 2>/dev/null
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] Cleaned old sessions"
elif [ "$RAM_PERCENT" -gt 80 ]; then
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] ⚠️ RAM usage high: ${RAM_PERCENT}%"
fi

# Check disk usage
DISK_PERCENT=$(df / | awk 'NR==2 {print $5}' | tr -d '%')
if [ "$DISK_PERCENT" -gt 90 ]; then
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] ⚠️ Disk usage high: ${DISK_PERCENT}%"
  # Clean logs older than 14 days
  find /tmp -name "picoclaw*" -mtime +14 -delete 2>/dev/null
  find ~/.picoclaw/workspace/sessions -name "*.json" -mtime +14 -delete 2>/dev/null
fi

# Check CPU temperature
if [ -f /sys/class/thermal/thermal_zone0/temp ]; then
  CPU_TEMP=$(awk '{printf "%.0f", $1/1000}' /sys/class/thermal/thermal_zone0/temp)
  if [ "$CPU_TEMP" -gt 75 ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ⚠️ CPU temperature high: ${CPU_TEMP}°C"
  fi
fi

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Watchdog check complete"
