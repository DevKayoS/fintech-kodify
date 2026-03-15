package health

import (
	"context"
	"time"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthRepository interface {
	GetDBStatus(ctx context.Context) (pgstore.GetDBStatusRow, error)
}

type HealthUseCase struct {
	repository HealthRepository
	pool       *pgxpool.Pool
}

func NewHealthUseCase(pool *pgxpool.Pool) *HealthUseCase {
	return &HealthUseCase{
		repository: pgstore.New(pool),
		pool:       pool,
	}
}

type PoolStats struct {
	Open int32 `json:"open"`
	Idle int32 `json:"idle"`
	Max  int32 `json:"max"`
}

type DBStatus struct {
	OK             bool      `json:"ok"`
	Version        string    `json:"version,omitempty"`
	MaxConnections int32     `json:"max_connections,omitempty"`
	LastMigration  int32     `json:"last_migration,omitempty"`
	LatencyMs      int64     `json:"latency_ms,omitempty"`
	Pool           PoolStats `json:"pool"`
}

type AppStatus struct {
	Status    string    `json:"status"`
	Env       string    `json:"env"`
	CheckedAt time.Time `json:"checked_at"`
	DB        DBStatus  `json:"db"`
}
