package user

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// mockRepo implementa UserRepository em memória para testes.
type mockRepo struct {
	conversations map[int64]pgstore.TelegramConversation
	insertErr     error
}

func newMockRepo() *mockRepo {
	return &mockRepo{conversations: make(map[int64]pgstore.TelegramConversation)}
}

func (m *mockRepo) GetUserByTelegramChatID(_ context.Context, _ pgtype.Int8) (pgstore.User, error) {
	return pgstore.User{}, errNotFound
}
func (m *mockRepo) InsertLinkToken(_ context.Context, _ pgstore.InsertLinkTokenParams) (int64, error) {
	return 0, nil
}
func (m *mockRepo) GetValidLinkToken(_ context.Context, _ string) (pgstore.TelegramLinkToken, error) {
	return pgstore.TelegramLinkToken{}, nil
}
func (m *mockRepo) MarkLinkTokenUsed(_ context.Context, _ int64) error { return nil }
func (m *mockRepo) UpdateUserTelegramChatID(_ context.Context, _ pgstore.UpdateUserTelegramChatIDParams) error {
	return nil
}
func (m *mockRepo) InsertUserFromTelegram(_ context.Context, _ pgstore.InsertUserFromTelegramParams) (int64, error) {
	return 1, m.insertErr
}
func (m *mockRepo) GetConversationState(_ context.Context, chatID int64) (pgstore.TelegramConversation, error) {
	c, ok := m.conversations[chatID]
	if !ok {
		return pgstore.TelegramConversation{}, errNotFound
	}
	return c, nil
}
func (m *mockRepo) UpsertConversationState(_ context.Context, arg pgstore.UpsertConversationStateParams) error {
	m.conversations[arg.ChatID] = pgstore.TelegramConversation{
		ChatID: arg.ChatID,
		Step:   arg.Step,
		Data:   arg.Data,
	}
	return nil
}
func (m *mockRepo) DeleteConversationState(_ context.Context, chatID int64) error {
	delete(m.conversations, chatID)
	return nil
}

// errNotFound é um sentinel error genérico para entidade não encontrada no mock.
var errNotFound = &pgconn.PgError{Code: "02000"}

// helpers para montar o estado inicial da conversa num passo específico.
func setStep(repo *mockRepo, chatID int64, step string, data conversationData) {
	raw, _ := json.Marshal(data)
	repo.conversations[chatID] = pgstore.TelegramConversation{
		ChatID: chatID,
		Step:   step,
		Data:   string(raw),
	}
}

func newUC(repo *mockRepo) *UserUseCase {
	return &UserUseCase{repository: repo}
}

// ─── ValidateCPF ─────────────────────────────────────────────────────────────

func TestValidateCPF(t *testing.T) {
	tests := []struct {
		name  string
		cpf   string
		valid bool
	}{
		{"válido com pontuação", "529.982.247-25", true},
		{"válido só números", "52998224725", true},
		{"inválido sequencial", "123.123.123-41", false},
		{"inválido dígito errado", "12312312341", false},
		{"todos iguais 1", "111.111.111-11", false},
		{"todos iguais 0", "000.000.000-00", false},
		{"curto demais", "1234567", false},
		{"longo demais", "123456789012", false},
		{"vazio", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateCPF(tt.cpf); got != tt.valid {
				t.Errorf("ValidateCPF(%q) = %v, want %v", tt.cpf, got, tt.valid)
			}
		})
	}
}

// ─── ValidatePhone ───────────────────────────────────────────────────────────

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		valid bool
	}{
		{"celular 11 dígitos", "11999999999", true},
		{"fixo 10 dígitos", "1133334444", true},
		{"com formatação", "(11) 99999-9999", true},
		{"curto demais", "9999999", false},
		{"longo demais", "119999999999", false},
		{"vazio", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidatePhone(tt.phone); got != tt.valid {
				t.Errorf("ValidatePhone(%q) = %v, want %v", tt.phone, got, tt.valid)
			}
		})
	}
}

