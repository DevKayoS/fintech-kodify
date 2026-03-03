# Fintech Kodify — Roadmap de Desenvolvimento

## Visão Geral

API serverless de controle financeiro pessoal, deployada como AWS Lambda, com dois pontos de entrada:
- **API REST** via API Gateway (uso direto ou integrações futuras)
- **Bot Telegram** via webhook Lambda (interface conversacional)

Usuários registram gastos e investimentos através de comandos simples no Telegram ou chamadas HTTP diretas.

---

## Stack

| Camada | Tecnologia |
|--------|-----------|
| Linguagem | Go 1.25+ |
| HTTP Framework | Gin |
| Lambda Adapter | aws-lambda-go-api-proxy |
| Banco de Dados | PostgreSQL (RDS ou Aurora Serverless) |
| Query Layer | SQLC + pgx/v5 |
| Migrations | Tern |
| Autenticação | JWT (HMAC-SHA256) |
| Deploy | AWS Lambda + API Gateway |
| Bot | Telegram Bot API (webhook) |
| Secrets | AWS SSM Parameter Store |

> Base estrutural espelhada do repositório `go-lambda` com os mesmos padrões de erros, middleware, DI e SQLC.

---

## Entidades do Domínio

### Users
Usuários do sistema. Cada usuário tem um `telegram_chat_id` opcional para vincular a conta ao bot.

### Expenses (Gastos)
Registro de saídas financeiras. Possui categoria, descrição e valor.

### Investments (Investimentos)
Registro de aportes financeiros por tipo (CDB, Fundos de Investimento/Cotas, Ações, Tesouro Direto, Cripto, etc).

### Categories (Categorias de Gasto)
Tabela auxiliar para classificar gastos (Alimentação, Transporte, Saúde, Lazer, etc). Futuramente customizável por usuário.

### Investment Types (Tipos de Investimento)
Tabela auxiliar com os tipos de investimento disponíveis (CDB, Cotas/FI, Ações, Tesouro Direto, Cripto, Poupança, etc).

---

## Comandos do Bot Telegram

### Registro

```
/gasto 39
/gasto 39 alimentacao
/gasto 39 alimentacao Almoço no restaurante

/investimento 1600 cdb
/investimento 1600 cdb Banco Inter
/investimento 800 cotas XP Multimercado
```

### Consulta

```
/resumo                  → resumo do mês atual (total gastos + total investido)
/resumo 2026-02          → resumo de um mês específico
/extrato                 → últimas 10 transações (gastos + investimentos)
/extrato gastos          → últimos 10 gastos
/extrato investimentos   → últimos 10 investimentos
/categorias              → lista categorias disponíveis de gasto
/tipos_investimento      → lista tipos de investimento disponíveis
```

### Conta

```
/start                   → vincular telegram_chat_id à conta (precisa de token gerado via API)
/ajuda                   → lista de comandos disponíveis
```

---

## Arquitetura de Lambdas

Dois Lambdas independentes:

### Lambda 1 — API REST
- Trigger: API Gateway HTTP
- Responsável por: autenticação, CRUD de usuários, registro e consulta de gastos/investimentos
- Padrão: mesmo do `go-lambda` (Gin + ginadapter)

### Lambda 2 — Telegram Webhook
- Trigger: API Gateway POST `/webhook/telegram`
- Responsável por: receber updates do Telegram, parsear comandos, chamar a lógica de negócio e responder
- Reutiliza os mesmos services do Lambda 1 (shared library)
- Sem Gin — handler Lambda puro

> Alternativa de arquitetura: um único Lambda com duas rotas (uma para API, uma para webhook). Mais simples inicialmente, avaliar conforme crescimento.

---

## Estrutura de Diretórios

