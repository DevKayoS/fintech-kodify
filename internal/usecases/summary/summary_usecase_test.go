package summary

import (
	"context"
	"errors"
	"testing"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/jackc/pgx/v5/pgtype"
)

// mockSummaryRepo implementa SummaryRepository em memória para testes.
type mockSummaryRepo struct {
	user          pgstore.User
	userErr       error
	expenseRows   []pgstore.GetExpenseSummaryByPeriodRow
	expenseErr    error
	revenueTotal  int64
	revenueErr    error
}

func (m *mockSummaryRepo) GetUserByTelegramChatID(_ context.Context, _ pgtype.Int8) (pgstore.User, error) {
	return m.user, m.userErr
}
func (m *mockSummaryRepo) GetExpenseSummaryByPeriod(_ context.Context, _ pgstore.GetExpenseSummaryByPeriodParams) ([]pgstore.GetExpenseSummaryByPeriodRow, error) {
	return m.expenseRows, m.expenseErr
}
func (m *mockSummaryRepo) GetRevenueSummaryByPeriod(_ context.Context, _ pgstore.GetRevenueSummaryByPeriodParams) (int64, error) {
	return m.revenueTotal, m.revenueErr
}

func newUC(repo *mockSummaryRepo) *SummaryUseCase {
	return &SummaryUseCase{repository: repo}
}

// ─── Saldo positivo ───────────────────────────────────────────────────────────

func TestGetMonthlySummary_SaldoPositivo(t *testing.T) {
	repo := &mockSummaryRepo{
		user:         pgstore.User{ID: 1},
		revenueTotal: 500_000, // R$ 5.000,00
		expenseRows: []pgstore.GetExpenseSummaryByPeriodRow{
			{CategorySlug: "alimentacao", CategoryName: "Alimentação", Total: 120_000}, // R$ 1.200,00
			{CategorySlug: "transporte", CategoryName: "Transporte", Total: 80_000},   // R$   800,00
		},
	}
	uc := newUC(repo)

	summary, err := uc.GetMonthlySummary(context.Background(), 42)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if summary.TotalExpenses != 200_000 {
		t.Errorf("TotalExpenses = %d, want 200000", summary.TotalExpenses)
	}
	if summary.TotalRevenues != 500_000 {
		t.Errorf("TotalRevenues = %d, want 500000", summary.TotalRevenues)
	}
	if summary.Balance != 300_000 {
		t.Errorf("Balance = %d, want 300000 (positivo)", summary.Balance)
	}
	if len(summary.ExpensesByCategory) != 2 {
		t.Errorf("ExpensesByCategory len = %d, want 2", len(summary.ExpensesByCategory))
	}
}

// ─── Saldo negativo ───────────────────────────────────────────────────────────

func TestGetMonthlySummary_SaldoNegativo(t *testing.T) {
	repo := &mockSummaryRepo{
		user:         pgstore.User{ID: 1},
		revenueTotal: 100_000, // R$ 1.000,00
		expenseRows: []pgstore.GetExpenseSummaryByPeriodRow{
			{CategorySlug: "moradia", CategoryName: "Moradia", Total: 300_000}, // R$ 3.000,00
		},
	}
	uc := newUC(repo)

	summary, err := uc.GetMonthlySummary(context.Background(), 42)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if summary.Balance != -200_000 {
		t.Errorf("Balance = %d, want -200000 (negativo)", summary.Balance)
	}
}

// ─── Sem receitas ─────────────────────────────────────────────────────────────

func TestGetMonthlySummary_SemReceitas(t *testing.T) {
	repo := &mockSummaryRepo{
		user:         pgstore.User{ID: 1},
		revenueTotal: 0,
		expenseRows: []pgstore.GetExpenseSummaryByPeriodRow{
			{CategorySlug: "lazer", CategoryName: "Lazer", Total: 50_000},
		},
	}
	uc := newUC(repo)

	summary, err := uc.GetMonthlySummary(context.Background(), 42)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if summary.TotalRevenues != 0 {
		t.Errorf("TotalRevenues = %d, want 0", summary.TotalRevenues)
	}
	if summary.Balance != -50_000 {
		t.Errorf("Balance = %d, want -50000", summary.Balance)
	}
}

// ─── Sem gastos ───────────────────────────────────────────────────────────────

func TestGetMonthlySummary_SemGastos(t *testing.T) {
	repo := &mockSummaryRepo{
		user:         pgstore.User{ID: 1},
		revenueTotal: 800_000,
		expenseRows:  nil,
	}
	uc := newUC(repo)

	summary, err := uc.GetMonthlySummary(context.Background(), 42)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if summary.TotalExpenses != 0 {
		t.Errorf("TotalExpenses = %d, want 0", summary.TotalExpenses)
	}
	if summary.Balance != 800_000 {
		t.Errorf("Balance = %d, want 800000", summary.Balance)
	}
	if len(summary.ExpensesByCategory) != 0 {
		t.Errorf("ExpensesByCategory len = %d, want 0", len(summary.ExpensesByCategory))
	}
}

// ─── Mês zerado ───────────────────────────────────────────────────────────────

func TestGetMonthlySummary_MesZerado(t *testing.T) {
	repo := &mockSummaryRepo{
		user:         pgstore.User{ID: 1},
		revenueTotal: 0,
		expenseRows:  nil,
	}
	uc := newUC(repo)

	summary, err := uc.GetMonthlySummary(context.Background(), 42)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if summary.Balance != 0 {
		t.Errorf("Balance = %d, want 0", summary.Balance)
	}
}

// ─── Usuário não encontrado ───────────────────────────────────────────────────

func TestGetMonthlySummary_UsuarioNaoEncontrado(t *testing.T) {
	repo := &mockSummaryRepo{
		userErr: errors.New("not found"),
	}
	uc := newUC(repo)

	_, err := uc.GetMonthlySummary(context.Background(), 99)
	if err == nil {
		t.Fatal("esperava erro para usuário não encontrado")
	}
}

// ─── Categorias separadas corretamente ───────────────────────────────────────

func TestGetMonthlySummary_CategoriasSeparadas(t *testing.T) {
	repo := &mockSummaryRepo{
		user:         pgstore.User{ID: 1},
		revenueTotal: 1_000_000,
		expenseRows: []pgstore.GetExpenseSummaryByPeriodRow{
			{CategorySlug: "alimentacao", CategoryName: "Alimentação", Total: 40_000},
			{CategorySlug: "saude", CategoryName: "Saúde", Total: 25_000},
			{CategorySlug: "educacao", CategoryName: "Educação", Total: 100_000},
		},
	}
	uc := newUC(repo)

	summary, err := uc.GetMonthlySummary(context.Background(), 42)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if len(summary.ExpensesByCategory) != 3 {
		t.Fatalf("ExpensesByCategory len = %d, want 3", len(summary.ExpensesByCategory))
	}

	expectedTotal := int64(40_000 + 25_000 + 100_000)
	if summary.TotalExpenses != expectedTotal {
		t.Errorf("TotalExpenses = %d, want %d", summary.TotalExpenses, expectedTotal)
	}

	// verifica que os slugs foram preservados
	slugs := map[string]bool{}
	for _, c := range summary.ExpensesByCategory {
		slugs[c.Slug] = true
	}
	for _, expected := range []string{"alimentacao", "saude", "educacao"} {
		if !slugs[expected] {
			t.Errorf("categoria %q não encontrada no resumo", expected)
		}
	}
}
