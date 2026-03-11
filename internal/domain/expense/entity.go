package expense

import "time"

// Expense representa a entidade de domínio de despesa.
type Expense struct {
	ID          int64
	UserID      int64
	CategoryID  int32
	Amount      int64 // armazenado em centavos
	Description string
	OccurredAt  time.Time
	CreatedAt   time.Time
}

// Repository define o contrato de persistência para despesas.
// Implementado por infrastructure/pgstore.
type Repository interface {
	// métodos adicionados conforme os use cases forem implementados
}
