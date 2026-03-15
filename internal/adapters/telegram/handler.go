package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/DevKayoS/fintech-kodify/internal/usecases/expense"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/investment"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/revenue"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/statement"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/summary"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/user"
	"github.com/DevKayoS/fintech-kodify/internal/utils"
	"github.com/aws/aws-lambda-go/events"
)

type Handler struct {
	expenseUC    *expense.ExpenseUseCase
	investmentUC *investment.InvestmentUseCase
	revenueUC    *revenue.RevenueUseCase
	summaryUC    *summary.SummaryUseCase
	statementUC  *statement.StatementUseCase
	userUC       *user.UserUseCase
}

func NewHandler(expenseUC *expense.ExpenseUseCase, investmentUC *investment.InvestmentUseCase, revenueUC *revenue.RevenueUseCase, summaryUC *summary.SummaryUseCase, statementUC *statement.StatementUseCase, userUC *user.UserUseCase) *Handler {
	return &Handler{expenseUC: expenseUC, investmentUC: investmentUC, revenueUC: revenueUC, summaryUC: summaryUC, statementUC: statementUC, userUC: userUC}
}

func (h *Handler) HandleUpdate(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !validateWebhookSecret(req) {
		slog.Warn("telegram webhook: invalid secret header")
		return events.APIGatewayProxyResponse{StatusCode: http.StatusUnauthorized}, nil
	}

	var update TelegramUpdate
	if err := json.Unmarshal([]byte(req.Body), &update); err != nil {
		slog.Error("telegram webhook: failed to parse update", "error", err)
		return okResponse(), nil
	}

	if update.Message == nil || update.Message.Text == "" {
		return okResponse(), nil
	}

	msg := update.Message
	chatID := msg.Chat.ID
	text := strings.TrimSpace(msg.Text)

	slog.Info("telegram message received", "chat_id", chatID, "text", text)

	// Usuário já cadastrado → processa comandos normalmente
	if h.userUC.IsRegistered(ctx, chatID) {
		command, args := parseCommand(text)
		h.handleCommand(ctx, chatID, command, args)
		return okResponse(), nil
	}

	// Usuário em fluxo de cadastro → processa o passo atual
	step, _ := h.userUC.GetStep(ctx, chatID)
	if step != "" {
		reply, err := h.userUC.HandleRegistrationStep(ctx, chatID, text)
		if err != nil {
			sendMessage(chatID, "❌ "+err.Error())
		} else {
			sendMessage(chatID, reply)
		}
		return okResponse(), nil
	}

	// Usuário desconhecido sem fluxo ativo
	command, _ := parseCommand(text)
	if command == "/start" {
		if err := h.userUC.StartRegistration(ctx, chatID); err != nil {
			sendMessage(chatID, "❌ Erro ao iniciar cadastro. Tente novamente.")
			return okResponse(), nil
		}
		sendMessage(chatID, "👋 Bem-vindo ao *Kodify*!\n\nVou precisar de algumas informações para criar sua conta.\n\nQual é o seu nome completo?")
		return okResponse(), nil
	}

	sendMessage(chatID, "Olá! Para usar o bot você precisa se cadastrar primeiro.\n\nDigite /start para começar.")
	return okResponse(), nil
}

func (h *Handler) handleCommand(ctx context.Context, chatID int64, command string, args []string) {
	switch command {
	case "/gasto":
		h.handleGasto(ctx, chatID, args)
	case "/receber":
		h.handleReceber(ctx, chatID, args)
	case "/categorias":
		h.handleCategorias(ctx, chatID)
	case "/resumo":
		h.handleResumo(ctx, chatID)
	case "/resumo-mensal":
		h.handleResumoMensal(ctx, chatID, args)
	case "/extrato":
		h.handleExtrato(ctx, chatID, args)
	case "/investimento":
		h.handleInvestimento(ctx, chatID, args)
	case "/tipos_investimento":
		h.handleTiposInvestimento(ctx, chatID)
	case "/ajuda":
		sendMessage(chatID, helpMessage())
	default:
		sendMessage(chatID, "Comando não reconhecido. Digite /ajuda para ver os comandos disponíveis.")
	}
}

