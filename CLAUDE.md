# fintech-kodify — Guia do Projeto

API serverless de controle financeiro pessoal com dois entry points: REST API (Gin + Lambda) e Bot Telegram (webhook Lambda).

**Module:** `github.com/DevKayoS/fintech-kodify`
**Go:** 1.25.5
**Arquitetura:** DDD (Domain-Driven Design)

---

## Estrutura de Diretórios

```
fintech-kodify/
├── cmd/
│   ├── api/main.go                   # Entry point Lambda — REST API
│   └── telegram/main.go              # Entry point Lambda — Bot Telegram
├── internal/
│   ├── domain/                       # Entidades e contratos de domínio (regras de negócio puras)
│   │   ├── expense/
│   │   │   └── entity.go             # Expense entity + Repository interface
│   │   ├── investment/
│   │   │   └── entity.go             # Investment entity + Repository interface
│   │   └── user/
│   │       └── entity.go             # User entity + Repository interface
│   │
│   ├── usecases/                     # Casos de uso (orquestração de domínio + infra)
│   │   └── health/
│   │       └── health_usecase.go     # HealthUseCase — GetStatus(), AppStatus, DBStatus
│   │
│   ├── adapters/                     # Adaptadores externos (HTTP, Telegram)
│   │   ├── rest/                     # Adaptador HTTP (REST API)
│   │   │   ├── api.go                # Struct API + BindRoutes (registra todas as rotas)
│   │   │   ├── setup.go              # SetupAPI() — cria gin.Engine e injeta dependências
│   │   │   ├── controllers/
│   │   │   │   └── health_controller.go  # GET /api/v1/status
│   │   │   └── routes/
│   │   │       └── health_routes.go      # SetupHealthRoutes
│   │   └── telegram/                 # Adaptador Telegram (bot webhook)
│   │       ├── handler.go            # HandleUpdate — processa webhooks do Telegram
│   │       ├── commands.go           # helpMessage()
│   │       ├── responder.go          # sendMessage()
│   │       └── models.go             # TelegramUpdate, TelegramMessage, TelegramUser, etc.
│   │
│   ├── infrastructure/               # Implementações de infraestrutura
│   │   └── pgstore/                  # SQLC — acesso ao banco de dados
│   │       ├── database/db.go        # Init(ctx) — pgxpool, Pool global, MaxConns=2
│   │       ├── db.go                 # SQLC: interface Querier + struct Queries
│   │       ├── models.go             # SQLC: structs geradas (User, Expense, Investment, etc.)
│   │       ├── user.sql.go           # SQLC: queries de usuário
│   │       ├── expense.sql.go        # SQLC: queries de despesa
│   │       ├── investment.sql.go     # SQLC: queries de investimento
│   │       ├── link_token.sql.go     # SQLC: queries de link token (Telegram)
│   │       ├── postgres.sql.go       # SQLC: query de status do DB
│   │       ├── queries/              # SQL fonte (input do SQLC)
│   │       │   ├── user.sql
│   │       │   ├── expense.sql
│   │       │   ├── investment.sql
│   │       │   ├── link_token.sql
│   │       │   └── postgres.sql
│   │       ├── migrations/           # Migrations Tern
│   │       │   ├── 001_create_users.sql
│   │       │   ├── 002_roles.sql
│   │       │   ├── 003_investment_types.sql
│   │       │   ├── 004_expense_categories.sql
│   │       │   ├── 005_expenses.sql
│   │       │   ├── 006_investments.sql
│   │       │   └── 007_telegram_link_tokens.sql
│   │       ├── schema_extras.sql
│   │       └── sqlc.yaml
│   │
│   ├── errors/
│   │   ├── api_error.go              # Struct ApiError {Code, Message, StatusCode, Err}
│   │   └── factory.go                # BadRequest, Unauthorized, Forbidden, NotFound, Conflict, UnprocessableEntity, Internal
│   ├── middleware/
│   │   ├── auth_middleware.go        # AuthMiddleware(), RequireRole(), RequirePermissions()
│   │   └── error_middleware.go       # ErrorHandler() — captura ApiError e formata JSON
│   ├── models/                       # DTOs HTTP (request/response structs)
│   │   ├── expense.go                # CreateExpenseRequest, ExpenseResponse, ListExpensesQuery
│   │   ├── investment.go             # CreateInvestmentRequest, InvestmentResponse, ListInvestmentsQuery
│   │   └── token.go                  # GenerateTokenRequest, SecretKey (lido de JWT_SECRET_KEY)
│   └── utils/
│       ├── date.go                   # ParseMonthRange, CurrentMonthRange
│       ├── money.go                  # ToReais, ToCentavos
│       └── hashed_password.go        # HashPassword, CheckPasswordHash (bcrypt)
├── infra/                            # Terraform (infraestrutura AWS)
├── .env.example
├── docker-compose.yml
├── go.mod
├── Makefile
└── ROADMAP.md
```

