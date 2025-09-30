# tailornado backend dev.
## Описание проекта
   Backend- часть интернет магазина tailornado, разработанная на языке go с использованием gin + gorilla.mux
## Стек используемых технологий :
   Kafka, NGINX, Docker, MongoDb, PostgreSQL, Redis, JWT + REFRESH, OAUTH2, ELASTIC SEARCH, WEBSOCKET CHAT(BETA),GraphQL, SMTP NOTIFICATIONS, gRPC.
## роли пользователей и  описание их действий в системе :
## схема бд :[диаграмма зависимостей](https://dbdiagram.io/d/67e14f9975d75cc8443d6fe0)
## API : [API](https://drive.google.com/file/d/1057l-up2nKAML1gSnxQFe91tPad1152u/view?usp=sharing)

Паттерн "Издатель-Подписчик" (Pub/Sub): "eventbus"
3 слойная архитектура :
    service.go основная логика (model)
    repository.go - обращение к базе(conn, query, res)
    handler.go - rest API, проверка данных, результат операции.
