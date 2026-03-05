.PHONY: help build-api build-telegram zip-api zip-telegram run test \
        migrate-up migrate-down migrate-status migrate-new \
        sqlc-generate docker-up docker-down install \
        deploy deploy-lambda grant-apigw deploy-api

# Variáveis
API_BINARY    = bootstrap-api
TELEGRAM_BINARY = bootstrap-telegram
API_ZIP       = lambda-api.zip
TELEGRAM_ZIP  = lambda-telegram.zip
MIGRATIONS_DIR = ./internal/pgstore/migrations
SQLC_CONFIG    = ./internal/pgstore/sqlc.yaml

# ─── AWS / Deploy ──────────────────────────────────────────────────────────────
AWS_REGION           ?= us-east-1
AWS_ACCOUNT_ID        = 738827671264
LAMBDA_FUNCTION_NAME  = minha-api-go
LAMBDA_ROLE_ARN       = arn:aws:iam::$(AWS_ACCOUNT_ID):role/lambda-execution-role
API_GATEWAY_ID        = jf1shm9leh
API_GATEWAY_STAGE     = dev

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
	@echo "  make deploy            - Build + deploy completo para AWS Lambda"
	@echo "  make deploy-lambda     - Cria ou atualiza apenas a função Lambda"
	@echo "  make grant-apigw       - Adiciona permissão API Gateway → Lambda"
	@echo "  make deploy-api        - Cria deployment no API Gateway (stage: dev)"
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
	cp $(API_BINARY) bootstrap
	zip -j $(API_ZIP) bootstrap
	rm -f bootstrap
	@echo "Zip criado: $(API_ZIP)"

zip-telegram: build-telegram
	cp $(TELEGRAM_BINARY) bootstrap
	zip -j $(TELEGRAM_ZIP) bootstrap
	rm -f bootstrap
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

# ─── Deploy AWS Lambda ─────────────────────────────────────────────────────────

deploy-lambda: zip-api
	@echo "Verificando função Lambda '$(LAMBDA_FUNCTION_NAME)'..."
	@if aws lambda get-function --function-name $(LAMBDA_FUNCTION_NAME) --region $(AWS_REGION) > /dev/null 2>&1; then \
		echo "→ Atualizando código..."; \
		aws lambda update-function-code \
			--function-name $(LAMBDA_FUNCTION_NAME) \
			--zip-file fileb://$(API_ZIP) \
			--region $(AWS_REGION) > /dev/null; \
	else \
		echo "→ Criando função Lambda..."; \
		aws lambda create-function \
			--function-name $(LAMBDA_FUNCTION_NAME) \
			--runtime provided.al2023 \
			--role $(LAMBDA_ROLE_ARN) \
			--handler bootstrap \
			--zip-file fileb://$(API_ZIP) \
			--region $(AWS_REGION) > /dev/null; \
	fi
	@echo "Lambda OK."

grant-apigw:
	@echo "Adicionando permissão API Gateway → Lambda..."
	@aws lambda add-permission \
		--function-name $(LAMBDA_FUNCTION_NAME) \
		--statement-id apigw-invoke \
		--action lambda:InvokeFunction \
		--principal apigateway.amazonaws.com \
		--source-arn "arn:aws:execute-api:$(AWS_REGION):$(AWS_ACCOUNT_ID):$(API_GATEWAY_ID)/*/*" \
		--region $(AWS_REGION) > /dev/null 2>&1 \
		|| echo "→ Permissão já existe, pulando."

deploy-api:
	@echo "Criando deployment no API Gateway (stage: $(API_GATEWAY_STAGE))..."
	@aws apigateway create-deployment \
		--rest-api-id $(API_GATEWAY_ID) \
		--stage-name $(API_GATEWAY_STAGE) \
		--region $(AWS_REGION) > /dev/null
	@echo "Deploy concluído!"
	@echo "→ https://api.kodify.com.br/api/v1/health"

deploy: deploy-lambda grant-apigw deploy-api

.DEFAULT_GOAL := help