---

## Arquitetura DDD — Camadas e Responsabilidades

```
┌─────────────────────────────────────────────────────────┐
│  adapters/rest   adapters/telegram                      │  ← entrada/saída externa
│  (HTTP, Bot)                                            │
├─────────────────────────────────────────────────────────┤
│  usecases/                                              │  ← orquestração
│  (casos de uso por domínio)                             │
├─────────────────────────────────────────────────────────┤
│  domain/                                                │  ← regras de negócio puras
│  (entities + repository interfaces)                     │
├─────────────────────────────────────────────────────────┤
│  infrastructure/pgstore                                 │  ← implementação de persistência
│  (SQLC + pgxpool)                                       │
└─────────────────────────────────────────────────────────┘
```

**Regra de dependência:** cada camada só depende das camadas abaixo. Adapters dependem de usecases; usecases dependem de domain; infrastructure implementa interfaces de domain.

---

## Fluxo de uma Requisição REST

```
API Gateway
    ↓
cmd/api/main.go → Handler(ctx, req)
    ↓
ginadapter (converte APIGatewayProxyRequest → http.Request)
    ↓
gin.Engine
    ↓
middleware.ErrorHandler()          ← captura panics e ApiError
    ↓
middleware.AuthMiddleware()         ← valida JWT, seta user_id/email/role/permissions no ctx
    ↓
adapters/rest/controllers/*.go
    ↓
usecases/<dominio>/<usecase>.go
    ↓
infrastructure/pgstore (SQLC)
    ↓
PostgreSQL (Neon em prod / Docker em dev)
```

---

## Fluxo do Bot Telegram

```
Telegram → API Gateway (webhook)
    ↓
cmd/telegram/main.go → Handler(ctx, req)
    ↓
adapters/telegram.HandleUpdate(ctx, req)
    ↓
validateWebhookSecret()  ← header X-Telegram-Bot-Api-Secret-Token
    ↓
parseCommand(text)       ← extrai "/comando arg1 arg2"
    ↓
switch command:
  /start         → vincula telegram_chat_id ao usuário
  /gasto         → cria despesa
  /investimento  → cria investimento
  /resumo        → resumo mensal
  /extrato       → histórico de transações
  /categorias    → lista categorias de despesa
  /tipos_investimento → lista tipos de investimento
  /ajuda         → mensagem de ajuda
    ↓
sempre retorna 200 OK (evita retry do Telegram)
```

---

## Banco de Dados

### Conexão (`internal/infrastructure/pgstore/database/db.go`)
- Singleton `Pool *pgxpool.Pool` inicializado via `Init(ctx)`
- Lê `DATABASE_URL` do ambiente
- MaxConns=2, MinConns=1 — otimizado para Lambda (cold start + Neon)
- MaxConnLifetime=5min, MaxConnIdleTime=1min

### Schema (7 migrations — formato Tern)

| Tabela | Descrição |
|--------|-----------|
| `users` | id, name, email, password, telegram_chat_id, role_id |
| `roles` | admin, user, moderator |
| `permissions` | read/write/delete:expenses, investments, users; manage:all |
| `role_permissions` | N:N entre roles e permissions |
| `investment_types` | cdb, cotas, acoes, tesouro, cripto, poupanca |
| `expense_categories` | alimentacao, transporte, saude, lazer, moradia, educacao, outros |
| `expenses` | user_id, category_id, amount (centavos), description, occurred_at |
| `investments` | user_id, investment_type_id, amount (centavos), description, invested_at |
| `telegram_link_tokens` | user_id, token, expires_at, used_at |

**Regra importante:** valores monetários são armazenados em **centavos** (`int64`). Use `utils.ToReais` / `utils.ToCentavos` para converter.

