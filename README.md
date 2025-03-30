# ShopOnGO
GO startUP

для поднятия приложения необходимо:
установить .env
далее:
CREATE DATABASE link; // база в которой будут храниться все таблицы (чтобы создать - посмотри гпт)
docker-compose up --build
данные:
backup.sql
создан чтобы передавать данные таблиц , их будет несколько:
запись в файл
docker exec -t go_shop_postgres pg_dump -U postgres -d link > backup.sql
добавление из файла
docker exec -i go_shop_postgres psql -U postgres -d link < backup.sql


работа со swagger:
для обновления документации : swag init -g cmd/main.go
NEW : swag init -g cmd/main.go --parseDependency --parseInternal --output ./docs
для обращения  : http://localhost:8081/swagger/index.html


auto-migrations:
заменить строку в shop_on_go_container:
было : command: ./wait-for-db.sh ./shop_on_go
стало : command: ./wait-for-db.sh ./shop_on_go "migrate"

database :
postgres:
Теперь, даже если вы удалите контейнер (docker-compose down), все данные останутся в postgres-data,
и при следующем запуске PostgreSQL сможет восстановить их
volumes:
      - ./postgres-data:/data/postgres
(я пока что отключаю)

redis:
volumes:
      - redis_data:/data
(пока что отключаю)

нынешний функционал и архитектура:

repo: (entities)
repo: (entities)
repo- link : Create(link),GetByHash(hash),Update(link),Delete(id),GetById(id)  && queries: Count(),GetAll(limit, offset int)
repo- user : Create(user *User), FindByEmail(email string)
repo- stat : AddClick(linkId uint), GetStats(by string, from, to time.Time)
repo- products :  Create(product *Product)
                  GetByCategory(category *category.Category)
                  GetFeaturedProducts(amount uint, random bool) ([]Product, error)
                  GetByName(name string) ([]Product, error)
repo- category : Create(category *Category), GetCategories(),

handlers:
auth-handler : Login(), GoogleLogin(), Register(), Logout(), ChangePassword(), ChangeRole()
oauth-handler : HandleToken()
link-handler : CRUD,GetAll()
stat-handler : GetStat()

mv:
Chain : CORS,validator(in request),logger

Паттерн "Издатель-Подписчик" (Pub/Sub): "eventbus"
3 слойная архитектура :
    service.go основная логика (model)
    repository.go - обращение к базе(conn, query, res)
    handler.go - rest API, проверка данных, результат операции.


usage:
    config:godotenv
    db:gorm
    auth:simple jwt (no refresh)
    swagger
