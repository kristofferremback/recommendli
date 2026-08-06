#!/bin/sh
set -eu

db_path="${SQLITE_DB_PATH:-/tmp/recommendli.sqlite}"
db_dir="$(dirname "$db_path")"

if [ "$db_dir" = "/" ]; then
  echo "SQLITE_DB_PATH must be inside a writable directory" >&2
  exit 1
fi

mkdir -p "$db_dir"
if ! su-exec recommendli:recommendli test -w "$db_dir"; then
  chown recommendli:recommendli "$db_dir"
fi
if [ -e "$db_path" ]; then
  chown recommendli:recommendli "$db_path"
fi

exec su-exec recommendli:recommendli /app/recommendli "$@"
