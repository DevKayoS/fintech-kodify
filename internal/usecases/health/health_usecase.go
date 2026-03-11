package health

import (
	"context"
	"log/slog"
	"os"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthRepository interface {
	GetDBStatus(ctx context.Context) (pgstore.GetDBStatusRow, error)
}

type HealthUseCase struct {
	repository HealthRepository
}

func NewHealthUseCase(pool *pgxpool.Pool) *HealthUseCase {
	return &HealthUseCase{
		repository: pgstore.New(pool),
	}
}

type DBStatus struct {
	OK             bool   `json:"ok"`
	Version        string `json:"version,omitempty"`
	MaxConnections int32  `json:"max_connections,omitempty"`
	LastMigration  int64  `json:"last_migration,omitempty"`
}

type AppStatus struct {
	Status string   `json:"status"`
	Env    string   `json:"env"`
	DB     DBStatus `json:"db"`
}

func (uc *HealthUseCase) GetStatus(ctx context.Context) AppStatus {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "production"
	}

	status := AppStatus{
		Status: "ok",
		Env:    env,
		DB:     DBStatus{OK: false},
	}

	row, err := uc.repository.GetDBStatus(ctx)
	if err != nil {
		slog.Error("[GetStatus] algo deu errado ao tentar consultar o banco de dados", "error", err)
		return status
	}

	status.DB = DBStatus{
		OK:             true,
		Version:        row.Version,
		MaxConnections: row.MaxConnections,
		LastMigration:  row.LastMigration,
	}

	return status
}
