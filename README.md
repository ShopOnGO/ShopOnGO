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
