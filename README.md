# Сервис назначения ревьюеров

Сервис для автоматического назначения ревьюеров на Pull Request'ы

## Установка и запуск

1. Склонируйте репозиторий 
```bash
git clone https://github.com/MAPiryazev/AvitoTest
```
2. Перейдите в папку с проектом
```bash
cd AvitoTest
```
3. Соберите и запустите сервис
```bash
docker-compose build
docker-compose up -d
```

Сервер запустится на порту **8080**. Нужно чтобы в environment/.env были указаны параметры подключения к БД.

### Команды

POST /team/add - создать команду с участниками  
GET /team/get?team_name=название - получить команду  

### Пользователи

POST /users/setIsActive - установить активность пользователя  
GET /users/getReview?user_id=id - получить PR где пользователь ревьювер  

### Pull Request

POST /pullRequest/create - создать PR и назначить ревьюверов  
POST /pullRequest/merge - пометить PR как выполненный  
POST /pullRequest/reassign - переназначить ревьювера  

## Примеры

Создать команду:
**POST /team/add**
```json
{
  "team_name": "backend",
  "members": [
    {
      "user_id": "u1",
      "username": "Alice",
      "is_active": true
    },
    {
      "user_id": "u2",
      "username": "Bob",
      "is_active": true
    }
  ]
}
```

Установить активность пользователя:
**POST /users/setIsActive**
```json
{
  "user_id": "u1",
  "is_active": false
}
```

Создать PR:
**POST /pullRequest/create**
```json
{
  "pull_request_id": "pr-1",
  "pull_request_name": "Fix bug",
  "author_id": "u1"
}
```

Переназначить ревьювера:
**POST /pullRequest/reassign**
```json
{
  "pull_request_id": "pr-1",
  "old_user_id": "u2"
}
```

## Дополнительные задания
Нагрузочные тесты находятся в tests/load/. запуск: `make test`
Конфигурация линтера описана в `.golangci.yml`

## Возможные улучшения 
 - Поддержка авторизации пользователей 
 - Метрики и мониторинг параметров api и БД с помощью prometheus и grafana
 - Redis для статистики и часто читаемых данных