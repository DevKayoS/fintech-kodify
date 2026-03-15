package summary

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockSummaryRepo struct {
	user               pgstore.User
	userErr            error
	expenseRows        []pgstore.GetExpenseSummaryByPeriodRow
	expenseErr         error
	revenueTotal       int64
	revenueErr         error
	investmentRows     []pgstore.GetInvestmentSummaryByPeriodRow
	investmentErr      error
	allTimeInvestments int64
	allTimeErr         error
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
func (m *mockSummaryRepo) GetInvestmentSummaryByPeriod(_ context.Context, _ pgstore.GetInvestmentSummaryByPeriodParams) ([]pgstore.GetInvestmentSummaryByPeriodRow, error) {
	return m.investmentRows, m.investmentErr
}
func (m *mockSummaryRepo) GetTotalInvestmentsByUser(_ context.Context, _ int64) (int64, error) {
	return m.allTimeInvestments, m.allTimeErr
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
			{CategorySlug: "alimentacao", CategoryName: "Alimentação", Total: 120_000},
			{CategorySlug: "transporte", CategoryName: "Transporte", Total: 80_000},
		},
		allTimeInvestments: 1_000_000, // R$ 10.000,00
	}
	uc := newUC(repo)

	s, err := uc.GetMonthlySummary(context.Background(), 42, time.Time{})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if s.TotalExpenses != 200_000 {
		t.Errorf("TotalExpenses = %d, want 200000", s.TotalExpenses)
	}
	if s.TotalRevenues != 500_000 {
		t.Errorf("TotalRevenues = %d, want 500000", s.TotalRevenues)
	}
	if s.Balance != 300_000 {
		t.Errorf("Balance = %d, want 300000", s.Balance)
	}
	if s.AllTimeInvestments != 1_000_000 {
		t.Errorf("AllTimeInvestments = %d, want 1000000", s.AllTimeInvestments)
	}
}

// ─── Saldo negativo ───────────────────────────────────────────────────────────

func TestGetMonthlySummary_SaldoNegativo(t *testing.T) {
	repo := &mockSummaryRepo{
		user:         pgstore.User{ID: 1},
		revenueTotal: 100_000,
		expenseRows: []pgstore.GetExpenseSummaryByPeriodRow{
			{CategorySlug: "moradia", CategoryName: "Moradia", Total: 300_000},
		},
	}

	s, err := newUC(repo).GetMonthlySummary(context.Background(), 42, time.Time{})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if s.Balance != -200_000 {
		t.Errorf("Balance = %d, want -200000", s.Balance)
	}
}

// ─── Sem receitas ─────────────────────────────────────────────────────────────

func TestGetMonthlySummary_SemReceitas(t *testing.T) {
	repo := &mockSummaryRepo{
		user:         pgstore.User{ID: 1},
		revenueTotal: 0,
		expenseRows:  []pgstore.GetExpenseSummaryByPeriodRow{{Total: 50_000}},
	}

	s, err := newUC(repo).GetMonthlySummary(context.Background(), 42, time.Time{})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if s.Balance != -50_000 {
		t.Errorf("Balance = %d, want -50000", s.Balance)
	}
}

// ─── Sem gastos ───────────────────────────────────────────────────────────────

func TestGetMonthlySummary_SemGastos(t *testing.T) {
	repo := &mockSummaryRepo{
		user:         pgstore.User{ID: 1},
		revenueTotal: 800_000,
		expenseRows:  nil,
	}

	s, err := newUC(repo).GetMonthlySummary(context.Background(), 42, time.Time{})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if s.TotalExpenses != 0 {
		t.Errorf("TotalExpenses = %d, want 0", s.TotalExpenses)
	}
	if s.Balance != 800_000 {
		t.Errorf("Balance = %d, want 800000", s.Balance)
	}
}

// ─── Mês zerado ───────────────────────────────────────────────────────────────

func TestGetMonthlySummary_MesZerado(t *testing.T) {
	repo := &mockSummaryRepo{user: pgstore.User{ID: 1}}

	s, err := newUC(repo).GetMonthlySummary(context.Background(), 42, time.Time{})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if s.Balance != 0 {
		t.Errorf("Balance = %d, want 0", s.Balance)
	}
	if s.TotalInvestments != 0 {
		t.Errorf("TotalInvestments = %d, want 0", s.TotalInvestments)
	}
}

// ─── Usuário não encontrado ───────────────────────────────────────────────────

func TestGetMonthlySummary_UsuarioNaoEncontrado(t *testing.T) {
	repo := &mockSummaryRepo{userErr: errors.New("not found")}

	_, err := newUC(repo).GetMonthlySummary(context.Background(), 42, time.Time{})
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

	s, err := newUC(repo).GetMonthlySummary(context.Background(), 42, time.Time{})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(s.ExpensesByCategory) != 3 {
		t.Fatalf("ExpensesByCategory len = %d, want 3", len(s.ExpensesByCategory))
	}
	if s.TotalExpenses != 165_000 {
		t.Errorf("TotalExpenses = %d, want 165000", s.TotalExpenses)
	}
}

// ─── Investimentos do mês separados por tipo ─────────────────────────────────

func TestGetMonthlySummary_InvestimentosPorTipo(t *testing.T) {
	repo := &mockSummaryRepo{
		user:         pgstore.User{ID: 1},
		revenueTotal: 500_000,
		investmentRows: []pgstore.GetInvestmentSummaryByPeriodRow{
			{TypeSlug: "cdb", TypeName: "CDB", Total: 200_000},
			{TypeSlug: "acoes", TypeName: "Ações", Total: 100_000},
		},
		allTimeInvestments: 1_500_000,
	}

	s, err := newUC(repo).GetMonthlySummary(context.Background(), 42, time.Time{})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(s.InvestmentsByType) != 2 {
		t.Fatalf("InvestmentsByType len = %d, want 2", len(s.InvestmentsByType))
	}
	if s.TotalInvestments != 300_000 {
		t.Errorf("TotalInvestments = %d, want 300000", s.TotalInvestments)
	}
	if s.AllTimeInvestments != 1_500_000 {
		t.Errorf("AllTimeInvestments = %d, want 1500000", s.AllTimeInvestments)
	}
}

// ─── AllTimeInvestments independe do período ─────────────────────────────────

func TestGetMonthlySummary_AllTimeInvestmentsIndependenteDoPeriodo(t *testing.T) {
	repo := &mockSummaryRepo{
		user:               pgstore.User{ID: 1},
		investmentRows:     nil,        // sem aportes no mês
		allTimeInvestments: 5_000_000,  // mas tem histórico acumulado
	}

	s, err := newUC(repo).GetMonthlySummary(context.Background(), 42, time.Time{})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if s.TotalInvestments != 0 {
		t.Errorf("TotalInvestments do mês = %d, want 0", s.TotalInvestments)
	}
	if s.AllTimeInvestments != 5_000_000 {
		t.Errorf("AllTimeInvestments = %d, want 5000000", s.AllTimeInvestments)
	}
}
