# AvitoTest - Сервис назначения ревьюеров для Pull Request'ов

## Запуск сервиса

Сервис запускается на порту **8080** (по умолчанию).

```bash
cd cmd/server
go run main.go
```

После запуска вы увидите:
```
2025/11/15 18:21:16 запуск сервера на  :8080
```

Это означает, что сервер успешно запущен и готов принимать запросы.

## Базовый URL

```
http://localhost:8080
```

---

## API Endpoints

### 1. Создать команду с участниками

**POST** `/team/add`

Создает команду и пользователей (или обновляет существующих).

**Request Body:**
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
    },
    {
      "user_id": "u3",
      "username": "Charlie",
      "is_active": true
    }
  ]
}
```

**Response 200 OK:**
```json
{}
```

**Response 409 Conflict** (если команда уже существует):
```json
{
  "error": {
    "code": "TEAM_EXISTS",
    "message": "команда с именем backend уже существует"
  }
}
```

---

### 2. Получить команду с участниками

**GET** `/team/get?team_name=backend`

**Query Parameters:**
- `team_name` (required) - имя команды

**Response 200 OK:**
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

**Response 404 Not Found:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "команда backend не найдена"
  }
}
```

---

### 3. Установить флаг активности пользователя

**POST** `/users/setIsActive`

**Request Body:**
```json
{
  "user_id": "u2",
  "is_active": false
}
```

**Response 200 OK:**
```json
{}
```

**Response 404 Not Found:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "user_id=u2"
  }
}
```

---

### 4. Получить PR'ы, где пользователь назначен ревьювером

**GET** `/users/getReview?user_id=u2`

**Query Parameters:**
- `user_id` (required) - идентификатор пользователя

**Response 200 OK:**
```json
[
  {
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add search",
    "author_id": "u1",
    "status": "OPEN"
  },
  {
    "pull_request_id": "pr-1002",
    "pull_request_name": "Fix bug",
    "author_id": "u3",
    "status": "MERGED"
  }
]
```

---

### 5. Создать PR и автоматически назначить ревьюверов

**POST** `/pullRequest/create`

Автоматически назначает до 2 активных ревьюверов из команды автора (исключая самого автора).

**Request Body:**
```json
{
  "pull_request_id": "pr-1001",
  "pull_request_name": "Add search feature",
  "author_id": "u1"
}
```

**Response 200 OK:**
```json
{
  "pull_request_id": "pr-1001",
  "pull_request_name": "Add search feature",
  "author_id": "u1",
  "status": "OPEN",
  "assigned_reviewers": ["u2", "u3"],
  "createdAt": "2025-11-15T18:21:16Z",
  "mergedAt": null
}
```

**Response 404 Not Found** (если автор не найден):
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "author_id=u1"
  }
}
```

**Response 409 Conflict** (если PR уже существует):
```json
{
  "error": {
    "code": "PR_EXISTS",
    "message": "pull_request_id=pr-1001"
  }
}
```

---

### 6. Пометить PR как MERGED

**POST** `/pullRequest/merge`

Идемпотентная операция - можно вызывать несколько раз без ошибок.

**Request Body:**
```json
{
  "pull_request_id": "pr-1001"
}
```

**Response 200 OK:**
```json
{
  "pull_request_id": "pr-1001",
  "pull_request_name": "Add search feature",
  "author_id": "u1",
  "status": "MERGED",
  "assigned_reviewers": ["u2", "u3"],
  "createdAt": "2025-11-15T18:21:16Z",
  "mergedAt": "2025-11-15T18:25:00Z"
}
```

**Response 404 Not Found:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "pr_id=pr-1001"
  }
}
```

---

### 7. Переназначить ревьювера

**POST** `/pullRequest/reassign`

Заменяет одного ревьювера на случайного активного участника из команды заменяемого ревьювера.

**Request Body:**
```json
{
  "pull_request_id": "pr-1001",
  "old_user_id": "u2"
}
```

**Response 200 OK:**
```json
{
  "pull_request_id": "pr-1001",
  "pull_request_name": "Add search feature",
  "author_id": "u1",
  "status": "OPEN",
  "assigned_reviewers": ["u3", "u4"],
  "createdAt": "2025-11-15T18:21:16Z",
  "mergedAt": null
}
```

**Response 404 Not Found:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "ревьювер 2 не назначен на PR"
  }
}
```

**Response 403 Forbidden** (если PR уже MERGED):
```json
{
  "error": {
    "code": "NOTASSIGNED",
    "message": "нельзя менять ревьюверов после MERGED"
  }
}
```

**Response 403 Forbidden** (если нет доступных кандидатов):
```json
{
  "error": {
    "code": "NOTASSIGNED",
    "message": "нет доступных кандидатов для замены ревьювера 2"
  }
}
```
