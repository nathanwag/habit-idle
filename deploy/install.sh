#!/bin/sh
# Roda na Pi a cada deploy: serviço, backup semanal, Caddy, nome .local e
# Tailscale. Idempotente.
set -eu
APP=habit-idle
PORT=3002
TS_PORT=10000
SRC="$HOME/apps/$APP/deploy"

mkdir -p "$HOME/data/$APP"
install -m 755 "$SRC/backup.sh" "$HOME/apps/$APP/backup.sh"

changed=0
for unit in $APP.service $APP-backup.service $APP-backup.timer; do
  if ! cmp -s "$SRC/$unit" "/etc/systemd/system/$unit"; then
    sudo install -m 644 "$SRC/$unit" "/etc/systemd/system/$unit"
    changed=1
  fi
done
[ "$changed" = 1 ] && sudo systemctl daemon-reload
sudo systemctl enable -q $APP.service
sudo systemctl enable -q --now $APP-backup.timer

if ! grep -q "^$APP.local$" /etc/avahi/aliases; then
  echo "$APP.local" | sudo tee -a /etc/avahi/aliases >/dev/null
fi

if ! grep -q "http://$APP.local" /etc/caddy/Caddyfile; then
  printf '\nhttp://%s.local {\n\treverse_proxy 127.0.0.1:%s\n}\n' "$APP" "$PORT" | sudo tee -a /etc/caddy/Caddyfile >/dev/null
  sudo systemctl reload caddy
fi

if ! sudo tailscale serve status | grep -q ":$TS_PORT"; then
  sudo tailscale serve --bg --https=$TS_PORT http://127.0.0.1:$PORT
fi
