# LinkForge

Сервис коротких ссылок. Учебный проект, на котором собираем настоящий бэкенд: HTTP API, PostgreSQL, Redis, миграции, Docker, Kubernetes, CI.

## Стек

- Go 1.26
- PostgreSQL 16 (драйвер `pgx/v5`)
- Redis 7 (`go-redis/v9`)
- chi router
- golang-migrate (миграции БД)
- Docker compose (локальная инфра)
- GitHub Actions (CI)

## Требования

- Go 1.26+
- Docker + docker compose
- `migrate` CLI: `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

## Быстрый старт

```bash
# поднять postgres + redis
make up

# применить миграции
make migrate-up

# запустить API
make run
```

API будет на `http://localhost:8080`.

### Создать ссылку

```bash
curl -X POST http://localhost:8080/api/v1/links \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com/some/long/path"}'
```

Ответ:

```json
{ "code": "1f", "short_url": "http://localhost:8080/1f", "url": "https://example.com/some/long/path" }
```

### Перейти по короткой ссылке

```bash
curl -i http://localhost:8080/1f
# 302 Found, Location: https://example.com/some/long/path
```

## Структура

```
cmd/api/            точка входа HTTP-сервиса
internal/config/    загрузка конфигурации из env
internal/domain/    доменные типы (Link)
internal/storage/   доступ к Postgres и Redis
internal/service/   бизнес-логика (генерация code, валидация URL)
internal/http/      HTTP роутер, handlers, middleware
migrations/         SQL миграции (golang-migrate)
deploy/k8s/         Kubernetes манифесты (TODO)
```

## Воркфлоу разработки

1. Берём issue из GitHub
2. Создаём ветку `feature/<issue-number>-<slug>`
3. Пишем код, тесты, прогоняем `make test lint`
4. Открываем PR — CI должен пройти зелёным
5. Получаем ревью, правим
6. Squash merge в `main`

## Что осознанно НЕ сделано (это будущие таски)

- Kubernetes манифесты (Deployment / Service / ConfigMap / Ingress)
- Helm chart
- HTTP-уровневые интеграционные тесты с `httptest`
- Worker для агрегации аналитики кликов
- Rate limiting через Redis
- Аутентификация / API-ключи
- Метрики (Prometheus) и трейсинг (OpenTelemetry)
- Структурированный конфиг с валидацией (например через `viper` или `envconfig`)
- Идемпотентность создания ссылок (если уже есть такой URL — вернуть существующий)
- Кастомные коды (alias)
- TTL и удаление просроченных ссылок
- Аналитика кликов
- Health check для зависимостей (postgres/redis) в `/healthz`

Каждый из этих пунктов станет отдельным issue.
