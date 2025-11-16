# PR Reviewer Service

Микросервис для автоматического назначения ревьюверов на Pull Request'ы.

## Быстрый старт

```bash
# Запуск с Docker Compose
docker-compose up --build

# Или сборка и запуск вручную
go build -o bin/server cmd/server/main.go
./bin/server