```
fintech-kodify/
├── cmd/
│   ├── api/
│   │   └── main.go              # Entry point Lambda API REST
│   └── telegram/
│       └── main.go              # Entry point Lambda Telegram Webhook
├── internal/
│   ├── api/
│   │   ├── api.go               # Struct API + BindRoutes
│   │   └── setup.go             # DI + gin.Engine
│   ├── bot/
│   │   ├── handler.go           # Lambda handler do webhook Telegram
│   │   ├── parser.go            # Parse de comandos (/gasto, /investimento, etc)
│   │   └── responder.go         # Envio de mensagens ao Telegram
│   ├── controllers/
│   │   ├── health_controller.go
│   │   ├── token_controller.go
│   │   ├── user_controller.go
│   │   ├── expense_controller.go
│   │   └── investment_controller.go
│   ├── errors/
│   │   ├── api_error.go         # (espelhado do go-lambda)
│   │   └── factory.go
│   ├── middleware/
│   │   ├── auth_middleware.go
│   │   └── error_middleware.go
│   ├── models/
│   │   ├── token.go
│   │   ├── expense.go
│   │   ├── investment.go
│   │   └── telegram.go          # Structs de update do Telegram
│   ├── pgstore/
│   │   ├── database/
│   │   │   └── db.go            # Pool pgxpool
│   │   ├── migrations/
│   │   │   ├── 001_create_users.sql
│   │   │   ├── 002_roles.sql
│   │   │   ├── 003_investment_types.sql
│   │   │   ├── 004_expense_categories.sql
│   │   │   ├── 005_expenses.sql
│   │   │   └── 006_investments.sql
│   │   ├── queries/
│   │   │   ├── user.sql
│   │   │   ├── expense.sql
│   │   │   └── investment.sql
│   │   └── sqlc.yaml
│   ├── routes/
│   │   ├── health_routes.go
│   │   ├── token_routes.go
│   │   ├── user_routes.go
│   │   ├── expense_routes.go
│   │   └── investment_routes.go
│   ├── services/
│   │   ├── token/
│   │   │   └── token_service.go
│   │   ├── user/
│   │   │   ├── user_service.go
│   │   │   └── user_service_test.go
│   │   ├── expense/
│   │   │   ├── expense_service.go
│   │   │   └── expense_service_test.go
│   │   └── investment/
│   │       ├── investment_service.go
│   │       └── investment_service_test.go
│   └── utils/
│       ├── hashed_password.go
│       └── date.go              # Helpers de parse de data/período
├── go.mod
├── go.sum
├── Makefile
├── sqlc.yaml
└── ROADMAP.md
```

---

## Schema do Banco de Dados

### Migration 001 — users
```sql
CREATE TABLE users (
    id           BIGSERIAL PRIMARY KEY,
    name         VARCHAR(255) NOT NULL,
    email        VARCHAR(255) UNIQUE NOT NULL,
    password     VARCHAR(255) NOT NULL,
    telegram_chat_id BIGINT UNIQUE,          -- vínculo com bot Telegram
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW()
);
```

### Migration 002 — roles e permissions
> Igual ao go-lambda (admin, user, moderator)

### Migration 003 — investment_types
```sql
CREATE TABLE investment_types (
    id          SERIAL PRIMARY KEY,
    slug        VARCHAR(50) UNIQUE NOT NULL,  -- ex: cdb, cotas, acoes
    name        VARCHAR(100) NOT NULL,        -- ex: CDB, Fundos de Investimento
    description TEXT
);

-- Seeds iniciais
INSERT INTO investment_types (slug, name) VALUES
    ('cdb',      'CDB'),
    ('cotas',    'Fundos de Investimento'),
    ('acoes',    'Ações'),
    ('tesouro',  'Tesouro Direto'),
    ('cripto',   'Criptomoedas'),
    ('poupanca', 'Poupança');
```

### Migration 004 — expense_categories
```sql
CREATE TABLE expense_categories (
    id          SERIAL PRIMARY KEY,
    slug        VARCHAR(50) UNIQUE NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description TEXT
);

-- Seeds iniciais
INSERT INTO expense_categories (slug, name) VALUES
    ('alimentacao', 'Alimentação'),
    ('transporte',  'Transporte'),
    ('saude',       'Saúde'),
    ('lazer',       'Lazer'),
    ('moradia',     'Moradia'),
    ('educacao',    'Educação'),
    ('outros',      'Outros');
```

