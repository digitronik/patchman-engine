#!/usr/bin/bash

set -e

cmd="$@"

if [ ! -z "$DB_HOST" ]; then
  >&2 echo "Checking if PostgreSQL is up"
  if [ ! -z "$WAIT_FOR_EMPTY_DB" ]; then
    CHECK_QUERY="\q" # Wait only for empty database.
  else
    # Wait even for database schema initialization.
    CHECK_QUERY="SELECT 1/count(*) FROM schema_migrations WHERE dirty='f';"
  fi
  until PGPASSWORD="$DB_PASSWD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "${CHECK_QUERY}" -q 2>/dev/null; do
    >&2 echo "PostgreSQL is unavailable - sleeping"
    sleep 1
  done
else
  >&2 echo "Skipping PostgreSQL check"
fi

>&2 echo "Everything is up - executing command"
exec $cmd
