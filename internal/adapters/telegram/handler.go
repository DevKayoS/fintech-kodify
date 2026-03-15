package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/DevKayoS/fintech-kodify/internal/usecases/expense"
	"github.com/aws/aws-lambda-go/events"
)

type Handler struct {
	expenseUC *expense.ExpenseUseCase
}

func NewHandler(expenseUC *expense.ExpenseUseCase) *Handler {
	return &Handler{expenseUC: expenseUC}
}

func (h *Handler) HandleUpdate(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !validateWebhookSecret(req) {
		slog.Warn("telegram webhook: invalid secret header")
		return events.APIGatewayProxyResponse{StatusCode: http.StatusUnauthorized}, nil
	}

	var update TelegramUpdate
	if err := json.Unmarshal([]byte(req.Body), &update); err != nil {
		slog.Error("telegram webhook: failed to parse update", "error", err)
		return okResponse(), nil // sempre 200 para o Telegram não re-tentar
	}

	if update.Message == nil || update.Message.Text == "" {
		return okResponse(), nil
	}

	msg := update.Message
	chatID := msg.Chat.ID

	command, args := parseCommand(msg.Text)

	slog.Info("telegram command received", "chat_id", chatID, "command", command, "args", args)

	switch command {
	case "/start":
		// TODO: vincular telegram_chat_id ao usuário via token
		// token := args[0]
		// userService.LinkTelegramChatID(ctx, token, chatID)
	case "/gasto":
		h.handleGasto(ctx, chatID, args)
	case "/investimento":
		// TODO: registrar investimento via investment use case
	case "/resumo":
		// TODO: buscar resumo mensal e enviar mensagem
	case "/extrato":
		// TODO: buscar extrato e enviar mensagem
	case "/categorias":
		h.handleCategorias(ctx, chatID)
	case "/tipos_investimento":
		// TODO: listar tipos de investimento
	case "/ajuda":
		sendMessage(chatID, helpMessage())
	default:
		slog.Info("telegram webhook: unknown command", "command", command)
	}

	return okResponse(), nil
}

func (h *Handler) handleGasto(ctx context.Context, chatID int64, args []string) {
	if len(args) < 2 {
		sendMessage(chatID, "Uso: `/gasto <valor> <categoria> <descrição>`\nEx: `/gasto 39.90 alimentacao Almoço`")
		return
	}

	amountStr := strings.ReplaceAll(args[0], ",", ".")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		sendMessage(chatID, "❌ Valor inválido. Use ponto ou vírgula como separador decimal.\nEx: `/gasto 39.90 alimentacao Almoço`")
		return
	}

	categorySlug := args[1]
	description := ""
	if len(args) > 2 {
		description = strings.Join(args[2:], " ")
	}

	result, err := h.expenseUC.CreateFromTelegram(ctx, chatID, amount, categorySlug, description)
	if err != nil {
		sendMessage(chatID, "❌ "+err.Error())
		return
	}

	msg := fmt.Sprintf("✅ *Gasto registrado!*\n\n💸 R$ %.2f\n🏷️ %s", result.Amount, result.CategoryName)
	if result.Description != "" {
		msg += "\n📝 " + result.Description
	}
	sendMessage(chatID, msg)
}

func (h *Handler) handleCategorias(ctx context.Context, chatID int64) {
	categories, err := h.expenseUC.ListCategories(ctx)
	if err != nil {
		sendMessage(chatID, "❌ Erro ao buscar categorias.")
		return
	}

	msg := "*Categorias disponíveis:*\n"
	for _, c := range categories {
		msg += fmt.Sprintf("• `%s` — %s\n", c.Slug, c.Name)
	}
	sendMessage(chatID, msg)
}

// parseCommand separa o comando dos argumentos de uma mensagem Telegram.
// Ex: "/gasto 39 alimentacao Almoço" → ("/gasto", ["39", "alimentacao", "Almoço"])
func parseCommand(text string) (string, []string) {
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return "", nil
	}

	command := parts[0]

	// Remove sufixo @BotName caso presente (ex: /start@MeuBot)
	if idx := strings.Index(command, "@"); idx != -1 {
		command = command[:idx]
	}

	return command, parts[1:]
}

// validateWebhookSecret verifica o header X-Telegram-Bot-Api-Secret-Token.
// Se TELEGRAM_WEBHOOK_SECRET não estiver definida, pula a validação.
func validateWebhookSecret(req events.APIGatewayProxyRequest) bool {
	secret := os.Getenv("TELEGRAM_WEBHOOK_SECRET")
	if secret == "" {
		return true
	}
	return req.Headers["X-Telegram-Bot-Api-Secret-Token"] == secret
}

func okResponse() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}
}
