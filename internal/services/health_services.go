package services

import (
	"context"
	"os"

	"github.com/DevKayoS/fintech-kodify/internal/pgstore"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthRepository interface {
	GetDBStatus(ctx context.Context) (pgstore.GetDBStatusRow, error)
}

type HealthService struct {
	repository HealthRepository
}

func NewHealthService(pool *pgxpool.Pool) *HealthService {
	return &HealthService{
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

func (s *HealthService) GetStatus(ctx context.Context) AppStatus {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	status := AppStatus{
		Status: "ok",
		Env:    env,
		DB:     DBStatus{OK: false},
	}

	row, err := s.repository.GetDBStatus(ctx)
	if err != nil {
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
