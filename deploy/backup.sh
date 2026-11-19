#!/bin/sh
# Backup completo em JSON (o mesmo de Ajustes › Exportar, que Importar restaura)
# no repositório privado de backups do homelab.
set -eu
REPO="$HOME/backups/homelab-backups"
mkdir -p "$REPO/habit-idle"
"$HOME/apps/habit-idle/habit-idle" backup "$REPO/habit-idle/habit-idle.json"
cd "$REPO"
git add -A
git diff --cached --quiet && exit 0
git commit -q -m "habit-idle $(date +%F)"
git push -q
