# Первый этап: сборка приложения
FROM golang:1.23.3 AS builder

WORKDIR /app

# Устанавливаем pg_isready и очищаем кеш
RUN apt-get update && apt-get install -y postgresql-client \
    && rm -rf /var/lib/apt/lists/* && apt-get clean

# Отключаем CGO для статической компиляции
 ENV CGO_ENABLED=0

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download && go mod verify

# Копируем весь код
COPY . .

# Компилируем бинарник
RUN go build -o /app/shop_on_go ./cmd/main.go



# Второй этап: финальный образ (без лишних инструментов)
FROM alpine:latest

WORKDIR /app

# Устанавливаем postgresql-client и dos2unix
RUN apk add --no-cache postgresql-client dos2unix

COPY .env /app/.env

# Копируем бинарный файл из предыдущего этапа
COPY --from=builder /app/shop_on_go /app/shop_on_go

# Копируем wait-for-db.sh и делаем исполняемым
COPY --from=builder /app/wait-for-db.sh /app/wait-for-db.sh
RUN chmod +x /app/wait-for-db.sh

# Преобразуем формат строки в скрипте wait-for-db.sh в Unix-формат
RUN dos2unix /app/wait-for-db.sh

# 🔥 Копируем папку docs для Swagger
COPY --from=builder /app/docs /app/docs
# 🔥 Копируем шаблоны и статику
COPY --from=builder /app/static /app/static

# Запуск приложения
CMD ["/app/shop_on_go"]
