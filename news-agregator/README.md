# news-agregator
## Порядок использования
1. В директории **db_docker** выполнить:
**$docker-compose up**. 
Это создаст контейнер с БД.
2. Далее применить миграции. Схема БД расположена в директории schema. Миграции можно выполнить при помощи утилиты **$migrate**:
- Применить миграции: 
**$migrate -path ./schema -database 'postgres://postgres:toor-555@localhost:5432/gonewsagregator?sslmode=disable' up**
- Откатить миграции: 
**$migrate -path ./schema -database 'postgres://postgres:toor-555@localhost:5432/gonewsagregator?sslmode=disable' down**
Для простоты можно воспользоваться утилитой **$make**:
- Применить миграции: **$make migrate_up**
- Откатить миграции: **$migrate_down**
3. Сбилдить приложение при помощи команды **$make** - это создаст бинрный файл в директории **build**.
4. Запустить созданый бинарник. 