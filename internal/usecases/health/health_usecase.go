package health

import (
	"context"
	"log/slog"
	"os"
	"time"
)

func (uc *HealthUseCase) GetStatus(ctx context.Context) AppStatus {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "production"
	}

	status := AppStatus{
		Status:    "ok",
		Env:       env,
		CheckedAt: time.Now().UTC(),
		DB:        DBStatus{OK: false},
	}

	start := time.Now()
	row, err := uc.repository.GetDBStatus(ctx)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		slog.Error("[GetStatus] algo deu errado ao tentar consultar o banco de dados", "error", err)
		return status
	}

	stat := uc.pool.Stat()
	status.DB = DBStatus{
		OK:             true,
		Version:        row.Version,
		MaxConnections: row.MaxConnections,
		LastMigration:  row.LastMigration,
		LatencyMs:      latency,
		Pool: PoolStats{
			Open: stat.TotalConns(),
			Idle: stat.IdleConns(),
			Max:  int32(stat.MaxConns()),
		},
	}

	return status
}
