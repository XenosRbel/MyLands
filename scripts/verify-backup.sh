#!/usr/bin/env bash
# Backup → restore into a scratch database → compare row counts. Usage: scripts/verify-backup.sh <db-url>
set -euo pipefail
URL="${1:?database url}"
BASE="${URL%/*}"; Q="${URL#*\?}"; [ "$Q" = "$URL" ] && Q="" || Q="?$Q"
SCRATCH="rf_restore_check_$$"
FILE=$(scripts/backup.sh "$URL" "${TMPDIR:-/tmp}")
psql "$BASE/postgres$Q" -qc "CREATE DATABASE $SCRATCH"
trap 'psql "$BASE/postgres$Q" -qc "DROP DATABASE IF EXISTS $SCRATCH"' EXIT
scripts/restore.sh "$BASE/$SCRATCH$Q" "$FILE" >/dev/null
for t in accounts players cities game_events missions resource_ledger; do
  a=$(psql "$URL" -Atc "SELECT count(*) FROM $t"); b=$(psql "$BASE/$SCRATCH$Q" -Atc "SELECT count(*) FROM $t")
  [ "$a" = "$b" ] || { echo "MISMATCH $t: $a vs $b"; exit 1; }
  echo "ok $t $a"
done
echo "backup verified: $FILE"