### Queries SQLC disponíveis (`internal/infrastructure/pgstore/`)

**Usuários:**
- `GetUserByEmail`, `GetUserByID`, `GetUserByTelegramChatID`
- `InsertUser`, `UpdateUserTelegramChatID`, `ClearUserTelegramChatID`
- `GetUserWithRole`, `GetUserPermissions`, `GetRoleByName`

**Despesas:**
- `InsertExpense`, `GetExpenseByID`, `DeleteExpense`
- `ListExpensesByUser`, `ListExpensesByUserAndPeriod`
- `GetExpenseCategoryBySlug`, `ListExpenseCategories`
- `GetExpenseSummaryByPeriod`

**Investimentos:**
- `InsertInvestment`, `GetInvestmentByID`, `DeleteInvestment`
- `ListInvestmentsByUser`, `ListInvestmentsByUserAndPeriod`
- `GetInvestmentTypeBySlug`, `ListInvestmentTypes`
- `GetInvestmentSummaryByPeriod`

**Telegram Link Tokens:**
- `InsertLinkToken`, `GetValidLinkToken`, `MarkLinkTokenUsed`

---

## Autenticação & Autorização

### JWT (`internal/middleware/auth_middleware.go`)
- Header: `Authorization: Bearer <token>`
- Segredo: env `JWT_SECRET_KEY` (fallback: `"secretKey"` — só dev), lido em `internal/models/token.go`
- Após validação, seta no contexto Gin: `user_id`, `email`, `role`, `permissions`, `claims`

### Middlewares
```go
AuthMiddleware()                         // valida token JWT
RequireRole("admin", "moderator")        // verifica role do usuário
RequirePermissions("write:expenses")     // verifica permissão específica
                                         // "manage:all" é wildcard de admin
```

---

## Tratamento de Erros

### ApiError (`internal/errors/`)
```go
errors.BadRequest("mensagem")           // 400
errors.Unauthorized("mensagem")         // 401
errors.Forbidden("mensagem")            // 403
errors.NotFound("mensagem")             // 404
errors.Conflict("mensagem")             // 409
errors.UnprocessableEntity("mensagem")  // 422
errors.Internal("mensagem", err)        // 500
```

### Formato da resposta de erro
```json
{ "error": { "code": "NOT_FOUND", "msg": "recurso não encontrado" } }
```

O middleware `ErrorHandler()` deve ser o **primeiro** middleware registrado no grupo de rotas.

---

## Modelos de Request/Response (`internal/models/`)

### Expense
```go
CreateExpenseRequest {
    CategorySlug string  `json:"category_slug"`
    Amount       float64 `json:"amount"`      // em reais
    Description  string  `json:"description"`
    OccurredAt   string  `json:"occurred_at"` // RFC3339 ou YYYY-MM-DD, opcional
}

ListExpensesQuery {
    Month    string `form:"month"`    // YYYY-MM
    Category string `form:"category"` // slug
}
```

### Investment
```go
CreateInvestmentRequest {
    TypeSlug    string  `json:"type_slug"`   // obrigatório
    Amount      float64 `json:"amount"`      // em reais
    Description string  `json:"description"`
    InvestedAt  string  `json:"invested_at"` // RFC3339 ou YYYY-MM-DD, opcional
}

ListInvestmentsQuery {
    Month string `form:"month"` // YYYY-MM
    Type  string `form:"type"`  // slug
}
```

---

## Utilitários

### `internal/utils/date.go`
```go
ParseMonthRange("2024-03")   // → (startOfMonth, endOfMonth, error)
```

### `internal/utils/money.go`
```go
CurrentMonthRange()          // → (startOfMonth, endOfMonth)
ToReais(centavos int64)      // → float64 (÷ 100)
ToCentavos(reais float64)    // → int64 (× 100)
```

### `internal/utils/hashed_password.go`
```go
HashPassword(password string)           // → (hash string, error)
CheckPasswordHash(password, hash string) // → bool
```

---

## Rotas Existentes

| Método | Path | Handler | Auth |
|--------|------|---------|------|
| GET | `/api/v1/status` | `HealthController.Check` | Nenhuma |

**Estrutura de prefixos:**
- `/api/v1/` — grupo base
- Rotas públicas: diretamente no grupo v1
- Rotas protegidas: passam por `AuthMiddleware()` (a implementar)

---

