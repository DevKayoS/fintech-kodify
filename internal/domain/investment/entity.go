package investment

import "time"

// Investment representa a entidade de domínio de investimento.
type Investment struct {
	ID               int64
	UserID           int64
	InvestmentTypeID int32
	Amount           int64 // armazenado em centavos
	Description      string
	InvestedAt       time.Time
	CreatedAt        time.Time
}

// Repository define o contrato de persistência para investimentos.
// Implementado por infrastructure/pgstore.
type Repository interface {
	// métodos adicionados conforme os use cases forem implementados
}
