package main

import (
	"context"
	"log/slog"

	"github.com/DevKayoS/fintech-kodify/internal/adapters/rest"
	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore/database"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
)

var ginLambda *ginadapter.GinLambda

func init() {
	ctx := context.Background()

	if err := database.Init(ctx); err != nil {
		slog.Error("failed to connect to database", "error", err)
	}

	slog.Info("database connected")
}

func main() {
	r := rest.SetupAPI()

	ginLambda = ginadapter.New(r)
	lambda.Start(Handler)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}