### Migration 005 — expenses
```sql
CREATE TABLE expenses (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    category_id INT REFERENCES expense_categories(id),
    amount      BIGINT NOT NULL,             -- em centavos
    description VARCHAR(500),
    occurred_at TIMESTAMPTZ DEFAULT NOW(),   -- data do gasto (pode ser retroativa)
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_expenses_user_id ON expenses(user_id);
CREATE INDEX idx_expenses_occurred_at ON expenses(occurred_at);
```

### Migration 006 — investments
```sql
CREATE TABLE investments (
    id                  BIGSERIAL PRIMARY KEY,
    user_id             BIGINT NOT NULL REFERENCES users(id),
    investment_type_id  INT NOT NULL REFERENCES investment_types(id),
    amount              BIGINT NOT NULL,     -- em centavos
    description         VARCHAR(500),        -- ex: nome do fundo, banco
    invested_at         TIMESTAMPTZ DEFAULT NOW(),
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_investments_user_id ON investments(user_id);
CREATE INDEX idx_investments_invested_at ON investments(invested_at);
```

> **Nota sobre valores:** todos os valores monetários são armazenados em **centavos** (inteiro) para evitar problemas de ponto flutuante. A API recebe e devolve em reais (float/string), a conversão ocorre na service layer.

---

## Endpoints da API REST

### Públicos (sem auth)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/api/v1/health` | Healthcheck |
| POST | `/api/v1/auth` | Gerar token JWT |
| POST | `/api/v1/users` | Criar usuário |
| POST | `/api/v1/webhook/telegram` | Webhook do bot Telegram |

### Protegidos (Bearer JWT)

**Gastos**

| Método | Rota | Descrição |
|--------|------|-----------|
| POST | `/api/v1/expenses` | Registrar gasto |
| GET | `/api/v1/expenses` | Listar gastos (filtros: mês, categoria) |
| GET | `/api/v1/expenses/:id` | Detalhe de um gasto |
| DELETE | `/api/v1/expenses/:id` | Remover gasto |

**Investimentos**

| Método | Rota | Descrição |
|--------|------|-----------|
| POST | `/api/v1/investments` | Registrar investimento |
| GET | `/api/v1/investments` | Listar investimentos (filtros: mês, tipo) |
| GET | `/api/v1/investments/:id` | Detalhe de um investimento |
| DELETE | `/api/v1/investments/:id` | Remover investimento |

**Resumo**

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/api/v1/summary` | Resumo do período (total gastos + investimentos, breakdown por categoria/tipo) |
| GET | `/api/v1/summary?month=2026-02` | Resumo de mês específico |

**Auxiliares**

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/api/v1/expense-categories` | Listar categorias de gasto |
| GET | `/api/v1/investment-types` | Listar tipos de investimento |
| GET | `/api/v1/users/me` | Perfil do usuário autenticado |
| POST | `/api/v1/users/link-telegram` | Vincular telegram_chat_id via token gerado |

---

## Fluxo do Bot Telegram

### Registro de conta (onboarding)
1. Usuário cria conta via API REST (ou admin cria para ele)
2. API gera um token temporário de vínculo
3. Usuário envia `/start <token>` no bot
4. Bot registra o `telegram_chat_id` no `user_id` correspondente
5. A partir daí, todas as mensagens do chat_id são associadas ao usuário

### Fluxo de comando
```
Telegram → POST /webhook/telegram → Lambda Telegram
→ parser.go identifica comando e argumentos
→ chama service correspondente (expense_service ou investment_service)
→ persiste no banco
→ responder.go envia mensagem de confirmação ao usuário
```

