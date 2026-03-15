package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	StepAwaitingName  = "awaiting_name"
	StepAwaitingCPF   = "awaiting_cpf"
	StepAwaitingEmail = "awaiting_email"
	StepAwaitingPhone = "awaiting_phone"
)

type conversationData struct {
	Name  string `json:"name,omitempty"`
	CPF   string `json:"cpf,omitempty"`
	Email string `json:"email,omitempty"`
}

// IsRegistered verifica se o chat_id já tem uma conta vinculada.
func (uc *UserUseCase) IsRegistered(ctx context.Context, chatID int64) bool {
	_, err := uc.repository.GetUserByTelegramChatID(ctx, pgtype.Int8{Int64: chatID, Valid: true})
	return err == nil
}

// StartRegistration inicia o fluxo de cadastro para um novo usuário.
func (uc *UserUseCase) StartRegistration(ctx context.Context, chatID int64) error {
	return uc.repository.UpsertConversationState(ctx, pgstore.UpsertConversationStateParams{
		ChatID: chatID,
		Step:   StepAwaitingName,
		Data:   "{}",
	})
}

// CancelRegistration cancela o fluxo de cadastro em andamento, se houver.
func (uc *UserUseCase) CancelRegistration(ctx context.Context, chatID int64) error {
	return uc.repository.DeleteConversationState(ctx, chatID)
}

// GetStep retorna o passo atual da conversa, ou ("", nil) se não houver estado.
func (uc *UserUseCase) GetStep(ctx context.Context, chatID int64) (string, error) {
	conv, err := uc.repository.GetConversationState(ctx, chatID)
	if err != nil {
		return "", nil // sem estado ativo
	}
	return conv.Step, nil
}

// HandleRegistrationStep processa a resposta do usuário no passo atual e avança o fluxo.
// Retorna (mensagem, nil) quando deve permanecer no passo ou avançar.
// Retorna ("", error) apenas em erros internos irrecuperáveis.
func (uc *UserUseCase) HandleRegistrationStep(ctx context.Context, chatID int64, text string) (string, error) {
	conv, err := uc.repository.GetConversationState(ctx, chatID)
	if err != nil {
		return "", fmt.Errorf("estado de conversa não encontrado")
	}

	var data conversationData
	if err := json.Unmarshal([]byte(conv.Data), &data); err != nil {
		return "", fmt.Errorf("erro interno ao ler dados da conversa")
	}

	switch conv.Step {
	case StepAwaitingName:
		if len(strings.TrimSpace(text)) < 2 {
			return "Nome muito curto. Por favor, informe seu nome completo:", nil
		}
		data.Name = strings.TrimSpace(text)
		return uc.advanceTo(ctx, chatID, StepAwaitingCPF, data, "Qual é o seu CPF? (apenas números ou com pontuação)")

	case StepAwaitingCPF:
		if !ValidateCPF(text) {
			return "CPF inválido. Por favor, informe um CPF válido.\nEx: `123.456.789-09` ou `12345678909`", nil
		}
		data.CPF = onlyDigits(text)
		return uc.advanceTo(ctx, chatID, StepAwaitingEmail, data, "Qual é o seu e-mail?")

	case StepAwaitingEmail:
		if !ValidateEmail(text) {
			return "E-mail inválido. Por favor, informe um e-mail válido.\nEx: `joao@email.com`", nil
		}
		data.Email = strings.TrimSpace(text)
		return uc.advanceTo(ctx, chatID, StepAwaitingPhone, data, "Qual é o seu telefone? (com DDD, ex: 11999999999)")

	case StepAwaitingPhone:
		if !ValidatePhone(text) {
			return "Telefone inválido. Informe um número com DDD (10 ou 11 dígitos).\nEx: `11999999999`", nil
		}
		msg, err := uc.finishRegistration(ctx, chatID, data, onlyDigits(text))
		if err != nil {
			// erros de duplicata retornam mensagem amigável para o usuário poder corrigir
			return err.Error(), nil
		}
		return msg, nil
	}

	return "", fmt.Errorf("passo desconhecido: %s", conv.Step)
}

func (uc *UserUseCase) advanceTo(ctx context.Context, chatID int64, nextStep string, data conversationData, question string) (string, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("erro interno")
	}
	if err := uc.repository.UpsertConversationState(ctx, pgstore.UpsertConversationStateParams{
		ChatID: chatID,
		Step:   nextStep,
		Data:   string(raw),
	}); err != nil {
		return "", fmt.Errorf("erro ao salvar progresso")
	}
	return question, nil
}

func (uc *UserUseCase) finishRegistration(ctx context.Context, chatID int64, data conversationData, phone string) (string, error) {
	_, err := uc.repository.InsertUserFromTelegram(ctx, pgstore.InsertUserFromTelegramParams{
		Name:           data.Name,
		Cpf:            pgtype.Text{String: data.CPF, Valid: data.CPF != ""},
		Phone:          pgtype.Text{String: phone, Valid: phone != ""},
		Email:          data.Email,
		TelegramChatID: pgtype.Int8{Int64: chatID, Valid: true},
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgErrUniqueViolation {
			if strings.Contains(pgErr.ConstraintName, pgConstraintCPF) {
				return "", fmt.Errorf("❌ Este CPF já está cadastrado em outra conta.\n\nInforme outro CPF:")
			}
			if strings.Contains(pgErr.ConstraintName, pgConstraintEmail) {
				return "", fmt.Errorf("❌ Este e-mail já está cadastrado.\n\nInforme outro e-mail:")
			}
		}
		return "", fmt.Errorf("❌ Erro ao criar conta, tente novamente")
	}

	_ = uc.repository.DeleteConversationState(ctx, chatID)
	return fmt.Sprintf("✅ *Conta criada com sucesso, %s!*\n\nAgora você pode:\n• `/gasto <valor> <categoria> <descrição>` — registrar um gasto\n• `/receber <valor> <descrição>` — registrar uma receita\n• `/ajuda` — ver todos os comandos", data.Name), nil
}
