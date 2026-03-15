package controllers

import (
	"net/http"
	"strconv"

	"github.com/DevKayoS/fintech-kodify/internal/errors"
	"github.com/DevKayoS/fintech-kodify/internal/models"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/expense"
	"github.com/DevKayoS/fintech-kodify/internal/utils"
	"github.com/gin-gonic/gin"
)

type ExpenseController struct {
	useCase *expense.ExpenseUseCase
}

func NewExpenseController(uc *expense.ExpenseUseCase) *ExpenseController {
	return &ExpenseController{useCase: uc}
}

// Create registra um novo gasto.
// POST /api/v1/expenses
func (c *ExpenseController) Create(ctx *gin.Context) {
	userID := ctx.GetInt64("user_id")

	var req models.CreateExpenseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(errors.BadRequest(err.Error()))
		return
	}

	occurredAt, err := utils.ParseDateTime(req.OccurredAt)
	if err != nil {
		ctx.Error(errors.BadRequest(err.Error()))
		return
	}

	result, err := c.useCase.Create(ctx.Request.Context(), userID, req.Amount, req.CategorySlug, req.Description, occurredAt)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, models.ExpenseResponse{
		ID:           result.ID,
		CategoryName: result.CategoryName,
		Amount:       result.Amount,
		Description:  result.Description,
		OccurredAt:   result.OccurredAt.Format("2006-01-02"),
		CreatedAt:    result.OccurredAt.Format("2006-01-02T15:04:05Z"),
	})
}

// List retorna os gastos do usuário autenticado.
// GET /api/v1/expenses?month=YYYY-MM&category=slug
func (c *ExpenseController) List(ctx *gin.Context) {
	userID := ctx.GetInt64("user_id")

	var q models.ListExpensesQuery
	if err := ctx.ShouldBindQuery(&q); err != nil {
		ctx.Error(errors.BadRequest(err.Error()))
		return
	}

	items, err := c.useCase.List(ctx.Request.Context(), userID, q.Month, q.Category)
	if err != nil {
		ctx.Error(err)
		return
	}

	resp := make([]models.ExpenseResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, models.ExpenseResponse{
			ID:           item.ID,
			CategorySlug: item.CategorySlug,
			CategoryName: item.CategoryName,
			Amount:       item.Amount,
			Description:  item.Description,
			OccurredAt:   item.OccurredAt.Format("2006-01-02"),
			CreatedAt:    item.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetByID retorna um gasto pelo ID.
// GET /api/v1/expenses/:id
func (c *ExpenseController) GetByID(ctx *gin.Context) {
	userID := ctx.GetInt64("user_id")

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(errors.BadRequest("id inválido"))
		return
	}

	item, err := c.useCase.GetByID(ctx.Request.Context(), userID, id)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, models.ExpenseResponse{
		ID:           item.ID,
		CategorySlug: item.CategorySlug,
		CategoryName: item.CategoryName,
		Amount:       item.Amount,
		Description:  item.Description,
		OccurredAt:   item.OccurredAt.Format("2006-01-02"),
		CreatedAt:    item.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// Delete remove um gasto.
// DELETE /api/v1/expenses/:id
func (c *ExpenseController) Delete(ctx *gin.Context) {
	userID := ctx.GetInt64("user_id")

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(errors.BadRequest("id inválido"))
		return
	}

	if err := c.useCase.Delete(ctx.Request.Context(), userID, id); err != nil {
		ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// ListCategories retorna as categorias de despesa disponíveis.
// GET /api/v1/expenses/categories
func (c *ExpenseController) ListCategories(ctx *gin.Context) {
	categories, err := c.useCase.ListCategories(ctx.Request.Context())
	if err != nil {
		ctx.Error(errors.Internal("erro ao buscar categorias", err))
		return
	}

	type categoryItem struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}

	resp := make([]categoryItem, 0, len(categories))
	for _, cat := range categories {
		resp = append(resp, categoryItem{Slug: cat.Slug, Name: cat.Name})
	}

	ctx.JSON(http.StatusOK, resp)
}
