#!/bin/sh

set -e
host="postgres"
port="5432"
user="postgres"

echo "Waiting PostgreSQL ($host:$port)..."

until PGPASSWORD=my_pass pg_isready -h "$host" -p "$port" -U "$user"; do
  echo "DB unavailable, waiting..."
  sleep 2
done

echo "DB ACCEPTING QUERIES, starting app!"
exec "$@"
