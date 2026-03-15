package main

import (
	"context"
	"log/slog"

	"github.com/DevKayoS/fintech-kodify/internal/adapters/telegram"
	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore/database"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/expense"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/investment"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/revenue"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/summary"
	"github.com/DevKayoS/fintech-kodify/internal/usecases/user"
	"github.com/aws/aws-lambda-go/events"
	awslambda "github.com/aws/aws-lambda-go/lambda"
)

var telegramHandler *telegram.Handler

func init() {
	ctx := context.Background()

	if err := database.Init(ctx); err != nil {
		slog.Error("failed to connect to database", "error", err)
	}

	expenseUC := expense.NewExpenseUseCase(database.Pool)
	investmentUC := investment.NewInvestmentUseCase(database.Pool)
	revenueUC := revenue.NewRevenueUseCase(database.Pool)
	summaryUC := summary.NewSummaryUseCase(database.Pool)
	userUC := user.NewUserUseCase(database.Pool)
	telegramHandler = telegram.NewHandler(expenseUC, investmentUC, revenueUC, summaryUC, userUC)

	slog.Info("database connected")
}

func main() {
	awslambda.Start(Handler)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return telegramHandler.HandleUpdate(ctx, req)
}
