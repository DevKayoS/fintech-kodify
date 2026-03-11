package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Init(ctx context.Context) error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL não configurada")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("erro ao parsear DATABASE_URL: %w", err)
	}

	// Lambda: poucos workers concorrentes, não precisa de pool grande.
	// Para Neon em produção, manter baixo para não estourar connection limit.
	config.MaxConns = 2
	config.MinConns = 1
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConnIdleTime = 1 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	Pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("erro ao criar pool: %w", err)
	}

	if err = Pool.Ping(ctx); err != nil {
		return fmt.Errorf("erro ao pingar banco de dados: %w", err)
	}

	return nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}