// ─── ValidateEmail ───────────────────────────────────────────────────────────

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{"válido", "joao@email.com", true},
		{"válido com subdomínio", "joao@mail.empresa.com.br", true},
		{"sem @", "joaoemail.com", false},
		{"sem domínio", "joao@", false},
		{"sem TLD", "joao@email", false},
		{"só @", "@", false},
		{"vazio", "", false},
		{"com espaço", "joao @email.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateEmail(tt.email); got != tt.valid {
				t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, got, tt.valid)
			}
		})
	}
}

// ─── HandleRegistrationStep — passo CPF ──────────────────────────────────────

func TestHandleRegistrationStep_CPF_Invalido(t *testing.T) {
	repo := newMockRepo()
	uc := newUC(repo)
	ctx := context.Background()

	setStep(repo, 1, StepAwaitingCPF, conversationData{Name: "João Silva"})

	msg, err := uc.HandleRegistrationStep(ctx, 1, "12312312341")

	if err != nil {
		t.Fatalf("esperava nil error, got %v", err)
	}
	if msg == "" {
		t.Fatal("esperava mensagem de validação, got vazio")
	}
	// deve permanecer no passo de CPF
	if repo.conversations[1].Step != StepAwaitingCPF {
		t.Errorf("passo esperado %q, got %q", StepAwaitingCPF, repo.conversations[1].Step)
	}
}

func TestHandleRegistrationStep_CPF_Valido(t *testing.T) {
	repo := newMockRepo()
	uc := newUC(repo)
	ctx := context.Background()

	setStep(repo, 1, StepAwaitingCPF, conversationData{Name: "João Silva"})

	_, err := uc.HandleRegistrationStep(ctx, 1, "529.982.247-25")

	if err != nil {
		t.Fatalf("CPF válido não deveria retornar erro: %v", err)
	}
	if repo.conversations[1].Step != StepAwaitingEmail {
		t.Errorf("passo esperado %q, got %q", StepAwaitingEmail, repo.conversations[1].Step)
	}
}

func TestHandleRegistrationStep_CPF_TodosIguais(t *testing.T) {
	repo := newMockRepo()
	uc := newUC(repo)
	ctx := context.Background()

	setStep(repo, 1, StepAwaitingCPF, conversationData{Name: "João Silva"})

	msg, err := uc.HandleRegistrationStep(ctx, 1, "111.111.111-11")

	if err != nil {
		t.Fatalf("esperava nil error, got %v", err)
	}
	if msg == "" {
		t.Fatal("esperava mensagem de validação")
	}
	if repo.conversations[1].Step != StepAwaitingCPF {
		t.Errorf("deveria permanecer em %q, got %q", StepAwaitingCPF, repo.conversations[1].Step)
	}
}

// ─── HandleRegistrationStep — passo Email ────────────────────────────────────

func TestHandleRegistrationStep_Email_Invalido(t *testing.T) {
	repo := newMockRepo()
	uc := newUC(repo)
	ctx := context.Background()

	setStep(repo, 1, StepAwaitingEmail, conversationData{Name: "João", CPF: "52998224725"})

	msg, err := uc.HandleRegistrationStep(ctx, 1, "emailsemarroba")

	if err != nil {
		t.Fatalf("esperava nil error, got %v", err)
	}
	if msg == "" {
		t.Fatal("esperava mensagem de validação")
	}
	if repo.conversations[1].Step != StepAwaitingEmail {
		t.Errorf("deveria permanecer em %q, got %q", StepAwaitingEmail, repo.conversations[1].Step)
	}
}

func TestHandleRegistrationStep_Email_Valido(t *testing.T) {
	repo := newMockRepo()
	uc := newUC(repo)
	ctx := context.Background()

	setStep(repo, 1, StepAwaitingEmail, conversationData{Name: "João", CPF: "52998224725"})

	_, err := uc.HandleRegistrationStep(ctx, 1, "joao@email.com")

	if err != nil {
		t.Fatalf("e-mail válido não deveria retornar erro: %v", err)
	}
	if repo.conversations[1].Step != StepAwaitingPhone {
		t.Errorf("passo esperado %q, got %q", StepAwaitingPhone, repo.conversations[1].Step)
	}
}

// ─── HandleRegistrationStep — passo Telefone ─────────────────────────────────

