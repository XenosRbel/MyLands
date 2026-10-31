#!/usr/bin/env bash
# Logical, consistent backup (pg_dump uses a single snapshot). Usage: scripts/backup.sh <db-url> [outdir]
set -euo pipefail
URL="${1:?database url}"; OUT="${2:-backups}"
mkdir -p "$OUT"
FILE="$OUT/realmforge-$(date -u +%Y%m%dT%H%M%SZ).dump"
pg_dump --format=custom --no-owner --dbname="$URL" --file="$FILE"
pg_restore --list "$FILE" >/dev/null   # verify archive is readable
echo "$FILE"
