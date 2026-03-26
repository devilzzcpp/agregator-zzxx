# Subscription Aggregator API

Тестовое задание: REST-сервис для учета онлайн-подписок пользователей.

## Требования

Для успешного запуска проекта, убедитесь, что у вас установлены следующие программы:

- `git`
- `docker` + `docker compose`
- `go` (1.25+)
- `make` (необязательно) 

## Установка и запуск

Все команды ниже выполняются в папке `backend`.

### Быстрый запуск через Makefile

1. Клонировать проект и перейти в backend

```bash
git clone <URL_РЕПОЗИТОРИЯ>
cd agregator-zzxx/backend
```

2. Первый запуск

```bash
make first_start
```

Команда `first_start` делает сразу все:

1. создает `.env` из `.env.example` (если файла нет)
2. устанавливает `goose`
3. поднимает PostgreSQL (`docker compose up -d db`)
4. применяет миграции
5. поднимает приложение (`docker compose up -d`)

После запуска:

- API: `http://localhost:8010`
- Swagger: `http://localhost:8010/swagger/index.html`
- DB: `jdbc:postgresql://localhost:1080/agregator`

#### Доступные команды Makefile

```bash
make help
make app-up
make app-down
make app-logs
make migrate-status
make migrate-up
make migrate-down
make run
```
### Ручной запуск без make

Если нужно запускать вручную:

1. Поднять контейнеры

```powershell
docker compose up -d
```

2. Применить миграции

```powershell
goose -dir ./migrations postgres "host=localhost port=1080 user=postgres password=secret dbname=agregator sslmode=disable" up
```

3. Локальный запуск

```powershell
docker compose up -d db #поднимет базу данных
go run ./cmd/server
```

## Swagger

Swagger защищен basic auth.

- URL: `http://localhost:8010/swagger/index.html`
- Логин: `SVG_LOGIN` из `.env` 
- Пароль: `SVG_PASSWORD` из `.env`


Для обновления Swagger:

```bash
go install github.com/swaggo/swag/cmd/swag@latest #устанавливает CLI Swagger
swag init -g cmd/server/main.go -o docs #обновляет документацию
```

## API

Базовый префикс: `/api/v1`

- `POST /subscriptions` - создать подписку
- `GET /subscriptions` - список подписок
- `GET /subscriptions/:id` - получить по id
- `PUT /subscriptions/:id` - обновить
- `DELETE /subscriptions/:id` - удалить
- `GET /subscriptions/total` - суммарная стоимость за период
- `GET /ping` - проверка, что сервис жив

## Конфиг

Шаблон переменных лежит в `.env.example`.

Основные параметры:

```env
HOST=0.0.0.0
PORT=8010

DB_HOST=localhost
DB_PORT=1080
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=agregator
DB_SSLMODE=disable

LOG_LEVEL=debug
LOG_FILE=./storage/logs/app.log
LOG_SIZE=10
LOG_BACKUP=5
LOG_AGE=30

SVG_LOGIN=svgadmin
SVG_PASSWORD=svgadmin

APP_TIMEZONE=Europe/Samara
```

## Тестирование

Есть интеграционный скрипт:

```powershell
./test.ps1
```

Он прогоняет основные happy-path и error-сценарии по API.

## Логи

- внутри контейнера: `/app/storage/logs`
- на хосте: `./storage/logs`

## Структура проекта

- `cmd/server` - entrypoint
- `internal/subscription` - handler/service/repository
- `internal/models` - модели БД
- `common` - конфиг, логгер, db
- `migrations` - SQL-миграции
- `docs` - swagger-файлы