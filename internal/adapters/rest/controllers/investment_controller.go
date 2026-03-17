package controllers

import (
	"net/http"
	"strconv"

	"github.com/DevKayoS/fintech-kodify/internal/errors"
	"github.com/DevKayoS/fintech-kodify/internal/models"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/investment"
	"github.com/DevKayoS/fintech-kodify/internal/utils"
	"github.com/gin-gonic/gin"
)

type InvestmentController struct {
	useCase *investment.InvestmentUseCase
}

func NewInvestmentController(uc *investment.InvestmentUseCase) *InvestmentController {
	return &InvestmentController{useCase: uc}
}

// Create registra um novo investimento (aporte).
// POST /api/v1/investments
func (c *InvestmentController) Create(ctx *gin.Context) {
	userID := ctx.GetInt64("user_id")

	var req models.CreateInvestmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(errors.BadRequest(err.Error()))
		return
	}

	investedAt, err := utils.ParseDateTime(req.InvestedAt)
	if err != nil {
		ctx.Error(errors.BadRequest(err.Error()))
		return
	}

	result, err := c.useCase.Create(ctx.Request.Context(), userID, req.Amount, req.TypeSlug, req.Description, investedAt)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, models.InvestmentResponse{
		ID:           result.ID,
		TypeName:     result.TypeName,
		Amount:       result.Amount,
		Description:  result.Description,
		MovementType: result.MovementType,
		InvestedAt:   result.InvestedAt.Format("2006-01-02"),
		CreatedAt:    result.InvestedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// Withdraw registra um resgate de investimento.
// POST /api/v1/investments/withdraw
func (c *InvestmentController) Withdraw(ctx *gin.Context) {
	userID := ctx.GetInt64("user_id")

	var req models.WithdrawInvestmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(errors.BadRequest(err.Error()))
		return
	}

	investedAt, err := utils.ParseDateTime(req.InvestedAt)
	if err != nil {
		ctx.Error(errors.BadRequest(err.Error()))
		return
	}

	result, err := c.useCase.Withdraw(ctx.Request.Context(), userID, req.Amount, req.TypeSlug, req.Description, investedAt)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, models.InvestmentResponse{
		ID:           result.ID,
		TypeName:     result.TypeName,
		Amount:       result.Amount,
		Description:  result.Description,
		MovementType: result.MovementType,
		InvestedAt:   result.InvestedAt.Format("2006-01-02"),
		CreatedAt:    result.InvestedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// List retorna os investimentos do usuário autenticado.
// GET /api/v1/investments?month=YYYY-MM&type=slug
func (c *InvestmentController) List(ctx *gin.Context) {
	userID := ctx.GetInt64("user_id")

	var q models.ListInvestmentsQuery
	if err := ctx.ShouldBindQuery(&q); err != nil {
		ctx.Error(errors.BadRequest(err.Error()))
		return
	}

	items, err := c.useCase.List(ctx.Request.Context(), userID, q.Month, q.Type)
	if err != nil {
		ctx.Error(err)
		return
	}

	resp := make([]models.InvestmentResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, models.InvestmentResponse{
			ID:           item.ID,
			TypeSlug:     item.TypeSlug,
			TypeName:     item.TypeName,
			Amount:       item.Amount,
			Description:  item.Description,
			MovementType: item.MovementType,
			InvestedAt:   item.InvestedAt.Format("2006-01-02"),
			CreatedAt:    item.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetByID retorna um investimento pelo ID.
// GET /api/v1/investments/:id
func (c *InvestmentController) GetByID(ctx *gin.Context) {
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

	ctx.JSON(http.StatusOK, models.InvestmentResponse{
		ID:           item.ID,
		TypeSlug:     item.TypeSlug,
		TypeName:     item.TypeName,
		Amount:       item.Amount,
		Description:  item.Description,
		MovementType: item.MovementType,
		InvestedAt:   item.InvestedAt.Format("2006-01-02"),
		CreatedAt:    item.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// Delete remove um investimento.
// DELETE /api/v1/investments/:id
func (c *InvestmentController) Delete(ctx *gin.Context) {
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

// ListTypes retorna os tipos de investimento disponíveis.
// GET /api/v1/investments/types
func (c *InvestmentController) ListTypes(ctx *gin.Context) {
	types, err := c.useCase.ListTypes(ctx.Request.Context())
	if err != nil {
		ctx.Error(errors.Internal("erro ao buscar tipos", err))
		return
	}

	type typeItem struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}

	resp := make([]typeItem, 0, len(types))
	for _, t := range types {
		resp = append(resp, typeItem{Slug: t.Slug, Name: t.Name})
	}

	ctx.JSON(http.StatusOK, resp)
}
