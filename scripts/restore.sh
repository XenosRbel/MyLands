#!/usr/bin/env bash
# Restores a custom-format dump into an EMPTY database. Usage: scripts/restore.sh <db-url> <file>
# Stop app servers first; the event worker resumes processing due events after restart.
set -euo pipefail
URL="${1:?database url}"; FILE="${2:?dump file}"
pg_restore --no-owner --exit-on-error --single-transaction --dbname="$URL" "$FILE"
psql "$URL" -Atc "SELECT 'restored: ' || count(*) || ' players' FROM players"
