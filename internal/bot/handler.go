package bot

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/DevKayoS/fintech-kodify/internal/models"
	"github.com/aws/aws-lambda-go/events"
)

func HandleUpdate(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !validateWebhookSecret(req) {
		slog.Warn("telegram webhook: invalid secret header")
		return events.APIGatewayProxyResponse{StatusCode: http.StatusUnauthorized}, nil
	}

	var update models.TelegramUpdate
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
		// TODO: registrar gasto via expense service
		// expenseService.CreateFromTelegram(ctx, chatID, args)
	case "/investimento":
		// TODO: registrar investimento via investment service
		// investmentService.CreateFromTelegram(ctx, chatID, args)
	case "/resumo":
		// TODO: buscar resumo mensal e enviar mensagem
		// summaryService.GetSummary(ctx, chatID, args)
	case "/extrato":
		// TODO: buscar extrato e enviar mensagem
	case "/categorias":
		// TODO: listar categorias de gasto
	case "/tipos_investimento":
		// TODO: listar tipos de investimento
	case "/ajuda":
		// TODO: enviar mensagem de ajuda
	default:
		slog.Info("telegram webhook: unknown command", "command", command)
	}

	return okResponse(), nil
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