func (h *Handler) handleGasto(ctx context.Context, chatID int64, args []string) {
	if len(args) < 2 {
		sendMessage(chatID, "Uso: `/gasto <valor> <categoria> <descrição>`\nEx: `/gasto 39.90 alimentacao Almoço`\n\nUse /categorias para ver as categorias disponíveis.")
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

func (h *Handler) handleReceber(ctx context.Context, chatID int64, args []string) {
	if len(args) < 1 {
		sendMessage(chatID, "Uso: `/receber <valor> <descrição>`\nEx: `/receber 5000 Salário de março`")
		return
	}

	amountStr := strings.ReplaceAll(args[0], ",", ".")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		sendMessage(chatID, "❌ Valor inválido. Ex: `/receber 5000 Salário de março`")
		return
	}

	description := ""
	if len(args) > 1 {
		description = strings.Join(args[1:], " ")
	}

	result, err := h.revenueUC.CreateFromTelegram(ctx, chatID, amount, description)
	if err != nil {
		sendMessage(chatID, "❌ "+err.Error())
		return
	}

	msg := fmt.Sprintf("✅ *Receita registrada!*\n\n💰 R$ %.2f", result.Amount)
	if result.Description != "" {
		msg += "\n📝 " + result.Description
	}
	sendMessage(chatID, msg)
}

func (h *Handler) handleResumo(ctx context.Context, chatID int64) {
	s, err := h.summaryUC.GetMonthlySummary(ctx, chatID, time.Time{})
	if err != nil {
		sendMessage(chatID, "❌ "+err.Error())
		return
	}

	balanceEmoji := "📈"
	if s.Balance < 0 {
		balanceEmoji = "📉"
	}

	msg := fmt.Sprintf(
		"📊 *Resumo de %s*\n\n"+
			"💸 Gastos: R$ %.2f\n"+
			"💰 Receitas: R$ %.2f\n"+
			"%s Saldo: R$ %.2f\n\n"+
			"📈 Investimentos (total acumulado): R$ %.2f",
		s.Month.Format("01/2006"),
		utils.ToReais(s.TotalExpenses),
		utils.ToReais(s.TotalRevenues),
		balanceEmoji,
		utils.ToReais(s.Balance),
		utils.ToReais(s.AllTimeInvestments),
	)
	sendMessage(chatID, msg)
}

func (h *Handler) handleResumoMensal(ctx context.Context, chatID int64, args []string) {
	var monthStart time.Time
	if len(args) > 0 {
		t, err := time.Parse("01/2006", args[0])
		if err != nil {
			sendMessage(chatID, "❌ Formato de mês inválido. Use MM/AAAA.\nEx: `/resumo-mensal 03/2026`")
			return
		}
		monthStart = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	}

	s, err := h.summaryUC.GetMonthlySummary(ctx, chatID, monthStart)
	if err != nil {
		sendMessage(chatID, "❌ "+err.Error())
		return
	}

	balanceEmoji := "📈"
	if s.Balance < 0 {
		balanceEmoji = "📉"
	}

	msg := fmt.Sprintf("📊 *Resumo Mensal — %s*\n", s.Month.Format("01/2006"))

	if len(s.ExpensesByCategory) > 0 {
		msg += "\n💸 *Gastos por categoria:*\n"
		for _, c := range s.ExpensesByCategory {
			msg += fmt.Sprintf("  • %s: R$ %.2f\n", c.Name, utils.ToReais(c.Total))
		}
		msg += fmt.Sprintf("  _Total: R$ %.2f_\n", utils.ToReais(s.TotalExpenses))
	} else {
		msg += "\n💸 Nenhum gasto registrado este mês.\n"
	}

	msg += fmt.Sprintf("\n💰 *Receitas:* R$ %.2f\n", utils.ToReais(s.TotalRevenues))

	if len(s.InvestmentsByType) > 0 {
		msg += "\n📈 *Investimentos do mês:*\n"
		for _, inv := range s.InvestmentsByType {
			msg += fmt.Sprintf("  • %s: R$ %.2f\n", inv.Name, utils.ToReais(inv.Total))
		}
		msg += fmt.Sprintf("  _Total: R$ %.2f_\n", utils.ToReais(s.TotalInvestments))
	}

	msg += fmt.Sprintf("\n%s *Saldo do mês:* R$ %.2f", balanceEmoji, utils.ToReais(s.Balance))

	sendMessage(chatID, msg)
}

func (h *Handler) handleInvestimento(ctx context.Context, chatID int64, args []string) {
	if len(args) < 2 {
		sendMessage(chatID, "Uso: `/investimento <valor> <tipo> <descrição>`\nEx: `/investimento 500 cdb Tesouro Selic`\n\nUse /tipos_investimento para ver os tipos disponíveis.")
		return
	}

	amountStr := strings.ReplaceAll(args[0], ",", ".")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		sendMessage(chatID, "❌ Valor inválido. Use ponto ou vírgula como separador decimal.\nEx: `/investimento 500 cdb Tesouro Selic`")
		return
	}

	typeSlug := args[1]
	description := ""
	if len(args) > 2 {
		description = strings.Join(args[2:], " ")
	}

	result, err := h.investmentUC.CreateFromTelegram(ctx, chatID, amount, typeSlug, description)
	if err != nil {
		sendMessage(chatID, "❌ "+err.Error())
		return
	}

	msg := fmt.Sprintf("✅ *Investimento registrado!*\n\n📈 R$ %.2f\n🏷️ %s", result.Amount, result.TypeName)
	if result.Description != "" {
		msg += "\n📝 " + result.Description
	}
	sendMessage(chatID, msg)
}

