# Plano de Implementação — Bot Telegram

## Visão Geral

O bot Telegram é o principal ponto de entrada para o usuário registrar gastos e investimentos no dia a dia. O frontend (dashboard) virá depois como visualização. O bot é o core.

A Lambda `cmd/telegram/main.go` já recebe webhooks do Telegram via API Gateway. A validação do header `X-Telegram-Bot-Api-Secret-Token` já funciona. Falta implementar a lógica dos comandos e a vinculação de identidade.

---

## 1. Problema de Identidade: Telegram ↔ Banco de Dados

O Telegram fornece um `chat_id` (número único por conversa). Nosso banco tem `users` com `email/password`. Precisamos vincular os dois de forma segura.

### Estratégia: Token temporário de vinculação

```
[Frontend / API REST]                         [Telegram]
       │                                           │
       │  1. Cadastro + Login                      │
       │  POST /api/v1/auth/register               │
       │  POST /api/v1/auth/login → JWT            │
       │                                           │
       │  2. Gerar token de vinculação             │
       │  POST /api/v1/telegram/link (com JWT)     │
       │  → retorna token UUID (TTL 5 min)         │
       │  → "Envie /start xK9mQ2... no Telegram"  │
       │                               ────────────┤
       │                                           │  3. Usuário envia:
       │                                           │     /start xK9mQ2...
       │                                           │
       │                                           │  Bot valida token
       │                                           │  → descobre user_id
       │                                           │  → UPDATE users SET telegram_chat_id = chat_id
       │                                           │  → "Conta vinculada! Use /ajuda"
       │                                           │
       │                                           │  4. A partir de agora:
       │                                           │     /gasto 42.50 alimentacao Pizza
       │                                           │     → bot identifica user por chat_id
       │                                           │     → registra no banco
       │                                           │
       │  5. Dashboard (futuro)                    │
       │  GET /api/v1/expenses?month=2026-03       │
       │  → mesmos dados registrados pelo bot      │
```

### Por que essa abordagem?

- **Segura** — não pede email/senha pelo Telegram (ficaria no histórico do chat)
- **Token expira rápido** — se vazar, não serve pra nada em 5 minutos
- **Compatível com frontend futuro** — o dashboard vai ter um botão "Vincular Telegram" que gera o token
- **Infraestrutura já existe** — coluna `telegram_chat_id BIGINT UNIQUE` na tabela `users`, queries `UpdateUserTelegramChatID` e `GetUserByTelegramChatID` já geradas pelo SQLC

---

## 2. O que precisa ser criado

### Fase 1 — Infraestrutura base

| Item | Descrição |
|------|-----------|
| `internal/services/telegram_service.go` | Função `SendMessage(chatID, text, parseMode)` — faz POST na API do Telegram (`https://api.telegram.org/bot<TOKEN>/sendMessage`). Todos os handlers usam isso pra responder. |
| Migration `007_telegram_link_tokens.sql` | Nova tabela para tokens temporários de vinculação. |
| `internal/pgstore/queries/link_token.sql` | Queries SQLC: `InsertLinkToken`, `GetValidLinkToken`, `MarkLinkTokenUsed`. |
| `make sqlc-generate` | Regerar código Go após novas queries. |

#### Schema da migration 007

```sql
CREATE TABLE telegram_link_tokens (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(36) UNIQUE NOT NULL,  -- UUID v4
    expires_at TIMESTAMPTZ NOT NULL,          -- created_at + 5 min
    used_at    TIMESTAMPTZ,                   -- NULL = não usado
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_link_tokens_token ON telegram_link_tokens(token);
```

#### Queries SQLC para link_token.sql

```sql
-- name: InsertLinkToken :one
INSERT INTO telegram_link_tokens (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING id;

-- name: GetValidLinkToken :one
SELECT * FROM telegram_link_tokens
WHERE token = $1
  AND expires_at > NOW()
  AND used_at IS NULL;

-- name: MarkLinkTokenUsed :exec
UPDATE telegram_link_tokens
SET used_at = NOW()
WHERE id = $1;
```

---

### Fase 2 — Vinculação (REST API + Bot)

