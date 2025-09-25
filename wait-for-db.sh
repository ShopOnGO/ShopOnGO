#!/bin/sh

set -e

# Проверяем, установлена ли переменная окружения DSN
if [ -z "$DSN" ]; then
  echo "Error: DSN environment variable is not set." # <-- Исправлено
  exit 1
fi

echo "Waiting for Neon PostgreSQL database..."

# Используем pg_isready с переменной, полученной из окружения
until pg_isready -d "$DSN" -q; do
  echo "DB unavailable, waiting..."
  sleep 2
done

echo "DB ACCEPTING QUERIES, starting app!"
exec "$@"
