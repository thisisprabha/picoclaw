# PicoClaw Mac -> Pi Build and Deploy Runbook

This guide builds PicoClaw on macOS and deploys the Linux ARM64 binary to Raspberry Pi.

## 1. Prerequisites

On Mac:

1. Install Go (same major/minor required by repo `go.mod`).
2. Have source code checked out locally.
3. Ensure network access to the Pi (`ssh`/`scp`).

On Pi:

1. PicoClaw repo exists at `~/picoclaw`.
2. systemd service `picoclaw` is installed and uses `~/picoclaw/build/picoclaw`.

## 2. Build on Mac

```bash
cd ~/Documents
git clone https://github.com/sipeed/picoclaw.git picoclaw-v2
cd picoclaw-v2
git checkout v0.2.0
git describe --tags --always
```

Expected: `v0.2.0`

Generate embedded workspace assets first:

```bash
make generate
```

Build Linux ARM64 binary with WhatsApp native tag:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
go build -tags "stdjson whatsapp_native" \
-o build/picoclaw-linux-arm64 ./cmd/picoclaw
```

## 3. Copy Binary to Pi

Copy to `/tmp` first:

```bash
scp ~/Documents/picoclaw-v2/build/picoclaw-linux-arm64 \
yoyoboy@<PI_IP>:/tmp/picoclaw-linux-arm64
```

## 4. Install Binary on Pi

```bash
ls -lh /tmp/picoclaw-linux-arm64
sudo systemctl stop picoclaw
mkdir -p ~/picoclaw/build
cp /tmp/picoclaw-linux-arm64 ~/picoclaw/build/picoclaw.new
chmod +x ~/picoclaw/build/picoclaw.new
mv ~/picoclaw/build/picoclaw.new ~/picoclaw/build/picoclaw
```

## 5. Configure Native WhatsApp

```bash
tmp=$(mktemp)
jq '.channels.whatsapp.enabled=true
  | .channels.whatsapp.use_native=true
  | .channels.whatsapp.allow_from=[]
  | del(.channels.whatsapp.bridge_url)' \
~/.picoclaw/config.json > "$tmp" && mv "$tmp" ~/.picoclaw/config.json
```

Disable bridge service if previously used:

```bash
sudo systemctl disable --now wa-bridge || true
```

## 6. Fix Cron Store if Corrupted

If you see `failed to load store: unexpected end of JSON input`:

```bash
mkdir -p ~/.picoclaw/workspace/cron
cp ~/.picoclaw/workspace/cron/jobs.json ~/.picoclaw/workspace/cron/jobs.json.bak.$(date +%s) 2>/dev/null || true
printf '{"version":1,"jobs":[]}\n' > ~/.picoclaw/workspace/cron/jobs.json
```

## 7. Start and Verify

```bash
sudo systemctl restart picoclaw
~/picoclaw/build/picoclaw status | sed -n '1,6p'
sudo journalctl -u picoclaw --since "2 min ago" --no-pager | egrep -i "whatsapp|native|qr|pair|ready|auth|error"
```

## 8. Troubleshooting

### A) `cmd/.../onboard/command.go: pattern workspace: no matching files found`

Cause: skipped `make generate`.

Fix:

```bash
make generate
```

### B) `fatal: Not a valid object name #`

Cause: copied inline comment into command.

Fix: run command without trailing `# ...` comment text.

### C) `scp: dest open ... Failure`

Cause: destination path issue/permissions.

Fix: copy to `/tmp` first, then move on Pi.

### D) `cp: ... Text file busy`

Cause: replacing binary while service is running.

Fix:

```bash
sudo systemctl stop picoclaw
mv ~/picoclaw/build/picoclaw.new ~/picoclaw/build/picoclaw
```

### E) `whatsapp native not compiled in; build with -tags whatsapp_native`

Cause: binary was built without native tag.

Fix: rebuild on Mac with:

```bash
go build -tags "stdjson whatsapp_native" -o build/picoclaw-linux-arm64 ./cmd/picoclaw
```

### F) Endless `WhatsApp read error ... websocket close 1006`

Cause: bridge mode still active or bridge disconnected.

Fix:

1. Disable `wa-bridge`.
2. Ensure config uses `use_native=true` and no `bridge_url`.
3. Restart `picoclaw`.

## 9. Optional Rollback

Keep old binary before replace:

```bash
cp ~/picoclaw/build/picoclaw ~/picoclaw/build/picoclaw.backup.$(date +%s)
```