### Tratamento de erros no bot
- Comando não reconhecido → mensagem de ajuda
- Valor inválido → mensagem explicativa
- Tipo/categoria não encontrado → lista as opções disponíveis
- Usuário não vinculado → instrução de como fazer o `/start`

---

## Fases de Desenvolvimento

### Fase 1 — Fundação
- [ ] Configurar `go.mod` com todas as dependências
- [ ] Copiar e adaptar camada de erros (`internal/errors/`)
- [ ] Copiar e adaptar middleware (`internal/middleware/`)
- [ ] Configurar `internal/pgstore/database/db.go` (pool pgxpool)
- [ ] Escrever migrations 001 e 002 (users + roles) — pode reaproveitar do go-lambda
- [ ] Configurar `sqlc.yaml` e escrever queries base de usuários
- [ ] Gerar código SQLC
- [ ] Implementar `token_service` e `user_service`
- [ ] Implementar endpoints públicos: health, auth, create user
- [ ] `cmd/api/main.go` funcional como Lambda
- [ ] Makefile com build, zip, deploy, migrate, sqlc-generate

### Fase 2 — Core Financeiro
- [ ] Migrations 003-006 (investment_types, expense_categories, expenses, investments)
- [ ] Seeds para tipos e categorias
- [ ] Queries SQLC para expenses e investments
- [ ] `expense_service` com interface + testes
- [ ] `investment_service` com interface + testes
- [ ] Controllers e routes de expenses e investments
- [ ] Endpoint `/api/v1/summary`
- [ ] Endpoints auxiliares (categorias, tipos)

### Fase 3 — Bot Telegram
- [ ] Definir structs de Update do Telegram em `internal/models/telegram.go`
- [ ] Implementar `internal/bot/parser.go` para parsear comandos
- [ ] Implementar `internal/bot/responder.go` para enviar mensagens via Telegram API
- [ ] Implementar `internal/bot/handler.go` (Lambda handler)
- [ ] Fluxo de vínculo de conta (`/start <token>`)
- [ ] Comandos de registro: `/gasto`, `/investimento`
- [ ] Comandos de consulta: `/resumo`, `/extrato`, `/categorias`, `/tipos_investimento`
- [ ] `cmd/telegram/main.go` funcional como Lambda
- [ ] Configurar webhook no Telegram

### Fase 4 — Qualidade e Infra
- [ ] Testes de integração básicos
- [ ] Logging estruturado (slog) em todas as camadas
- [ ] Configuração de CORS no API Gateway
- [ ] Variáveis de ambiente via SSM Parameter Store
- [ ] Documentação dos endpoints (comentários ou Swagger básico)
- [ ] CI/CD básico (GitHub Actions para build e deploy)

---

## Variáveis de Ambiente

| Variável | Descrição | Obrigatória |
|----------|-----------|-------------|
| `DATABASE_URL` | Connection string PostgreSQL | Sim |
| `JWT_SECRET_KEY` | Chave secreta para assinar tokens | Sim |
| `TELEGRAM_BOT_TOKEN` | Token do bot (BotFather) | Sim (Lambda Telegram) |
| `TELEGRAM_WEBHOOK_SECRET` | Header secret para validar chamadas do Telegram | Recomendado |
| `APP_ENV` | `production` ou `development` | Não |

---

## Decisões de Design

- **Valores em centavos:** evita float, sem perda de precisão, padrão em fintechs.
- **Dois Lambdas separados:** API e webhook independentes facilitam logs, scaling e permissões IAM distintas.
- **SQLC:** type-safe, sem ORM pesado, queries explícitas e rastreáveis — mesmo padrão do go-lambda.
- **Tern para migrations:** mesmo padrão do go-lambda, simples e sem dependências extras no binário.
- **Slugs para categorias/tipos:** facilita o parse de comandos do Telegram sem necessidade de IDs numéricos.
- **telegram_chat_id UNIQUE em users:** um chat_id por conta, evita duplicidade de vínculo.
- **occurred_at / invested_at separado de created_at:** permite registrar transações retroativas ("gastei ontem R$ 50").
