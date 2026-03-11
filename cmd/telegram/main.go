package main

import (
	"context"
	"log/slog"

	"github.com/DevKayoS/fintech-kodify/internal/adapters/telegram"
	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore/database"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func init() {
	ctx := context.Background()

	if err := database.Init(ctx); err != nil {
		slog.Error("failed to connect to database", "error", err)
	}

	slog.Info("database connected")
}

func main() {
	lambda.Start(Handler)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return telegram.HandleUpdate(ctx, req)
}
