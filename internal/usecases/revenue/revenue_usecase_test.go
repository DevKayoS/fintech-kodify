package revenue

import (
	"context"
	"errors"
	"testing"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockRevenueRepo struct {
	user      pgstore.User
	userErr   error
	insertErr error
	inserted  *pgstore.InsertRevenueParams
}

func (m *mockRevenueRepo) GetUserByTelegramChatID(_ context.Context, _ pgtype.Int8) (pgstore.User, error) {
	return m.user, m.userErr
}
func (m *mockRevenueRepo) InsertRevenue(_ context.Context, arg pgstore.InsertRevenueParams) (pgstore.Revenue, error) {
	m.inserted = &arg
	return pgstore.Revenue{
		ID:     1,
		UserID: arg.UserID,
		Amount: arg.Amount,
	}, m.insertErr
}
func (m *mockRevenueRepo) ListRevenuesByUser(_ context.Context, _ int64) ([]pgstore.Revenue, error) {
	return nil, nil
}
func (m *mockRevenueRepo) ListRevenuesByUserAndPeriod(_ context.Context, _ pgstore.ListRevenuesByUserAndPeriodParams) ([]pgstore.Revenue, error) {
	return nil, nil
}

func newUC(repo *mockRevenueRepo) *RevenueUseCase {
	return &RevenueUseCase{repository: repo}
}

// ─── Criação bem-sucedida ─────────────────────────────────────────────────────

func TestCreateFromTelegram_Sucesso(t *testing.T) {
	repo := &mockRevenueRepo{user: pgstore.User{ID: 5}}
	uc := newUC(repo)

	result, err := uc.CreateFromTelegram(context.Background(), 42, 1500.00, "Salário")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.Amount != 1500.00 {
		t.Errorf("Amount = %.2f, want 1500.00", result.Amount)
	}
	if result.Description != "Salário" {
		t.Errorf("Description = %q, want %q", result.Description, "Salário")
	}
}

// ─── Conversão reais → centavos ───────────────────────────────────────────────

func TestCreateFromTelegram_ConversaoCentavos(t *testing.T) {
	repo := &mockRevenueRepo{user: pgstore.User{ID: 5}}
	uc := newUC(repo)

	cases := []struct {
		reais    float64
		centavos int64
	}{
		{39.90, 3990},
		{1000.00, 100_000},
		{0.01, 1},
		{5000.50, 500_050},
	}

	for _, c := range cases {
		repo.inserted = nil
		_, err := uc.CreateFromTelegram(context.Background(), 42, c.reais, "")
		if err != nil {
			t.Fatalf("R$ %.2f: erro inesperado: %v", c.reais, err)
		}
		if repo.inserted == nil {
			t.Fatalf("R$ %.2f: InsertRevenue não foi chamado", c.reais)
		}
		if repo.inserted.Amount != c.centavos {
			t.Errorf("R$ %.2f → %d centavos, want %d", c.reais, repo.inserted.Amount, c.centavos)
		}
	}
}

// ─── Sem descrição ────────────────────────────────────────────────────────────

func TestCreateFromTelegram_SemDescricao(t *testing.T) {
	repo := &mockRevenueRepo{user: pgstore.User{ID: 5}}
	uc := newUC(repo)

	result, err := uc.CreateFromTelegram(context.Background(), 42, 500.00, "")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.Description != "" {
		t.Errorf("Description = %q, want vazio", result.Description)
	}
	if repo.inserted.Description.Valid {
		t.Error("Description.Valid deveria ser false quando descrição está vazia")
	}
}

// ─── Usuário não encontrado ───────────────────────────────────────────────────

func TestCreateFromTelegram_UsuarioNaoEncontrado(t *testing.T) {
	repo := &mockRevenueRepo{userErr: errors.New("not found")}
	uc := newUC(repo)

	_, err := uc.CreateFromTelegram(context.Background(), 99, 100.00, "Teste")
	if err == nil {
		t.Fatal("esperava erro para usuário não encontrado")
	}
}

// ─── Erro no banco ao inserir ─────────────────────────────────────────────────

func TestCreateFromTelegram_ErroBanco(t *testing.T) {
	repo := &mockRevenueRepo{
		user:      pgstore.User{ID: 5},
		insertErr: errors.New("db error"),
	}
	uc := newUC(repo)

	_, err := uc.CreateFromTelegram(context.Background(), 42, 100.00, "Teste")
	if err == nil {
		t.Fatal("esperava erro quando banco falha")
	}
}

// ─── UserID correto passado ao banco ──────────────────────────────────────────

func TestCreateFromTelegram_UserIDCorreto(t *testing.T) {
	repo := &mockRevenueRepo{user: pgstore.User{ID: 99}}
	uc := newUC(repo)

	_, err := uc.CreateFromTelegram(context.Background(), 42, 100.00, "")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if repo.inserted.UserID != 99 {
		t.Errorf("UserID = %d, want 99", repo.inserted.UserID)
	}
}