| Item | Descrição |
|------|-----------|
| `POST /api/v1/telegram/link` | Controller na API REST. Requer JWT. Gera UUID, salva na tabela `telegram_link_tokens` com TTL de 5 min, retorna o token e instruções. |
| `/start <token>` no bot | Valida token → busca user_id → `UpdateUserTelegramChatID(user_id, chat_id)` → marca token como usado → responde com boas-vindas. |
| `/start` sem token | Responde com instruções: "Para vincular sua conta, gere um token no app." |
| `/desvincular` | Seta `telegram_chat_id = NULL` para o user. Útil se trocar de conta. |

#### Resposta do `POST /api/v1/telegram/link`

```json
{
  "token": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "expires_in": 300,
  "instructions": "Envie a mensagem abaixo para @KodifyBot no Telegram:",
  "command": "/start a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

---

### Fase 3 — Comandos do bot (core)

Todos os comandos seguem o mesmo padrão:
1. `GetUserByTelegramChatID(chat_id)` — se não encontrar, responde "Vincule sua conta primeiro com /start"
2. Executa a lógica de negócio
3. `SendMessage(chat_id, resposta formatada em Markdown)`

| Comando | Exemplo de uso | O que faz |
|---------|---------------|-----------|
| `/gasto <valor> <categoria> [descrição]` | `/gasto 35.90 alimentacao Almoço` | `InsertExpense` com `ToCentavos(35.90)` |
| `/investimento <valor> <tipo> [descrição]` | `/investimento 500 cdb CDB Banco Inter` | `InsertInvestment` com `ToCentavos(500)` |
| `/resumo [YYYY-MM]` | `/resumo` ou `/resumo 2026-02` | Total de gastos e investimentos do mês, agrupado por categoria/tipo |
| `/extrato [YYYY-MM]` | `/extrato 2026-03` | Lista os últimos lançamentos do mês |
| `/categorias` | `/categorias` | `ListExpenseCategories` formatado |
| `/tipos` | `/tipos` | `ListInvestmentTypes` formatado |
| `/ajuda` | `/ajuda` | Lista todos os comandos com exemplos |

#### Exemplos de resposta do bot

**`/gasto 35.90 alimentacao Almoço`**
```
✅ Gasto registrado!

💰 R$ 35,90
📁 Alimentação
📝 Almoço
📅 09/03/2026
```

**`/resumo`**
```
📊 Resumo de Março/2026

💸 Gastos: R$ 1.245,00
  • Alimentação: R$ 580,00
  • Transporte: R$ 320,00
  • Lazer: R$ 345,00

📈 Investimentos: R$ 2.000,00
  • CDB: R$ 1.500,00
  • Tesouro: R$ 500,00

💵 Saldo: -R$ 1.245,00 gastos / +R$ 2.000,00 investidos
```

**`/ajuda`**
```
🤖 Comandos disponíveis:

/gasto <valor> <categoria> [descrição]
  Ex: /gasto 35.90 alimentacao Almoço

/investimento <valor> <tipo> [descrição]
  Ex: /investimento 500 cdb CDB Inter

/resumo [YYYY-MM]
  Ex: /resumo ou /resumo 2026-02

/extrato [YYYY-MM]
  Ex: /extrato 2026-03

/categorias — lista categorias de gasto
/tipos — lista tipos de investimento
/desvincular — desvincula sua conta

Categorias: alimentacao, transporte, saude, lazer, moradia, educacao, outros
Tipos: cdb, cotas, acoes, tesouro, cripto, poupanca
```

---

### Fase 4 — Setup do webhook na AWS

| Passo | Como |
|-------|------|
| 1. Criar o bot | Falar com `@BotFather` no Telegram → `/newbot` → salvar o token retornado |
| 2. Configurar variáveis de ambiente na Lambda | `TELEGRAM_BOT_TOKEN=<token do BotFather>` e `TELEGRAM_WEBHOOK_SECRET=<segredo que você escolher>` |
| 3. Registrar webhook no Telegram | Ver comando abaixo |
| 4. Testar | Mandar `/start` pro bot e verificar logs no CloudWatch |

#### Comando para registrar o webhook

```bash
curl -X POST "https://api.telegram.org/bot<TELEGRAM_BOT_TOKEN>/setWebhook" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://<API_GATEWAY_ID>.execute-api.us-east-1.amazonaws.com/<STAGE>/telegram",
    "secret_token": "<TELEGRAM_WEBHOOK_SECRET>",
    "allowed_updates": ["message"]
  }'
