package health

import (
	"context"
	"log/slog"
	"os"
)

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
