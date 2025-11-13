# Avito Internship - PR Reviewer Assignment Service

Сервис для назначения ревьюверов Pull Request'ам на основе команд и пользователей.

## Описание

Сервис предоставляет REST API для управления командами, пользователями и Pull Request'ами. Автоматически назначает ревьюверов из команды автора PR.

## Структура проекта

```
.
├── cmd/
│   └── server/
│       └── main.go          
├── internal/
│   ├── api/
│   │   └── generated.go     
│   ├── handler/
│   │   └── handlers.go      
│   ├── models/
│   │   └── models.go        
│   ├── service/
│   │   └── service.go       
│   └── storage/
│       └── storage.go       
├── docker-compose.yml       
├── Dockerfile               
├── go.mod                   
├── Makefile                 
├── openapi.yml              
└── README.md                
```

## Запуск

### Docker Compose

1. Запустите сервис с базой данных:
   ```bash
   make up
   ```

   Или напрямую:
   ```bash
   docker-compose up --build
   ```

Сервис будет доступен на `http://localhost:8080`.

### Локально

1. Установите зависимости:
   ```bash
   go mod tidy
   ```

2. Запустите сервер:
   ```bash
   go run cmd/server/main.go
   ```

## API Endpoints

- `POST /team/add` - Добавить команду
- `GET /team/get?team_name=...` - Получить команду
- `POST /users/setIsActive` - Установить активность пользователя
- `POST /pullRequest/create` - Создать PR
- `POST /pullRequest/merge` - Слить PR
- `POST /pullRequest/reassign` - Переназначить ревьювера
- `GET /users/getReview?user_id=...` - Получить PR'ы пользователя

## Проверка

Используйте curl для тестирования:

```bash
# Добавить команду
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{"team_name": "team1", "members": [{"user_id": "u1", "username": "user1", "is_active": true}]}'

# Получить команду
curl "http://localhost:8080/team/get?team_name=team1"

# Создать PR
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{"pull_request_id": "pr1", "pull_request_name": "Fix bug", "author_id": "u1"}'
```