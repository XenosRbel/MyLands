#!/usr/bin/env bash
# Starts a native PostgreSQL 16 cluster for local dev/tests when Docker is unavailable.
# Usage: scripts/dev-postgres.sh start|stop|status
set -euo pipefail
PGBIN="${PGBIN:-$(ls -d /usr/lib/postgresql/*/bin | sort -V | tail -1)}"
PGDATA="${PGDATA:-/var/tmp/realmforge-pg}"
PGPORT="${PGPORT:-5432}"
RUN_AS="${PG_RUN_AS:-postgres}"
as_pg() { if [ "$(id -u)" = "0" ]; then su "$RUN_AS" -s /bin/bash -c "$*"; else bash -c "$*"; fi; }
case "${1:-start}" in
  start)
    if [ "$(id -u)" = "0" ] && ! id "$RUN_AS" >/dev/null 2>&1; then useradd -r -m "$RUN_AS"; fi
    mkdir -p "$PGDATA" && { [ "$(id -u)" = "0" ] && chown "$RUN_AS" "$PGDATA" || true; }
    if [ ! -f "$PGDATA/PG_VERSION" ]; then
      as_pg "$PGBIN/initdb -D $PGDATA -U postgres --auth=trust -E UTF8 >/dev/null"
    fi
    as_pg "$PGBIN/pg_ctl -D $PGDATA -o '-p $PGPORT -k /tmp' -l $PGDATA/server.log -w start" || true
    for db in realmforge realmforge_test; do
      psql -h 127.0.0.1 -p "$PGPORT" -U postgres -tc "SELECT 1 FROM pg_database WHERE datname='$db'" | grep -q 1 \
        || psql -h 127.0.0.1 -p "$PGPORT" -U postgres -c "CREATE DATABASE $db" >/dev/null
    done
    echo "postgres ready on 127.0.0.1:$PGPORT (dbs: realmforge, realmforge_test)";;
  stop) as_pg "$PGBIN/pg_ctl -D $PGDATA -m fast stop";;
  status) as_pg "$PGBIN/pg_ctl -D $PGDATA status";;
esac
