#!/bin/sh
# Compila para a Pi (arm64, sem CGO), envia e reinicia. Uso: sh scripts/deploy.sh
set -eu
cd "$(dirname "$0")/.."
go test ./...
mkdir -p bin/pi
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags=-s -o bin/pi/habit-idle ./cmd/habit-idle
ssh rpi "mkdir -p ~/apps/habit-idle/deploy"
scp -q bin/pi/habit-idle rpi:~/apps/habit-idle/habit-idle.new
scp -q deploy/* rpi:~/apps/habit-idle/deploy/
ssh rpi "sh ~/apps/habit-idle/deploy/install.sh && chmod +x ~/apps/habit-idle/habit-idle.new && mv ~/apps/habit-idle/habit-idle.new ~/apps/habit-idle/habit-idle && sudo systemctl restart habit-idle && sleep 1 && echo \"habit-idle na Pi: \$(systemctl is-active habit-idle)\""
