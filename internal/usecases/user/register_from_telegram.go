package user

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
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

// GetStep retorna o passo atual da conversa, ou ("", nil) se não houver estado.
func (uc *UserUseCase) GetStep(ctx context.Context, chatID int64) (string, error) {
	conv, err := uc.repository.GetConversationState(ctx, chatID)
	if err != nil {
		return "", nil // sem estado ativo
	}
	return conv.Step, nil
}

// HandleRegistrationStep processa a resposta do usuário no passo atual e avança o fluxo.
// Retorna a próxima mensagem a enviar ao usuário.
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
		if len(text) < 2 {
			return "Nome muito curto. Por favor, informe seu nome completo:", nil
		}
		data.Name = text
		return uc.advanceTo(ctx, chatID, StepAwaitingCPF, data, "Qual é o seu CPF? (apenas números ou com pontuação)")

	case StepAwaitingCPF:
		data.CPF = text
		return uc.advanceTo(ctx, chatID, StepAwaitingEmail, data, "Qual é o seu e-mail?")

	case StepAwaitingEmail:
		if len(text) < 5 || !containsAt(text) {
			return "E-mail inválido. Por favor, informe um e-mail válido:", nil
		}
		data.Email = text
		return uc.advanceTo(ctx, chatID, StepAwaitingPhone, data, "Qual é o seu telefone? (com DDD)")

	case StepAwaitingPhone:
		if err := uc.finishRegistration(ctx, chatID, data, text); err != nil {
			return "", err
		}
		return fmt.Sprintf("✅ *Conta criada com sucesso, %s!*\n\nAgora você pode:\n• `/gasto <valor> <categoria> <descrição>` — registrar um gasto\n• `/receber <valor> <descrição>` — registrar uma receita\n• `/ajuda` — ver todos os comandos", data.Name), nil
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

func (uc *UserUseCase) finishRegistration(ctx context.Context, chatID int64, data conversationData, phone string) error {
	_, err := uc.repository.InsertUserFromTelegram(ctx, pgstore.InsertUserFromTelegramParams{
		Name:           data.Name,
		Cpf:            pgtype.Text{String: data.CPF, Valid: data.CPF != ""},
		Phone:          pgtype.Text{String: phone, Valid: phone != ""},
		Email:          data.Email,
		TelegramChatID: pgtype.Int8{Int64: chatID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("erro ao criar conta. Verifique se o e-mail já está em uso")
	}
	_ = uc.repository.DeleteConversationState(ctx, chatID)
	return nil
}

func containsAt(s string) bool {
	for _, c := range s {
		if c == '@' {
			return true
		}
	}
	return false
}
