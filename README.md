# ShopOnGO
GO startUP

для поднятия приложения необходимо:
установить .env
далее:
docker-compose up --build
данные:
backup.sql
создан чтобы передавать данные таблиц , их будет несколько:
запись в файл
docker exec -t go_shop_postgres pg_dump -U postgres -d link > backup.sql
добавление из файла
docker exec -i go_shop_postgres psql -U postgres -d link < backup.sql


нынешний функционал и архитектура:

repo:
repo- link : Create(link),GetByHash(hash),Update(link),Delete(id),GetById(id)  && queries: Count(),GetAll(limit, offset int)
repo- user : Create(user *User),FindByEmail(email string)
repo- stat : AddClick(linkId uint),GetStats(by string, from, to time.Time)

handlers:
auth-handler : login(), register()
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