func (h *Handler) handleTiposInvestimento(ctx context.Context, chatID int64) {
	types, err := h.investmentUC.ListTypes(ctx)
	if err != nil {
		sendMessage(chatID, "❌ Erro ao buscar tipos de investimento.")
		return
	}

	msg := "*Tipos de investimento disponíveis:*\n"
	for _, t := range types {
		msg += fmt.Sprintf("• `%s` — %s\n", t.Slug, t.Name)
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

func (h *Handler) handleExtrato(ctx context.Context, chatID int64, args []string) {
	var monthStart time.Time
	if len(args) > 0 {
		t, err := time.Parse("01/2006", args[0])
		if err != nil {
			sendMessage(chatID, "❌ Formato de mês inválido. Use MM/AAAA.\nEx: `/extrato 03/2026`")
			return
		}
		monthStart = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	}

	result, err := h.statementUC.GetStatement(ctx, chatID, monthStart)
	if err != nil {
		sendMessage(chatID, "❌ "+err.Error())
		return
	}

	if len(result.Entries) == 0 {
		sendMessage(chatID, fmt.Sprintf("📋 *Extrato — %s*\n\nNenhuma movimentação encontrada.", result.Month.Format("01/2006")))
		return
	}

	msg := fmt.Sprintf("📋 *Extrato — %s*\n\n", result.Month.Format("01/2006"))
	for _, e := range result.Entries {
		switch e.Kind {
		case "expense":
			line := fmt.Sprintf("💸 %s — %s — R$ %.2f", e.Date.Format("02/01"), e.Label, e.Amount)
			if e.Description != "" {
				line += " — " + e.Description
			}
			msg += line + "\n"
		case "revenue":
			line := fmt.Sprintf("💰 %s — R$ %.2f", e.Date.Format("02/01"), e.Amount)
			if e.Description != "" {
				line += " — " + e.Description
			}
			msg += line + "\n"
		case "investment":
			line := fmt.Sprintf("📈 %s — %s — R$ %.2f", e.Date.Format("02/01"), e.Label, e.Amount)
			if e.Description != "" {
				line += " — " + e.Description
			}
			msg += line + "\n"
		}
	}

	sendMessage(chatID, msg)
}

func parseCommand(text string) (string, []string) {
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return "", nil
	}
	command := parts[0]
	if idx := strings.Index(command, "@"); idx != -1 {
		command = command[:idx]
	}
	return command, parts[1:]
}

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
