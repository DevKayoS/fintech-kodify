package user

// User representa a entidade de domínio de usuário.
type User struct {
	ID             int64
	Name           string
	Email          string
	Password       string
	TelegramChatID *int64
	RoleID         *int64
}

// Repository define o contrato de persistência para usuários.
// Implementado por infrastructure/pgstore.
type Repository interface {
	// métodos adicionados conforme os use cases forem implementados
}
