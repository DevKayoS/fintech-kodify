package expense

import (
	"context"

	"github.com/DevKayoS/fintech-kodify/internal/infrastructure/pgstore"
)

func (uc *ExpenseUseCase) ListCategories(ctx context.Context) ([]pgstore.ExpenseCategory, error) {
	return uc.repository.ListExpenseCategories(ctx)
}