## Variáveis de Ambiente

| Variável | Descrição | Exemplo |
|----------|-----------|---------|
| `DATABASE_URL` | Connection string PostgreSQL | `postgres://kodify:kodify@localhost:5432/fintech_kodify` |
| `JWT_SECRET_KEY` | Segredo para assinar JWT | `minha-chave-secreta` |
| `TELEGRAM_BOT_TOKEN` | Token do bot Telegram | `123456:ABC-DEF...` |
| `TELEGRAM_WEBHOOK_SECRET` | Segredo do webhook Telegram | `webhook-secret` |
| `APP_ENV` | Ambiente | `development` / `production` |

---

## Makefile — Targets Principais

```bash
# Dev local
make docker-up          # sobe PostgreSQL (porta 5432)
make docker-down        # derruba PostgreSQL
make run                # roda API local (go run ./cmd/api) — requer SAM ou adaptação local

# Banco
make migrate-up         # aplica todas as migrations (requer: go install github.com/jackc/tern/v2@latest)
make migrate-down       # reverte última migration
make migrate-status     # status das migrations
make migrate-new        # cria nova migration

# Codegen
make sqlc-generate      # gera código Go a partir dos .sql (requer: sqlc instalado)

# Build
make build-api          # compila API para Linux/amd64 (Lambda)
make build-telegram     # compila bot para Linux/amd64
make zip-api            # cria lambda-api.zip (bootstrap)
make zip-telegram       # cria lambda-telegram.zip (bootstrap)

# Deploy AWS
make deploy             # build + zip + deploy Lambda + API Gateway
make deploy-lambda      # só cria/atualiza a função Lambda
```

---

## O que ainda não está implementado

- **Use Cases**: TokenUseCase, UserUseCase, ExpenseUseCase, InvestmentUseCase
- **Controllers**: TokenController, UserController, ExpenseController, InvestmentController
- **Rotas protegidas**: registrar com `AuthMiddleware()` em `adapters/rest/api.go`
- **Bot commands**: implementações reais dos handlers em `adapters/telegram/handler.go`
- **Testes**: nenhum arquivo `_test.go` ainda

---

## Padrão para Novos Use Cases / Controllers

Cada feature segue este fluxo de arquivos:

```
1. internal/models/<feature>.go                           → DTOs HTTP (request/response)
2. internal/domain/<feature>/entity.go                   → entidade de domínio + Repository interface
3. internal/infrastructure/pgstore/queries/<feature>.sql  → SQL queries
   (rodar: make sqlc-generate)                           → pgstore/<feature>.sql.go
4. internal/usecases/<feature>/<action>_usecase.go        → caso de uso
5. internal/adapters/rest/controllers/<feature>_controller.go
6. internal/adapters/rest/routes/<feature>_routes.go
7. internal/adapters/rest/api.go                          → registrar controller + rotas
```

### Padrão de injeção de dependência nos Use Cases

Todo use case define uma interface local para o repositório (facilita testes) e recebe `*pgxpool.Pool` no construtor. Referência: `internal/usecases/health/health_usecase.go`:

```go
// 1. Interface local — declara apenas os métodos que o use case usa
type FeatureRepository interface {
    MetodoDoBanco(ctx context.Context, ...) (pgstore.XxxRow, error)
}

// 2. Struct do use case — campo privado do tipo da interface
type FeatureUseCase struct {
    repository FeatureRepository
}

// 3. Construtor — recebe *pgxpool.Pool e injeta pgstore.New(pool)
func NewFeatureUseCase(pool *pgxpool.Pool) *FeatureUseCase {
    return &FeatureUseCase{
        repository: pgstore.New(pool),
    }
}

// 4. Métodos — chamam uc.repository.<Query>(ctx, ...)
func (uc *FeatureUseCase) FazAlgo(ctx context.Context) (Result, error) {
    row, err := uc.repository.MetodoDoBanco(ctx, ...)
    // ...
}
```

---

## Dependências Go

```
github.com/aws/aws-lambda-go              v1.52.0
github.com/awslabs/aws-lambda-go-api-proxy v0.16.2  # ginadapter
github.com/gin-gonic/gin                  v1.9.1
github.com/golang-jwt/jwt                 v3.2.2
github.com/jackc/pgx/v5                   v5.7.6
golang.org/x/crypto                       v0.37.0  # bcrypt
```
