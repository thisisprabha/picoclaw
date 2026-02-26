#!/bin/bash
# Deploy PicoClaw workspace to Raspberry Pi 3
# Push code to git from Mac, pull on Pi, then sync workspace files
#
# Usage: ./deploy.sh
# Assumes you've already pushed to git and built on the Pi

PI_USER="yoyoboy"
PI_HOST="192.168.1.23"
PI_WORKSPACE="/home/yoyoboy/.picoclaw/workspace"

echo "🦐 PicoClaw Workspace Sync"
echo "=========================="

# 1. Sync workspace files (skills, scripts, config templates)
echo "📤 Syncing workspace files to Pi..."
rsync -avz --exclude='.git' --exclude='sessions/' --exclude='state/' --exclude='capture/' \
  workspace/ ${PI_USER}@${PI_HOST}:${PI_WORKSPACE}/

# 2. Make scripts executable on Pi
echo "🔧 Setting permissions..."
ssh ${PI_USER}@${PI_HOST} "chmod +x ${PI_WORKSPACE}/scripts/*.sh 2>/dev/null || true"

# 3. Restart picoclaw service
echo "🔄 Restarting picoclaw service..."
ssh ${PI_USER}@${PI_HOST} "sudo systemctl restart picoclaw"

# 4. Verify
echo "🔍 Verifying..."
sleep 3
ssh ${PI_USER}@${PI_HOST} "pgrep -f 'picoclaw gateway' > /dev/null && echo '✅ PicoClaw is running' || echo '❌ PicoClaw failed to start'"

echo ""
echo "🦐 Sync complete!"