```

#### Verificar webhook registrado

```bash
curl "https://api.telegram.org/bot<TELEGRAM_BOT_TOKEN>/getWebhookInfo"
```

---

## 3. Arquitetura dos arquivos

### Arquivos novos a criar

```
internal/
  services/
    telegram_service.go      # SendMessage — HTTP client para API do Telegram
    link_service.go          # GenerateLinkToken, ValidateLinkToken
    expense_service.go       # CreateExpense, ListExpenses, GetMonthSummary
    investment_service.go    # CreateInvestment, ListInvestments
  bot/
    handler.go               # (já existe) — implementar os switch/cases
    formatter.go             # (novo) formata respostas em Markdown para Telegram
  controllers/
    telegram_controller.go   # POST /api/v1/telegram/link
  routes/
    telegram_routes.go       # registra rota com AuthMiddleware
  pgstore/
    queries/
      link_token.sql         # queries SQLC para telegram_link_tokens
    migrations/
      007_telegram_link_tokens.sql
```

### Arquivos existentes a modificar

```
internal/
  api/
    api.go                   # adicionar TelegramController + registrar rotas
    setup.go                 # injetar services e controllers novos
  bot/
    handler.go               # implementar cada case do switch com chamadas aos services
```

---

## 4. Dependências entre os services

```
bot/handler.go
  ├── services/telegram_service.go   ← SendMessage (responder ao user)
  ├── services/link_service.go       ← ValidateLinkToken (/start <token>)
  ├── services/expense_service.go    ← CreateExpense, ListExpenses, GetSummary
  └── services/investment_service.go ← CreateInvestment, ListInvestments

controllers/telegram_controller.go
  └── services/link_service.go       ← GenerateLinkToken

Todos os services usam:
  └── pgstore.Queries                ← SQLC-generated, recebe *pgxpool.Pool
```

---

## 5. Ordem de implementação

| Etapa | O que | Pré-requisito |
|-------|-------|---------------|
| 1 | Migration 007 + queries SQLC + `sqlc-generate` | — |
| 2 | `telegram_service.go` — `SendMessage` | `TELEGRAM_BOT_TOKEN` na env |
| 3 | `/ajuda` e `/start` sem token | Etapa 2 |
| 4 | `link_service.go` + `telegram_controller.go` + rota REST | Etapa 1, auth na API (JWT) |
| 5 | `/start <token>` — vinculação completa | Etapas 2 + 4 |
| 6 | `expense_service.go` + `/gasto` | Etapas 2 + 5 |
| 7 | `investment_service.go` + `/investimento` | Etapas 2 + 5 |
| 8 | `/categorias` + `/tipos` | Etapa 2 |
| 9 | `/resumo` + `/extrato` | Etapas 6 + 7 |
| 10 | `formatter.go` — polir mensagens Markdown | Etapa 3+ |

---

## 6. Decisões em aberto

| # | Questão | Opções | Recomendação |
|---|---------|--------|--------------|
| 1 | **Auth na API REST primeiro?** | (a) Implementar auth completo antes (register/login/JWT) (b) Começar só pelo bot com user de teste | (a) — as rotas protegidas são pré-requisito para `/telegram/link` |
| 2 | **TTL do token de vinculação** | 5 min, 10 min, 30 min | 5 min — segurança > conveniência, gerar outro é trivial |
| 3 | **Um chat_id por user?** | (a) Manter `UNIQUE` — 1:1 (b) Tabela N:N para múltiplos chats | (a) — KISS, caso de uso é pessoal |
| 4 | **Formato das mensagens** | Markdown, HTML | Markdown — mais natural no Telegram |
| 5 | **`/desvincular` precisa de confirmação?** | (a) Executa direto (b) Pede "Tem certeza? envie /desvincular confirmar" | (b) — evita acidente |
