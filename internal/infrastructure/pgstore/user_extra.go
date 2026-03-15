package pgstore

import "context"

// UpdateUserPassword atualiza a senha do usuário. Não gerado pelo SQLC.
func (q *Queries) UpdateUserPassword(ctx context.Context, userID int64, hashedPassword string) error {
	_, err := q.db.Exec(ctx,
		`UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2`,
		hashedPassword, userID,
	)
	return err
}
