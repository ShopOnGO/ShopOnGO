# Названия переменных
PROJECT_NAME = shop_on_go
DOCKER_IMAGE = shop_on_go_image
DOCKER_CONTAINER = shop_on_go_container
POSTGRES_CONTAINER = go_shop_postgres

# Установки для команд
GO_CMD = go run cmd/main.go
DOCKER_BUILD = docker-compose build
DOCKER_UP = docker-compose up -d
DOCKER_DOWN = docker-compose down
DOCKER_LOGS = docker logs -f $(DOCKER_CONTAINER)
DOCKER_PS = docker ps
POSTGRES_DUMP = docker exec -t $(POSTGRES_CONTAINER) pg_dump -U postgres -d link > backup.sql
POSTGRES_LOAD = docker exec -i $(POSTGRES_CONTAINER) psql -U postgres -d link < backup.sql

# Стандартная цель - сборка проекта
all: build

# Сборка Docker образа
build:
	$(DOCKER_BUILD)

# Запуск контейнеров с сервисами
up: 
	$(DOCKER_UP)

# Остановка контейнеров
down:
	$(DOCKER_DOWN)

# Печать логов контейнера с приложением
logs:
	$(DOCKER_LOGS)

# Запуск PostgreSQL бэкапа
backup:
	$(POSTGRES_DUMP)

# Загрузка данных в PostgreSQL
restore:
	$(POSTGRES_LOAD)

# Статус работающих контейнеров
status:
	$(DOCKER_PS)

# Вход в контейнер приложения
bash:
	$(DOCKER_EXEC)

# Остановка всех контейнеров и очистка
clean:
	$(DOCKER_DOWN)
	docker system prune -f
