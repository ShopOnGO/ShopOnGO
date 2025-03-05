#!/bin/sh

set -e
#данные из  DSN(ENV)
host="postgres"
port="5432"
user="postgres"

echo "Ожидание PostgreSQL ($host:$port)..."
#проверка бд на доступность подключения + инициализации
until PGPASSWORD=my_pass pg_isready -h "$host" -p "$port" -U "$user"; do
  echo "БД недоступна, жду..."
  sleep 2
done

echo "БД доступна, запускаем приложение!"
exec "$@"
