# PR Reviewer Service

Микросервис для автоматического назначения ревьюверов на Pull Request'ы. Сервис предоставляет REST API для управления командами, пользователями и автоматического назначения ревьюверов на PR.

## Быстрый старт

### Запуск с Docker Compose
```bash
docker-compose up --build
```

Сервис будет доступен по адресу: `http://localhost:8080`

### Проверка работоспособности
```bash
curl http://localhost:8080/health
```

## Основные возможности

- Управление командами - создание и получение команд с участниками
- Автоназначение ревьюверов - автоматическое назначение до 2 ревьюверов из команды автора
- Управление пользователями - установка активности, массовая деактивация
- Полный lifecycle PR - создание, мердж, переназначение ревьюверов
- Статистика - получение статистики по назначениям
- Идемпотентные операции - безопасные повторные вызовы

## Эндпоинты API

### Команды
- POST /team/add - Создать команду с участниками
- GET /team/get?team_name=name - Получить команду

### Пользователи
- POST /users/setIsActive - Установить активность пользователя
- GET /users/getReview?user_id=id - Получить PR пользователя
- POST /users/bulkDeactivate - Массовая деактивация команды

### Pull Requests
- POST /pullRequest/create - Создать PR с автоназначением
- POST /pullRequest/merge - Мердж PR (идемпотентный)
- POST /pullRequest/reassign - Переназначить ревьювера
- GET /pullRequest/get?pull_request_id=id - Получить PR

### Системные
- GET /health - Проверка здоровья сервиса
- GET /stats - Статистика назначений

## База данных

PostgreSQL с автоматическим созданием таблиц через миграции:
- users - пользователи
- teams - команды  
- pull_requests - PR с назначенными ревьюверами

## Тестирование

### Интеграционные тесты
```bash
go test -v -tags=integration simple_integration_test.go
```

### Примеры использования

#### Для Linux/Mac:
```bash
curl -X POST http://localhost:8080/team/add -H "Content-Type: application/json" -d '{
  "team_name": "dev-team",
  "members": [
    {"user_id": "u1", "username": "Polina", "is_active": true},
    {"user_id": "u2", "username": "Dasha", "is_active": true}
  ]
}'
```

#### Для Windows:
```bash
curl -X POST http://localhost:8080/team/add -H "Content-Type: application/json" -d "{\"team_name\":\"dev-team\",\"members\":[{\"user_id\":\"u1\",\"username\":\"Polina\",\"is_active\":true},{\"user_id\":\"u2\",\"username\":\"Dasha\",\"is_active\":true}]}"
```

Создание PR:
```bash
curl -X POST http://localhost:8080/pullRequest/create -H "Content-Type: application/json" -d "{\"pull_request_id\":\"pr-1\",\"pull_request_name\":\"Test Feature\",\"author_id\":\"u1\"}"
```

Получение статистики:
```bash
curl http://localhost:8080/stats
```

## Дополнительные функции

### Автоматическое назначение ревьюверов
- Назначает до 2 активных ревьюверов из команды автора
- Исключает автора из списка кандидатов
- Учитывает флаг is_active пользователей

### Безопасное переназначение
- Автоматическая замена деактивированных ревьюверов
- Сохранение ограничения до 2 ревьюверов
- Валидация доменных правил

### Массовая деактивация
- Атомарная деактивация всех пользователей команды
- Автоматическое переназначение открытых PR
- Транзакционная безопасность

## Технические детали

- Язык: Go 1.25+
- База данных: PostgreSQL 15
- Контейнеризация: Docker + Docker Compose
- Тестирование: Integration tests

## Конфигурация линтера

Проект использует `golangci-lint` с конфигурацией в `.golangci.yml`:

```yaml
linters:
  disable-all: true
  enable:
    - gofmt
    - govet
    - staticcheck
    - gosimple
    - unused

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - dupl

output:
  format: colored-line-number
```

### Запуск линтера
```bash
# Установка
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Запуск
golangci-lint run ./...
```

## Структура проекта
```
pr-reviewer-service/
├── cmd/server/main.go         # Точка входа
├── internal/
│   ├── handlers/              # HTTP обработчики
│   ├── storage/               # Работа с БД
│   ├── models/                # Модели данных
│   └── config/                # Конфигурация
├── migrations/                # Миграции БД
├── docker-compose.yml         # Docker композ
└── README.md                  # Документация
```

## Выполненные дополнительные задания

1. Эндпоинт статистики (/stats) - статистика назначений по пользователям
2. Массовая деактивация (/users/bulkDeactivate) - безопасное переназначение PR
3. Интеграционное тестирование - тесты базовой функциональности
4. Конфигурация линтера - настройка golangci-lint

## Примеры работы

### Полный workflow

1. Создаем команду:
```bash
curl -X POST http://localhost:8080/team/add -H "Content-Type: application/json" -d '{
  "team_name": "backend",
  "members": [
    {"user_id": "u1", "username": "Polina", "is_active": true},
    {"user_id": "u2", "username": "Dasha", "is_active": true},
    {"user_id": "u3", "username": "Sonya", "is_active": true}
  ]
}'
```

2. Создаем PR - автоматически назначатся 2 ревьювера:
```bash
curl -X POST http://localhost:8080/pullRequest/create -H "Content-Type: application/json" -d '{
  "pull_request_id": "pr-1", 
  "pull_request_name": "Add auth",
  "author_id": "u1"
}'
```

3. Переназначаем ревьювера:
```bash
curl -X POST http://localhost:8080/pullRequest/reassign -H "Content-Type: application/json" -d '{
  "pull_request_id": "pr-1",
  "old_user_id": "u2"
}'
```

4. Мерджим PR:
```bash
curl -X POST http://localhost:8080/pullRequest/merge -H "Content-Type: application/json" -d '{
  "pull_request_id": "pr-1"
}'
```

## Особенности реализации

- **Случайный выбор ревьюверов** - равномерное распределение нагрузки
- **Транзакционность** - массовые операции в транзакциях
- **Валидация** - проверка всех бизнес-правил
- **Идемпотентность** - безопасные повторные вызовы мерджа

## Производительность

- Время ответа: < 50ms для большинства запросов
- Поддерживаемая нагрузка: 20+ RPS
- Потребление памяти: ~50MB

## Локальная разработка

```bash
# Запуск без Docker
go run cmd/server/main.go

# Тестирование
go test ./...

# Линтинг
golangci-lint run
```

## Деплой

```bash
# Продакшен сборка
docker build -t pr-reviewer .

# Запуск с внешней БД
docker run -p 8080:8080 -e DB_HOST=prod-db pr-reviewer
```

## Решение проблем

### Ошибка подключения к БД
Убедитесь, что PostgreSQL запущен:
```bash
docker-compose ps
```

### Проблемы с JSON в Windows
Используйте файлы для передачи данных:
```bash
curl -X POST http://localhost:8080/team/add -H "Content-Type: application/json" -d "@team_data.json"
```

## Контакты
Разработчик: Полина Карцева  
Для обратной связи: email - polypoly098765432@mail.ru tg - @Vakkrehevn