.PHONY: help ensure-env goose-install db-up db-down app-up app-down app-logs migrate-up migrate-down migrate-status run first_start

help:
	@echo "make ensure-env      - создать .env из .env.example если нет"
	@echo "make goose-install   - установить goose CLI"
	@echo "make db-up           - поднять только postgres"
	@echo "make app-up          - поднять db + app через docker compose"
	@echo "make migrate-up      - применить миграции"
	@echo "make migrate-down    - откатить последнюю миграцию"
	@echo "make migrate-status  - показать статус миграций"
	@echo "make first_start     - env + goose + db + migrate + app"

ensure-env:
	@pwsh -NoProfile -Command "if (!(Test-Path '.env')) { Copy-Item '.env.example' '.env'; Write-Host 'created .env from .env.example' } else { Write-Host '.env already exists' }"

goose-install:
	@pwsh -NoProfile -Command "go install github.com/pressly/goose/v3/cmd/goose@latest; $$gopath = (go env GOPATH).Trim(); $$goose = Join-Path $$gopath 'bin/goose.exe'; if (Test-Path $$goose) { Write-Host \"goose installed: $$goose\" } else { Write-Host 'goose installed, but not found in GOPATH/bin PATH yet'; exit 1 }"

db-up: ensure-env
	@docker compose up -d db

db-down:
	@docker compose stop db

app-up: ensure-env
	@docker compose up -d

app-down:
	@docker compose down

app-logs:
	@docker compose logs -f app

run: ensure-env
	@go run ./cmd/server

migrate-up: ensure-env
	@pwsh -NoProfile -Command "$$kv = @{}; Get-Content '.env' | ForEach-Object { if ($$_ -match '^\s*#' -or $$_ -match '^\s*$$') { return }; if ($$_ -match '^\s*([^=]+)=(.*)$$') { $$k=$$matches[1].Trim(); $$v=$$matches[2].Trim(); if ($$v -match '^(.*)\s+#') { $$v=$$matches[1].Trim() }; $$kv[$$k]=$$v } }; $$dsn = \"host=$$($$kv.DB_HOST) port=$$($$kv.DB_PORT) user=$$($$kv.DB_USER) password=$$($$kv.DB_PASSWORD) dbname=$$($$kv.DB_NAME) sslmode=$$($$kv.DB_SSLMODE)\"; $$goose=(Get-Command goose -ErrorAction SilentlyContinue).Source; if (!$$goose) { $$goose = Join-Path ((go env GOPATH).Trim()) 'bin/goose.exe' }; & $$goose -dir ./migrations postgres $$dsn up"

migrate-down: ensure-env
	@pwsh -NoProfile -Command "$$kv = @{}; Get-Content '.env' | ForEach-Object { if ($$_ -match '^\s*#' -or $$_ -match '^\s*$$') { return }; if ($$_ -match '^\s*([^=]+)=(.*)$$') { $$k=$$matches[1].Trim(); $$v=$$matches[2].Trim(); if ($$v -match '^(.*)\s+#') { $$v=$$matches[1].Trim() }; $$kv[$$k]=$$v } }; $$dsn = \"host=$$($$kv.DB_HOST) port=$$($$kv.DB_PORT) user=$$($$kv.DB_USER) password=$$($$kv.DB_PASSWORD) dbname=$$($$kv.DB_NAME) sslmode=$$($$kv.DB_SSLMODE)\"; $$goose=(Get-Command goose -ErrorAction SilentlyContinue).Source; if (!$$goose) { $$goose = Join-Path ((go env GOPATH).Trim()) 'bin/goose.exe' }; & $$goose -dir ./migrations postgres $$dsn down"

migrate-status: ensure-env
	@pwsh -NoProfile -Command "$$kv = @{}; Get-Content '.env' | ForEach-Object { if ($$_ -match '^\s*#' -or $$_ -match '^\s*$$') { return }; if ($$_ -match '^\s*([^=]+)=(.*)$$') { $$k=$$matches[1].Trim(); $$v=$$matches[2].Trim(); if ($$v -match '^(.*)\s+#') { $$v=$$matches[1].Trim() }; $$kv[$$k]=$$v } }; $$dsn = \"host=$$($$kv.DB_HOST) port=$$($$kv.DB_PORT) user=$$($$kv.DB_USER) password=$$($$kv.DB_PASSWORD) dbname=$$($$kv.DB_NAME) sslmode=$$($$kv.DB_SSLMODE)\"; $$goose=(Get-Command goose -ErrorAction SilentlyContinue).Source; if (!$$goose) { $$goose = Join-Path ((go env GOPATH).Trim()) 'bin/goose.exe' }; & $$goose -dir ./migrations postgres $$dsn status"

first_start: ensure-env goose-install db-up migrate-up app-up
	@echo "first_start completed: db + migrations + app are up"
