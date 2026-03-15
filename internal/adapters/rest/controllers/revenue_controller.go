package controllers

import (
	"net/http"

	"github.com/DevKayoS/fintech-kodify/internal/errors"
	"github.com/DevKayoS/fintech-kodify/internal/models"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/revenue"
	"github.com/DevKayoS/fintech-kodify/internal/utils"
	"github.com/gin-gonic/gin"
)

type RevenueController struct {
	useCase *revenue.RevenueUseCase
}

func NewRevenueController(uc *revenue.RevenueUseCase) *RevenueController {
	return &RevenueController{useCase: uc}
}

// Create registra uma nova receita.
// POST /api/v1/revenues
func (c *RevenueController) Create(ctx *gin.Context) {
	userID := ctx.GetInt64("user_id")

	var req models.CreateRevenueRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(errors.BadRequest(err.Error()))
		return
	}

	receivedAt, err := utils.ParseDateTime(req.ReceivedAt)
	if err != nil {
		ctx.Error(errors.BadRequest(err.Error()))
		return
	}

	result, err := c.useCase.Create(ctx.Request.Context(), userID, req.Amount, req.Description, receivedAt)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, models.RevenueResponse{
		ID:          result.ID,
		Amount:      result.Amount,
		Description: result.Description,
		ReceivedAt:  result.ReceivedAt.Format("2006-01-02"),
		CreatedAt:   result.ReceivedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// List retorna as receitas do usuário autenticado.
// GET /api/v1/revenues?month=YYYY-MM
func (c *RevenueController) List(ctx *gin.Context) {
	userID := ctx.GetInt64("user_id")

	var q models.ListRevenuesQuery
	if err := ctx.ShouldBindQuery(&q); err != nil {
		ctx.Error(errors.BadRequest(err.Error()))
		return
	}

	items, err := c.useCase.List(ctx.Request.Context(), userID, q.Month)
	if err != nil {
		ctx.Error(err)
		return
	}

	resp := make([]models.RevenueResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, models.RevenueResponse{
			ID:          item.ID,
			Amount:      item.Amount,
			Description: item.Description,
			ReceivedAt:  item.ReceivedAt.Format("2006-01-02"),
			CreatedAt:   item.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	ctx.JSON(http.StatusOK, resp)
}
