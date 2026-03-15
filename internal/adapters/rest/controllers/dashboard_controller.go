package controllers

import (
	"net/http"
	"time"

	"github.com/DevKayoS/fintech-kodify/internal/errors"
	"github.com/DevKayoS/fintech-kodify/internal/models"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/statement"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/summary"
	"github.com/DevKayoS/fintech-kodify/internal/utils"
	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	summaryUC   *summary.SummaryUseCase
	statementUC *statement.StatementUseCase
}

func NewDashboardController(summaryUC *summary.SummaryUseCase, statementUC *statement.StatementUseCase) *DashboardController {
	return &DashboardController{summaryUC: summaryUC, statementUC: statementUC}
}

// Summary retorna o resumo financeiro mensal do usuário autenticado.
// GET /api/v1/dashboard/summary?month=YYYY-MM
func (c *DashboardController) Summary(ctx *gin.Context) {
	userID := ctx.GetInt64("user_id")

	monthStart, err := parseMonthParam(ctx.Query("month"))
	if err != nil {
		ctx.Error(errors.BadRequest(err.Error()))
		return
	}

	s, err := c.summaryUC.GetForUser(ctx.Request.Context(), userID, monthStart)
	if err != nil {
		ctx.Error(err)
		return
	}

	expensesByCategory := make([]models.CategorySummaryResponse, 0, len(s.ExpensesByCategory))
	for _, cat := range s.ExpensesByCategory {
		expensesByCategory = append(expensesByCategory, models.CategorySummaryResponse{
			Slug:  cat.Slug,
			Name:  cat.Name,
			Total: utils.ToReais(cat.Total),
		})
	}

	investmentsByType := make([]models.InvestmentTypeSummaryResponse, 0, len(s.InvestmentsByType))
	for _, inv := range s.InvestmentsByType {
		investmentsByType = append(investmentsByType, models.InvestmentTypeSummaryResponse{
			Slug:  inv.Slug,
			Name:  inv.Name,
			Total: utils.ToReais(inv.Total),
		})
	}

	ctx.JSON(http.StatusOK, models.SummaryResponse{
		Month:              s.Month.Format("2006-01"),
		TotalExpenses:      utils.ToReais(s.TotalExpenses),
		TotalRevenues:      utils.ToReais(s.TotalRevenues),
		Balance:            utils.ToReais(s.Balance),
		TotalInvestments:   utils.ToReais(s.TotalInvestments),
		AllTimeInvestments: utils.ToReais(s.AllTimeInvestments),
		ExpensesByCategory: expensesByCategory,
		InvestmentsByType:  investmentsByType,
	})
}

// Statement retorna o extrato de transações do usuário autenticado.
// GET /api/v1/dashboard/statement?month=YYYY-MM
func (c *DashboardController) Statement(ctx *gin.Context) {
	userID := ctx.GetInt64("user_id")

	monthStart, err := parseMonthParam(ctx.Query("month"))
	if err != nil {
		ctx.Error(errors.BadRequest(err.Error()))
		return
	}

	result, err := c.statementUC.GetForUser(ctx.Request.Context(), userID, monthStart)
	if err != nil {
		ctx.Error(err)
		return
	}

	entries := make([]models.StatementEntryResponse, 0, len(result.Entries))
	for _, e := range result.Entries {
		entries = append(entries, models.StatementEntryResponse{
			Kind:        e.Kind,
			Date:        e.Date.Format("2006-01-02"),
			Amount:      e.Amount,
			Label:       e.Label,
			Description: e.Description,
		})
	}

	ctx.JSON(http.StatusOK, models.StatementResponse{
		Month:   result.Month.Format("2006-01"),
		Entries: entries,
	})
}

func parseMonthParam(month string) (time.Time, error) {
	if month == "" {
		return time.Time{}, nil
	}
	start, _, err := utils.ParseMonthRange(month)
	return start, err
}
