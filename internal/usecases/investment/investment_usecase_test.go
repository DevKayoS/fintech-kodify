package investment

import (
	"context"
	"errors"
	"testing"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockRepo struct {
	user       pgstore.User
	userErr    error
	invType    pgstore.InvestmentType
	invTypeErr error
	insertErr  error
	inserted   *pgstore.InsertInvestmentParams
}

func (m *mockRepo) GetUserByTelegramChatID(_ context.Context, _ pgtype.Int8) (pgstore.User, error) {
	return m.user, m.userErr
}
func (m *mockRepo) GetInvestmentTypeBySlug(_ context.Context, _ string) (pgstore.InvestmentType, error) {
	return m.invType, m.invTypeErr
}
func (m *mockRepo) InsertInvestment(_ context.Context, arg pgstore.InsertInvestmentParams) (pgstore.Investment, error) {
	m.inserted = &arg
	return pgstore.Investment{ID: 1, UserID: arg.UserID, Amount: arg.Amount}, m.insertErr
}
func (m *mockRepo) ListInvestmentTypes(_ context.Context) ([]pgstore.InvestmentType, error) {
	return []pgstore.InvestmentType{m.invType}, nil
}
func (m *mockRepo) GetInvestmentByID(_ context.Context, _ pgstore.GetInvestmentByIDParams) (pgstore.GetInvestmentByIDRow, error) {
	return pgstore.GetInvestmentByIDRow{}, nil
}
func (m *mockRepo) DeleteInvestment(_ context.Context, _ pgstore.DeleteInvestmentParams) error {
	return nil
}
func (m *mockRepo) ListInvestmentsByUser(_ context.Context, _ int64) ([]pgstore.ListInvestmentsByUserRow, error) {
	return nil, nil
}
func (m *mockRepo) ListInvestmentsByUserAndPeriod(_ context.Context, _ pgstore.ListInvestmentsByUserAndPeriodParams) ([]pgstore.ListInvestmentsByUserAndPeriodRow, error) {
	return nil, nil
}

func newUC(repo *mockRepo) *InvestmentUseCase {
	return &InvestmentUseCase{repository: repo}
}

// ─── Criação bem-sucedida ─────────────────────────────────────────────────────

func TestCreateFromTelegram_Sucesso(t *testing.T) {
	repo := &mockRepo{
		user:    pgstore.User{ID: 7},
		invType: pgstore.InvestmentType{ID: 1, Slug: "cdb", Name: "CDB"},
	}
	uc := newUC(repo)

	result, err := uc.CreateFromTelegram(context.Background(), 42, 1000.00, "cdb", "Tesouro Selic")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.Amount != 1000.00 {
		t.Errorf("Amount = %.2f, want 1000.00", result.Amount)
	}
	if result.TypeName != "CDB" {
		t.Errorf("TypeName = %q, want %q", result.TypeName, "CDB")
	}
	if result.Description != "Tesouro Selic" {
		t.Errorf("Description = %q, want %q", result.Description, "Tesouro Selic")
	}
}

// ─── Conversão reais → centavos ───────────────────────────────────────────────

func TestCreateFromTelegram_ConversaoCentavos(t *testing.T) {
	repo := &mockRepo{
		user:    pgstore.User{ID: 7},
		invType: pgstore.InvestmentType{ID: 1},
	}
	uc := newUC(repo)

	cases := []struct {
		reais    float64
		centavos int64
	}{
		{500.00, 50_000},
		{1234.56, 123_456},
		{0.01, 1},
		{10000.00, 1_000_000},
	}

	for _, c := range cases {
		repo.inserted = nil
		_, err := uc.CreateFromTelegram(context.Background(), 42, c.reais, "cdb", "")
		if err != nil {
			t.Fatalf("R$ %.2f: erro inesperado: %v", c.reais, err)
		}
		if repo.inserted.Amount != c.centavos {
			t.Errorf("R$ %.2f → %d centavos, want %d", c.reais, repo.inserted.Amount, c.centavos)
		}
	}
}

// ─── Sem descrição ────────────────────────────────────────────────────────────

func TestCreateFromTelegram_SemDescricao(t *testing.T) {
	repo := &mockRepo{
		user:    pgstore.User{ID: 7},
		invType: pgstore.InvestmentType{ID: 1},
	}
	uc := newUC(repo)

	_, err := uc.CreateFromTelegram(context.Background(), 42, 500.00, "cdb", "")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if repo.inserted.Description.Valid {
		t.Error("Description.Valid deveria ser false quando descrição está vazia")
	}
}

// ─── Usuário não encontrado ───────────────────────────────────────────────────

func TestCreateFromTelegram_UsuarioNaoEncontrado(t *testing.T) {
	repo := &mockRepo{userErr: errors.New("not found")}
	uc := newUC(repo)

	_, err := uc.CreateFromTelegram(context.Background(), 99, 500.00, "cdb", "")
	if err == nil {
		t.Fatal("esperava erro para usuário não encontrado")
	}
}

// ─── Tipo de investimento inválido ────────────────────────────────────────────

func TestCreateFromTelegram_TipoInvalido(t *testing.T) {
	repo := &mockRepo{
		user:       pgstore.User{ID: 7},
		invTypeErr: errors.New("not found"),
	}
	uc := newUC(repo)

	_, err := uc.CreateFromTelegram(context.Background(), 42, 500.00, "invalido", "")
	if err == nil {
		t.Fatal("esperava erro para tipo inválido")
	}
}

// ─── Erro no banco ao inserir ─────────────────────────────────────────────────

func TestCreateFromTelegram_ErroBanco(t *testing.T) {
	repo := &mockRepo{
		user:      pgstore.User{ID: 7},
		invType:   pgstore.InvestmentType{ID: 1},
		insertErr: errors.New("db error"),
	}
	uc := newUC(repo)

	_, err := uc.CreateFromTelegram(context.Background(), 42, 500.00, "cdb", "")
	if err == nil {
		t.Fatal("esperava erro quando banco falha")
	}
}

// ─── UserID e TypeID corretos passados ao banco ───────────────────────────────

func TestCreateFromTelegram_IDsCorretos(t *testing.T) {
	repo := &mockRepo{
		user:    pgstore.User{ID: 99},
		invType: pgstore.InvestmentType{ID: 3},
	}
	uc := newUC(repo)

	_, err := uc.CreateFromTelegram(context.Background(), 42, 100.00, "acoes", "")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if repo.inserted.UserID != 99 {
		t.Errorf("UserID = %d, want 99", repo.inserted.UserID)
	}
	if repo.inserted.InvestmentTypeID != 3 {
		t.Errorf("InvestmentTypeID = %d, want 3", repo.inserted.InvestmentTypeID)
	}
}

// ─── ListTypes ────────────────────────────────────────────────────────────────

func TestListTypes_RetornaLista(t *testing.T) {
	repo := &mockRepo{
		invType: pgstore.InvestmentType{ID: 1, Slug: "cdb", Name: "CDB"},
	}
	uc := newUC(repo)

	types, err := uc.ListTypes(context.Background())
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(types) == 0 {
		t.Fatal("esperava pelo menos um tipo de investimento")
	}
	if types[0].Slug != "cdb" {
		t.Errorf("Slug = %q, want %q", types[0].Slug, "cdb")
	}
}
