.PHONY: help build-api build-telegram zip-api zip-telegram run test \
        migrate-up migrate-down migrate-status migrate-new \
        sqlc-generate docker-up docker-down install

# Variáveis
API_BINARY    = bootstrap-api
TELEGRAM_BINARY = bootstrap-telegram
API_ZIP       = lambda-api.zip
TELEGRAM_ZIP  = lambda-telegram.zip
MIGRATIONS_DIR = ./internal/pgstore/migrations
SQLC_CONFIG    = ./internal/pgstore/sqlc.yaml

ifneq (,$(wildcard .env))
    include .env
    export
endif

help:
	@echo "Comandos disponíveis:"
	@echo "  make install           - Instala dependências Go"
	@echo "  make run               - Roda a API localmente"
	@echo "  make test              - Roda os testes"
	@echo "  make build-api         - Compila binário da API para Lambda"
	@echo "  make build-telegram    - Compila binário do bot para Lambda"
	@echo "  make zip-api           - Cria zip da API para deploy"
	@echo "  make zip-telegram      - Cria zip do bot para deploy"
	@echo "  make docker-up         - Sobe o banco local (Docker)"
	@echo "  make docker-down       - Para o banco local"
	@echo "  make migrate-up        - Roda todas as migrations"
	@echo "  make migrate-down      - Reverte a última migration"
	@echo "  make migrate-status    - Status das migrations"
	@echo "  make migrate-new       - Cria nova migration (interativo)"
	@echo "  make sqlc-generate     - Gera código Go a partir dos SQLs"

# ─── Dependencies ──────────────────────────────────────────────────────────────

install:
	go mod download
	go mod tidy

# ─── Build ─────────────────────────────────────────────────────────────────────

build-api:
	@echo "Building API para Lambda..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(API_BINARY) ./cmd/api
	@echo "Build concluído: $(API_BINARY)"

build-telegram:
	@echo "Building bot Telegram para Lambda..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(TELEGRAM_BINARY) ./cmd/telegram
	@echo "Build concluído: $(TELEGRAM_BINARY)"

zip-api: build-api
	zip -j $(API_ZIP) $(API_BINARY)
	@echo "Zip criado: $(API_ZIP)"

zip-telegram: build-telegram
	zip -j $(TELEGRAM_ZIP) $(TELEGRAM_BINARY)
	@echo "Zip criado: $(TELEGRAM_ZIP)"

clean:
	rm -f $(API_BINARY) $(TELEGRAM_BINARY) $(API_ZIP) $(TELEGRAM_ZIP)

# ─── Run & Test ────────────────────────────────────────────────────────────────

run:
	go run ./cmd/api

test:
	go test -v ./...

# ─── Docker (dev local) ────────────────────────────────────────────────────────

docker-up:
	docker compose up -d
	@echo "Banco PostgreSQL disponível em localhost:5432"

docker-down:
	docker compose down

# ─── Migrations (Tern) ─────────────────────────────────────────────────────────

migrate-up:
	@if [ -z "$(DATABASE_URL)" ]; then echo "ERRO: DATABASE_URL não definida!"; exit 1; fi
	tern migrate -m $(MIGRATIONS_DIR) --conn-string "$(DATABASE_URL)"

migrate-down:
	@if [ -z "$(DATABASE_URL)" ]; then echo "ERRO: DATABASE_URL não definida!"; exit 1; fi
	tern migrate -m $(MIGRATIONS_DIR) --conn-string "$(DATABASE_URL)" -d -1

migrate-status:
	@if [ -z "$(DATABASE_URL)" ]; then echo "ERRO: DATABASE_URL não definida!"; exit 1; fi
	tern status -m $(MIGRATIONS_DIR) --conn-string "$(DATABASE_URL)"

migrate-new:
	@read -p "Nome da migration: " name; \
	tern new -m $(MIGRATIONS_DIR) $$name

# ─── SQLC ──────────────────────────────────────────────────────────────────────

sqlc-generate:
	sqlc generate -f $(SQLC_CONFIG)

.DEFAULT_GOAL := help