func TestHandleRegistrationStep_Telefone_Invalido(t *testing.T) {
	repo := newMockRepo()
	uc := newUC(repo)
	ctx := context.Background()

	setStep(repo, 1, StepAwaitingPhone, conversationData{Name: "João", CPF: "52998224725", Email: "joao@email.com"})

	msg, err := uc.HandleRegistrationStep(ctx, 1, "9999")

	if err != nil {
		t.Fatalf("esperava nil error, got %v", err)
	}
	if msg == "" {
		t.Fatal("esperava mensagem de validação")
	}
	if repo.conversations[1].Step != StepAwaitingPhone {
		t.Errorf("deveria permanecer em %q, got %q", StepAwaitingPhone, repo.conversations[1].Step)
	}
}

// ─── HandleRegistrationStep — duplicatas ─────────────────────────────────────

func TestHandleRegistrationStep_CPF_JaCadastrado(t *testing.T) {
	repo := newMockRepo()
	repo.insertErr = &pgconn.PgError{Code: pgErrUniqueViolation, ConstraintName: pgConstraintCPFKey}
	uc := newUC(repo)
	ctx := context.Background()

	setStep(repo, 1, StepAwaitingPhone, conversationData{Name: "João", CPF: "52998224725", Email: "joao@email.com"})

	msg, err := uc.HandleRegistrationStep(ctx, 1, "11999999999")

	if err != nil {
		t.Fatalf("duplicata de CPF deveria retornar mensagem, não error: %v", err)
	}
	if msg == "" {
		t.Fatal("esperava mensagem informando que CPF já está cadastrado")
	}
	// conversa deve ser mantida para o usuário poder corrigir
	if _, ok := repo.conversations[1]; !ok {
		t.Error("estado da conversa não deveria ter sido apagado após erro de duplicata")
	}
}

func TestHandleRegistrationStep_Email_JaCadastrado(t *testing.T) {
	repo := newMockRepo()
	repo.insertErr = &pgconn.PgError{Code: pgErrUniqueViolation, ConstraintName: pgConstraintEmailKey}
	uc := newUC(repo)
	ctx := context.Background()

	setStep(repo, 1, StepAwaitingPhone, conversationData{Name: "João", CPF: "52998224725", Email: "joao@email.com"})

	msg, err := uc.HandleRegistrationStep(ctx, 1, "11999999999")

	if err != nil {
		t.Fatalf("duplicata de e-mail deveria retornar mensagem, não error: %v", err)
	}
	if msg == "" {
		t.Fatal("esperava mensagem informando que e-mail já está cadastrado")
	}
}

// ─── Fluxo completo ───────────────────────────────────────────────────────────

func TestRegistrationFlow_Completo(t *testing.T) {
	repo := newMockRepo()
	uc := newUC(repo)
	ctx := context.Background()
	const chatID int64 = 42

	// Inicia cadastro
	if err := uc.StartRegistration(ctx, chatID); err != nil {
		t.Fatalf("StartRegistration: %v", err)
	}
	if repo.conversations[chatID].Step != StepAwaitingName {
		t.Fatalf("passo inicial esperado %q", StepAwaitingName)
	}

	steps := []struct {
		input    string
		wantStep string
	}{
		{"João Silva", StepAwaitingCPF},
		{"529.982.247-25", StepAwaitingEmail},
		{"joao@email.com", StepAwaitingPhone},
	}

	for _, s := range steps {
		msg, err := uc.HandleRegistrationStep(ctx, chatID, s.input)
		if err != nil {
			t.Fatalf("step %q: erro inesperado: %v", s.input, err)
		}
		if msg == "" {
			t.Fatalf("step %q: mensagem vazia", s.input)
		}
		if repo.conversations[chatID].Step != s.wantStep {
			t.Fatalf("step %q: passo esperado %q, got %q", s.input, s.wantStep, repo.conversations[chatID].Step)
		}
	}

	// Último passo: telefone → cria conta
	msg, err := uc.HandleRegistrationStep(ctx, chatID, "11999999999")
	if err != nil {
		t.Fatalf("último passo: erro inesperado: %v", err)
	}
	if msg == "" {
		t.Fatal("esperava mensagem de confirmação")
	}
	// estado deve ter sido apagado após cadastro bem-sucedido
	if _, ok := repo.conversations[chatID]; ok {
		t.Error("estado da conversa deveria ter sido apagado após cadastro")
	}
}
